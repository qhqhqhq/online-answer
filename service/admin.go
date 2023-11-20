package service

import (
	"context"
	"encoding/json"
	"net/http"
	"online-answer/service/round"
)

func HandleStartRound1(w http.ResponseWriter, r *http.Request) {
	// TODO: auth

	var startRound1Req StartRound1Request
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&startRound1Req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	err = round.Round1Instance.Init(
		startRound1Req.Candidates,
		startRound1Req.TargetEliminatedCount,
		startRound1Req.AnswerTime,
		startRound1Req.InitialScore,
		cancel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	go round.Round1Instance.Run(ctx)
}

func HandleCancelRound1(w http.ResponseWriter, r *http.Request) {
	// TODO: auth

	round.Round1Instance.Terminate()
}

func HandleStartRound2(w http.ResponseWriter, r *http.Request) {
	// TODO: auth

	var startRound2Req StartRound2Request
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&startRound2Req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	err = round.Round2Instance.Init(startRound2Req.TargetScore, startRound2Req.TargetPromotionCount, startRound2Req.RemainingTime, cancel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	go round.Round2Instance.Run(ctx)
}

func HandleCancelRound2(w http.ResponseWriter, r *http.Request) {
	// TODO: auth

	round.Round2Instance.Terminate()
}
