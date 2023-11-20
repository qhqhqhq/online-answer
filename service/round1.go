package service

import (
	"fmt"
	"log"
	"net/http"
	"online-answer/service/round"
	"online-answer/utils"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWSRound1Player(w http.ResponseWriter, r *http.Request) {
	openId, groupNumber, err := utils.Authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	round.PlayerConnections[openId] = conn
	defer func() {
		delete(round.PlayerConnections, openId)
	}()

	for {
		_, buf, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		msg := string(buf)
		if msg == "true" {
			round.Round1Instance.AnswerQuestion(groupNumber, openId, true)
		} else if msg == "false" {
			round.Round1Instance.AnswerQuestion(groupNumber, openId, false)
		}
	}
}

func HandleWSRound1Display(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	round.CenterConnection = conn
	defer func() {
		round.CenterConnection = nil
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func HandleRound1DisplayIndex(w http.ResponseWriter, r *http.Request) {
	b, err := os.ReadFile("./Round1DisplayIndex.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(b))
}
