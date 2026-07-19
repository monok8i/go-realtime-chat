package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

type message struct {
	UserID  int32  `json:"user_id"`
	ChatID  string `json:"chat_id"`
	Message string `json:"message"`
}

func main() {
	addr := flag.String("addr", "localhost:8000", "server address")
	flag.Parse()

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/api/ws/chat"}
	log.Printf("connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("read: %v", err)
				return
			}
			log.Printf("received: %s", msg)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			var msg message
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				log.Printf("invalid json: %v", err)
				continue
			}
			if err := conn.WriteJSON(msg); err != nil {
				log.Printf("write: %v", err)
				return
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("stdin error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("shutting down")
}
