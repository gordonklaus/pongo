package main

import (
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("")))
	handleWebSocket()
	log.Fatal("ListenAndServe:", http.ListenAndServe(":8080", nil))
}

func handleWebSocket() {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade:", err)
			return
		}
		webSockets <- ws
	})

	go pairWebSockets()
}

var webSockets = make(chan *websocket.Conn)

func pairWebSockets() {
	var c1 *websocket.Conn
	for c2 := range webSockets {
		if err := c2.WriteJSON(c1 == nil); err != nil {
			log.Println("WriteJSON:", err)
			continue
		}
		if c1 == nil {
			c1 = c2
			continue
		}
		go copy(c1, c2)
		go copy(c2, c1)
		c1 = nil
	}
}

func copy(dst, src *websocket.Conn) {
	defer dst.Close()
	for {
		messageType, r, err := src.NextReader()
		if err != nil {
			break
		}
		w, err := dst.NextWriter(messageType)
		if err != nil {
			break
		}
		if _, err := io.Copy(w, r); err != nil {
			log.Println("Copy:", err)
			break
		}
		if err := w.Close(); err != nil {
			log.Println("Close:", err)
			break
		}
	}
}
