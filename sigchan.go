//+build js

package main

import (
	"encoding/json"
	// "fmt"
	"io"

	"github.com/gopherjs/websocket"
	"github.com/gordonklaus/webrtc"
)

type signallingChannel struct {
	ws  *websocket.Conn
	enc *json.Encoder
	dec *json.Decoder

	initiator           bool
	sessionDescriptions chan webrtc.SessionDescription
	iceCandidates       chan webrtc.ICECandidate
	closed              chan struct{}
}

func openSignallingChannel() *signallingChannel {
	ws, err := websocket.Dial("ws://localhost:12345/ws")
	chk(err)
	s := &signallingChannel{
		ws,
		json.NewEncoder(ws),
		json.NewDecoder(ws),
		false,
		make(chan webrtc.SessionDescription),
		make(chan webrtc.ICECandidate),
		make(chan struct{}),
	}

	err = s.dec.Decode(&s.initiator)
	chk(err)

	go s.recv()
	return s
}

func (s *signallingChannel) close() {
	s.ws.Close()
}

func (s *signallingChannel) recv() {
	for {
		var msg wsMessage
		err := s.dec.Decode(&msg)
		if err == io.EOF {
			close(s.closed)
			break
		}
		chk(err)
		if msg.SessionDescription != nil {
			s.sessionDescriptions <- *msg.SessionDescription
		} else if msg.ICECandidate != nil {
			s.iceCandidates <- *msg.ICECandidate
		} else {
			close(s.iceCandidates)
		}
	}
}

type wsMessage struct {
	SessionDescription *webrtc.SessionDescription
	ICECandidate       *webrtc.ICECandidate
}
