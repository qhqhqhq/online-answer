package round

import (
	"context"
	"errors"
	"fmt"
	"log"
	"online-answer/db"
	"online-answer/db/model"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var Round1Instance = Round1{}
var PlayerConnections = make(map[string]*websocket.Conn)
var CenterConnection *websocket.Conn

type Round1 struct {
	AnswerTime            int
	InitialScore          int
	TargetEliminatedCount int
	GroupCount            int
	EliminatedGroupCount  int
	PlayersMap            map[uint]*Round1GroupState

	Start                bool
	QuestionNumber       uint
	Content              string
	Answer               bool
	LastEliminatedGroups []uint

	initialized bool
	mutex       sync.Mutex
	cancel      context.CancelFunc
}

type Round1GroupState struct {
	Eliminated bool
	MembersMap map[string]*Round1MemberState
}

type Round1MemberState struct {
	Eliminated    bool
	AnswerCorrect bool
	Score         int

	mutex sync.Mutex
}

func (r *Round1) Init(candidates []uint, targetEC int, answerTime int, initialScore int, cancel context.CancelFunc) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.initialized {
		return errors.New("round initialized")
	}

	r.AnswerTime = answerTime
	r.InitialScore = initialScore

	cli := db.Get()
	var groups []*model.Group
	if len(candidates) != 0 {
		// 指定小组开始轮次
		err := cli.Model(&model.Group{}).Where("number IN ?", candidates).Update("eliminate", false).Error
		if err != nil {
			return err
		}

		err = cli.Preload("Members").Where("number IN ?", candidates).Find(&groups).Error
		if err != nil {
			return err
		}

	} else {
		err := cli.Preload("Members").Where(&model.Group{Eliminate: false, Promotion: false}).Find(&groups).Error
		if err != nil {
			return err
		}
	}

	r.TargetEliminatedCount = targetEC
	r.GroupCount = len(groups)
	r.EliminatedGroupCount = 0
	r.PlayersMap = make(map[uint]*Round1GroupState)
	for _, group := range groups {
		newMemberMap := make(map[string]*Round1MemberState)
		for _, member := range group.Members {
			memberState := &Round1MemberState{Eliminated: false, AnswerCorrect: false, Score: initialScore}
			newMemberMap[member.OpenID] = memberState
		}
		r.PlayersMap[group.Number] = &Round1GroupState{Eliminated: false, MembersMap: newMemberMap}
	}

	r.Start = false
	r.LastEliminatedGroups = make([]uint, 0)

	r.cancel = cancel

	r.initialized = true

	return nil

}

func (r *Round1) Run(ctx context.Context) {
	defer r.destroy()
	for r.EliminatedGroupCount < r.TargetEliminatedCount {
		select {
		case <-ctx.Done():
			break
		default:
			r.sendGroupsScore()

			if err := r.setNewRoundSlice(); err != nil {
				log.Println("error when set new round slice: ", err)
				break
			}
			r.sendMetadata()

			r.Start = true
			r.spin()
			r.Start = false

			if err := r.settle(); err != nil {
				log.Println("error when settle the scores: ", err)
				break
			}

			r.sendResult()

		}

	}

	log.Println("group count: ", r.GroupCount)
	log.Println("eliminated group count: ", r.EliminatedGroupCount)
	log.Println("last eliminated groups: ", r.LastEliminatedGroups)
	log.Println("players map: ", r.PlayersMap)

	log.Println("round ended")

}

func (r *Round1) Terminate() {
	if r.cancel != nil {
		r.cancel()
	}

}

func (r *Round1) AnswerQuestion(groupNumber uint, openId string, answer bool) (bool, error) {
	if !r.Start {
		return false, errors.New("round slice not started")
	}

	groupState, ok := r.PlayersMap[groupNumber]
	if !ok {
		return false, errors.New("group not found")
	}

	memberState, ok := groupState.MembersMap[openId]
	if !ok {
		return false, errors.New("member not found")
	}

	if memberState.Eliminated {
		return false, errors.New("already eliminated")
	}

	memberState.mutex.Lock()
	defer memberState.mutex.Unlock()

	if answer == r.Answer {
		memberState.AnswerCorrect = true
		return true, nil
	}

	return false, nil

}

