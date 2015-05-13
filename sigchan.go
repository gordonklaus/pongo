//+build js

package main

import (
	"encoding/json"

	"github.com/gopherjs/websocket"
	"github.com/gordonklaus/webrtc"
)

type signallingChannel struct {
	ws  *websocket.Conn
	enc *json.Encoder
	dec *json.Decoder
}

func openSignallingChannel() *signallingChannel {
	ws, err := websocket.Dial("ws://localhost:12345/ws")
	chk(err)
	s := &signallingChannel{
		ws,
		json.NewEncoder(ws),
		json.NewDecoder(ws),
	}
	return s
}

func (s *signallingChannel) initiator() (bool, error) {
	var x bool
	err := s.dec.Decode(&x)
	return x, err
}

func (s *signallingChannel) Recv() (webrtc.Message, error) {
	var m webrtc.Message
	err := s.dec.Decode(&m)
	return m, err
}

func (s *signallingChannel) Send(m webrtc.Message) error {
	return s.enc.Encode(&m)
}

func (s *signallingChannel) close() {
	s.ws.Close()
}
