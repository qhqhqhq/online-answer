package round

import (
	"context"
	"errors"
	"fmt"
	"log"
	"online-answer/db"
	"online-answer/db/model"
	"online-answer/utils"
	"sort"
	"sync"
	"time"
)

var Round2Instance = Round2{Start: false}

type Round2 struct {
	Start                bool
	TargetScore          int
	TargetPromotionCount int
	RemainingTime        int
	PromotionGroups      []uint
	PlayerStateMap       map[uint]*GroupState

	initialized bool
	cancel      context.CancelFunc
	mutex       sync.Mutex
}

type GroupState struct {
	promotion bool
	MemberMap map[string]*PlayerState
}

type PlayerState struct {
	score     int
	promotion bool
}

func (r *Round2) Init(targetScore int, targetPC int, remainingTime int, cancel context.CancelFunc) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.initialized {
		return errors.New("round initialized")
	}

	r.Start = false
	r.TargetScore = targetScore
	r.TargetPromotionCount = targetPC
	r.RemainingTime = remainingTime
	r.PromotionGroups = make([]uint, 0)

	r.PlayerStateMap = make(map[uint]*GroupState)
	cli := db.Get()
	var groups []*model.Group
	err := cli.Preload("Members").Where("eliminate = ? AND promotion = ?", false, false).Find(&groups).Error
	if err != nil {
		return err
	}
	for _, group := range groups {
		newMemberMap := make(map[string]*PlayerState)
		for _, member := range group.Members {
			playerState := &PlayerState{score: 0, promotion: false}
			newMemberMap[member.OpenID] = playerState
		}
		r.PlayerStateMap[group.Number] = &GroupState{promotion: false, MemberMap: newMemberMap}
	}

	r.cancel = cancel

	r.initialized = true

	return nil
}

func (r *Round2) Run(ctx context.Context) {
	defer r.Destroy()
	r.Start = true
	loop := true
	for loop && r.RemainingTime > 0 {
		select {
		case <-ctx.Done():
			loop = false
			break
		default:
			time.Sleep(1 * time.Second)
			r.RemainingTime--
		}
	}
	r.Start = false

	var candidates []uint
	residue := r.TargetPromotionCount - len(r.PromotionGroups)
	if residue > 0 && r.RemainingTime == 0 {
		type groupScore struct {
			number uint
			score  int
		}
		var groupScores []groupScore
		for k, v := range r.PlayerStateMap {
			if !v.promotion {
				sum := 0
				for _, state := range v.MemberMap {
					sum += state.score
				}
				groupScores = append(groupScores, groupScore{number: k, score: sum})
			}
		}
		sort.Slice(groupScores, func(i, j int) bool {
			return groupScores[i].score > groupScores[j].score
		})

		beg := 0
		end := 0
		count := 0
		for end <= len(groupScores) && count < residue && end-beg <= residue-count {
			for i := beg; i < end; i++ {
				group := groupScores[i]
				err := r.promoteGroup(group.number)
				if err != nil {
					log.Printf("group %d promote failed, err: %v\n", group.number, err)
					continue
				}
				count++
			}
			beg = end
			end++
			for end < len(groupScores) && groupScores[end-1].score == groupScores[end].score {
				end++
			}
		}

		if beg < len(groupScores) {
			for _, v := range groupScores[beg:end] {
				candidates = append(candidates, v.number)
			}
		}

		log.Println("group scores: ", groupScores)
	}

	db.Get().Save(&model.Record{
		Result:     utils.JoinUintSlice(r.PromotionGroups, ","),
		Candidates: utils.JoinUintSlice(candidates, ","),
	})
	log.Println("promotion map: ")
	for number, groupState := range r.PlayerStateMap {
		log.Printf("group %d: promoted %v\n", number, groupState.promotion)
		log.Printf("members score: ")
		for openId, memberState := range groupState.MemberMap {
			log.Printf("%s: %d, ", openId, memberState.score)
		}
		log.Printf("\n")
	}
	log.Println("promotion groups: ", r.PromotionGroups)

	log.Println("round ended")

}

func (r *Round2) Terminate() {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *Round2) Destroy() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.Start = false
	r.RemainingTime = 0
	r.PromotionGroups = nil
	r.PlayerStateMap = nil

	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}

	r.initialized = false

	log.Println("round destroy")
}

func (r *Round2) IncrementScore(groupNumber uint, OpenId string) (int, error) {
	if !r.Start {
		return 0, errors.New("round not started")
	}

	groupState, ok := r.PlayerStateMap[groupNumber]
	if !ok {
		return 0, errors.New(fmt.Sprintf("group %d not found", groupNumber))
	}

	playerState, ok := groupState.MemberMap[OpenId]
	if !ok {
		return 0, errors.New(fmt.Sprintf("member %s not found", OpenId))
	}

	if playerState.score < r.TargetScore {
		playerState.score++
	}

	if playerState.score >= r.TargetScore {
		playerState.promotion = true
	}

	if playerState.promotion {
		ifGroupPromote := true
		for _, v := range groupState.MemberMap {
			if !v.promotion {
				ifGroupPromote = false
				break
			}
		}
		if ifGroupPromote {
			err := r.promoteGroup(groupNumber)
			if err != nil {
				log.Printf("group %d promote failed, err: %v\n", groupNumber, err)
			}
		}
	}

	return playerState.score, nil

}

func (r *Round2) DecrementScore(groupNumber uint, OpenId string) (int, error) {
	if !r.Start {
		return 0, errors.New("round not started")
	}

	groupState, ok := r.PlayerStateMap[groupNumber]
	if !ok {
		return 0, errors.New(fmt.Sprintf("group %d not found", groupNumber))
	}

	playerState, ok := groupState.MemberMap[OpenId]
	if !ok {
		return 0, errors.New(fmt.Sprintf("member %s not found", OpenId))
	}

	if playerState.promotion {
		return playerState.score, nil
	}
	playerState.score--

	return playerState.score, nil

}

func (r *Round2) promoteGroup(groupNumber uint) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.PromotionGroups) >= r.TargetPromotionCount {
		r.Terminate()
		return errors.New("no promotion count")
	}

	groupState, ok := r.PlayerStateMap[groupNumber]
	if !ok {
		return errors.New(fmt.Sprintf("group %d not found", groupNumber))
	}
	if groupState.promotion {
		return errors.New(fmt.Sprintf("group %d already promoted", groupNumber))
	}

	cli := db.Get()
	var group model.Group
	err := cli.Where(&model.Group{Number: groupNumber}).First(&group).Error
	if err != nil {
		return err
	}

	var collegePromotedCount int64
	err = cli.Model(&model.Group{}).Where(&model.Group{College: group.College, Promotion: true}).Count(&collegePromotedCount).Error
	if err != nil {
		return err
	}

	if collegePromotedCount >= 2 {
		return errors.New("college of group has more than 2 groups promoted")
	}
	group.Promotion = true
	err = cli.Save(&group).Error
	if err != nil {
		return err
	}
	r.PromotionGroups = append(r.PromotionGroups, groupNumber)
	groupState.promotion = true
	log.Println("success promote the group: ", groupNumber)

	if len(r.PromotionGroups) >= r.TargetPromotionCount {
		r.Terminate()
	}

	return nil
}
