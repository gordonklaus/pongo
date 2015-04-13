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
	log.Fatal("ListenAndServe: ", http.ListenAndServe(":12345", nil))
}

func handleWebSocket() {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ch1 := make(chan *websocket.Conn, 1)
	ch2 := make(chan *websocket.Conn, 1)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		select {
		case ch1 <- c:
			if err := c.WriteMessage(websocket.TextMessage, []byte{'0'}); err != nil {
				log.Fatal("WriteMessage: ", err)
			}
			relay(c, <-ch2)
		case ch2 <- c:
			if err := c.WriteMessage(websocket.TextMessage, []byte{'1'}); err != nil {
				log.Fatal("WriteMessage: ", err)
			}
			relay(c, <-ch1)
		}
	})
}

func relay(dst, src *websocket.Conn) {
	for {
		messageType, r, err := src.NextReader()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("NextReader: ", err)
		}
		w, err := dst.NextWriter(messageType)
		if err != nil {
			log.Fatal("NextWriter: ", err)
		}
		if _, err := io.Copy(w, r); err != nil {
			log.Println("Copy: ", err)
		}
		if err := w.Close(); err != nil {
			log.Fatal("Close: ", err)
		}
	}
}
