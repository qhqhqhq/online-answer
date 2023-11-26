package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"online-answer/db"
	"online-answer/db/model"
	"online-answer/service/round"
	"online-answer/utils"
	"os"
)

func HandleRound2State(w http.ResponseWriter, r *http.Request) {
	round2StateResp := Round2StateResponse{
		Start:          round.Round2Instance.Start,
		RemainingTime:  round.Round2Instance.RemainingTime,
		PromotionCount: len(round.Round2Instance.PromotionGroups),
		Target:         round.Round2Instance.TargetScore,
	}

	msg, _ := json.Marshal(&round2StateResp)
	w.Header().Set("content-type", "application/json")
	w.Write(msg)
}

func HandleRound2GetQuestion(w http.ResponseWriter, r *http.Request) {
	_, _, err := utils.Authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	cli := db.Get()
	var question model.ChoiceQuestion
	err = cli.Preload("Options").Order("RAND()").First(&question).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	options := make(map[string]string)
	for _, v := range question.Options {
		options[v.Key] = v.Value
	}

	round2GetQuestionResp := Round2GetQuestionResponse{
		Number:     question.Number,
		IsMultiple: question.IsMultipleChoice,
		Content:    question.Content,
		Options:    options,
	}

	msg, err := json.Marshal(&round2GetQuestionResp)
	w.Header().Set("content-type", "application/json")
	w.Write(msg)
}

func HandleRound2Submit(w http.ResponseWriter, r *http.Request) {
	openId, groupNumber, err := utils.Authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var round2SubmitReq Round2SubmitRequest
	var round2SubmitResp Round2SubmitResponse
	cli := db.Get()

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err = decoder.Decode(&round2SubmitReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var question model.ChoiceQuestion
	err = cli.Where(&model.ChoiceQuestion{Number: round2SubmitReq.Number}).First(&question).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var score int
	if round2SubmitReq.Answer == question.Answer {
		score, err = round.Round2Instance.IncrementScore(groupNumber, openId)
		round2SubmitResp.Correct = true
	} else {
		score, err = round.Round2Instance.DecrementScore(groupNumber, openId)
		round2SubmitResp.Correct = false
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	round2SubmitResp.Score = score

	msg, _ := json.Marshal(&round2SubmitResp)
	w.Header().Set("content-type", "application/json")
	w.Write(msg)

}

func HandleRound2DisplayIndex(w http.ResponseWriter, r *http.Request) {
	b, err := os.ReadFile("./Round2DisplayIndex.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(b))
}

func HandleWSRound2Display(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	round.Round2CenterConnection = conn
	defer func() {
		round.Round2CenterConnection = nil
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
	}
}
