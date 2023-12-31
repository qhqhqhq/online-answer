package main

import (
	"fmt"
	"log"
	"net/http"
	"online-answer/db"
	"online-answer/service"
)

func main() {
	if err := db.Init(); err != nil {
		panic(fmt.Sprintf("mysql init failed with %+v", err))
	}

	// 设置静态文件目录
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// auth
	http.HandleFunc("/login", service.HandleLogin)
	http.HandleFunc("/logout", service.HandleLogout)

	// round2
	http.HandleFunc("/round2/state", service.HandleRound2State)
	http.HandleFunc("/round2/getquestion", service.HandleRound2GetQuestion)
	http.HandleFunc("/round2/submit", service.HandleRound2Submit)
	http.HandleFunc("/round2/display", service.HandleRound2DisplayIndex)
	http.HandleFunc("/round2/ws/display", service.HandleWSRound2Display)

	// round1
	http.HandleFunc("/round1/ws/player", service.HandleWSRound1Player)
	http.HandleFunc("/round1/ws/display", service.HandleWSRound1Display)
	http.HandleFunc("/round1/display", service.HandleRound1DisplayIndex)

	// admin
	http.HandleFunc("/admin/start_round1", service.HandleStartRound1)
	http.HandleFunc("/admin/cancel_round1", service.HandleCancelRound1)
	http.HandleFunc("/admin/start_round2", service.HandleStartRound2)
	http.HandleFunc("/admin/cancel_round2", service.HandleCancelRound2)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
