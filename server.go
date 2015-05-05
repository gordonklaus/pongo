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
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		select {
		case ch1 <- ws:
			ws2 := <-ch2
			if err := ws2.WriteJSON(false); err != nil {
				log.Println(err)
				return
			}
			relay(ws2, ws)
		case ch2 <- ws:
			ws1 := <-ch1
			if err := ws1.WriteJSON(true); err != nil {
				log.Println(err)
				return
			}
			relay(ws1, ws)
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
			log.Println("NextReader: ", err)
			break
		}
		w, err := dst.NextWriter(messageType)
		if err != nil {
			log.Println("NextWriter: ", err)
			break
		}
		if _, err := io.Copy(w, r); err != nil {
			log.Println("Copy: ", err)
			break
		}
		if err := w.Close(); err != nil {
			log.Println("Close: ", err)
			break
		}
	}
}