func (r *Round1) eliminateGroup(groupNumber uint) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	groupState, _ := r.PlayersMap[groupNumber]
	if groupState.Eliminated {
		return nil
	}

	cli := db.Get()
	var group model.Group
	err := cli.Where(&model.Group{Number: groupNumber}).First(&group).Error
	if err != nil {
		return err
	}

	group.Eliminate = true
	err = cli.Save(&group).Error
	if err != nil {
		return err
	}
	r.LastEliminatedGroups = append(r.LastEliminatedGroups, groupNumber)
	groupState.Eliminated = true
	r.EliminatedGroupCount++

	return nil
}

func (r *Round1) setNewRoundSlice() error {
	var question model.JudgementQuestion
	cli := db.Get()
	err := cli.Order("RAND()").First(&question).Error
	if err != nil {
		return err
	}

	r.QuestionNumber = question.Number
	r.Content = question.Content
	r.Answer = question.Answer
	r.LastEliminatedGroups = make([]uint, 0)

	return nil

}

func (r *Round1) sendMetadata() {
	// 向中心节点发送元数据
	if CenterConnection != nil {

		CenterConnection.WriteJSON(&Round1SliceMetadata{
			Type:                  "metadata",
			TargetEliminatedCount: r.TargetEliminatedCount,
			GroupCount:            r.GroupCount,
			EliminatedGroupCount:  r.EliminatedGroupCount,
			QuestionNumber:        r.QuestionNumber,
			Content:               r.Content,
		})
	}
}

func (r *Round1) sendRemainingTime(remainingTime int) {
	if CenterConnection != nil {

		CenterConnection.WriteJSON(&Round1RemainingTime{
			Type:          "time",
			RemainingTime: remainingTime,
		})
	}
}

func (r *Round1) sendResult() {
	if CenterConnection != nil {

		CenterConnection.WriteJSON(&Round1SliceResult{
			Type:                 "result",
			Answer:               r.Answer,
			LastEliminatedGroups: r.LastEliminatedGroups,
		})
	}
}

func (r *Round1) sendGroupsScore() {
	for _, groupState := range r.PlayersMap {
		for openId, memberState := range groupState.MembersMap {
			if conn, ok := PlayerConnections[openId]; ok {
				scoreMsg := fmt.Sprintf("score,%d", memberState.Score)
				conn.WriteMessage(websocket.TextMessage, []byte(scoreMsg))
				groupEliminatedMsg := fmt.Sprintf("group_eliminated,%v", groupState.Eliminated)
				conn.WriteMessage(websocket.TextMessage, []byte(groupEliminatedMsg))
			}
		}
	}

}

func (r *Round1) spin() {
	remaining := r.AnswerTime
	for remaining > 0 {
		time.Sleep(1 * time.Second)
		remaining--
		r.sendRemainingTime(remaining)
	}
}

func (r *Round1) settle() error {
	for number, groupState := range r.PlayersMap {
		if groupState.Eliminated {
			continue
		}
		e := true
		for _, memberState := range groupState.MembersMap {
			if memberState.Eliminated {
				continue
			}
			memberState.mutex.Lock()

			if memberState.AnswerCorrect {
				memberState.Score++
			} else {
				memberState.Score--
			}
			memberState.AnswerCorrect = false

			if memberState.Score <= 0 {
				memberState.Eliminated = true
			}

			if !memberState.Eliminated {
				e = false
			}

			memberState.mutex.Unlock()
		}
		if e {
			err := r.eliminateGroup(number)
			if err != nil {
				return err
			}
		}
	}
	return nil

}

func (r *Round1) destroy() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.AnswerTime = 0
	r.InitialScore = 0
	r.GroupCount = 0
	r.EliminatedGroupCount = 0
	r.PlayersMap = nil

	r.Start = false
	r.QuestionNumber = 0
	r.Content = ""
	r.Answer = false

	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}

	r.initialized = false

}
