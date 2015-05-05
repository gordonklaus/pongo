//+build js

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/gopherjs/websocket"
	"github.com/gordonklaus/webrtc"
	"honnef.co/go/js/dom"
)

func main() {
	ws, err := websocket.Dial("ws://localhost:12345/ws")
	chk(err)
	// defer ws.Close()

	enc := json.NewEncoder(ws)
	dec := json.NewDecoder(ws)

	var offerer bool
	err = dec.Decode(&offerer)
	chk(err)

	cfg := webrtc.Config{func(cand webrtc.ICECandidate) {
		err := enc.Encode(wsMessage{ICECandidate: cand})
		chk(err)
	}}
	c := webrtc.NewConn(cfg)
	go play(offerer, c.CreateDataChannel("gameState", webrtc.Negotiated(0)))
	go negotiateSession(offerer, c, enc, dec)
}

func negotiateSession(offerer bool, c webrtc.Conn, enc *json.Encoder, dec *json.Decoder) {
	if offerer {
		offer, err := c.CreateOffer()
		chk(err)
		err = c.SetLocalDescription(offer)
		chk(err)
		err = enc.Encode(wsMessage{SessionDescription: offer})
		chk(err)
	}

	for {
		var msg wsMessage
		err := dec.Decode(&msg)
		if err == io.EOF {
			break
		}
		chk(err)
		if msg.SessionDescription.Valid() {
			fmt.Println("received", msg.SessionDescription.Type)
			err := c.SetRemoteDescription(msg.SessionDescription)
			chk(err)
			if !offerer {
				answer, err := c.CreateAnswer()
				chk(err)
				err = c.SetLocalDescription(answer)
				chk(err)
				err = enc.Encode(wsMessage{SessionDescription: answer})
				chk(err)
			}
		} else {
			fmt.Println("received", msg.ICECandidate)
			err := c.AddICECandidate(msg.ICECandidate)
			chk(err)
		}
	}
}

type wsMessage struct {
	SessionDescription webrtc.SessionDescription
	ICECandidate       webrtc.ICECandidate
}

func play(offerer bool, dc webrtc.DataChannel) {
	var table table
	table.reset()
	me := &table.player1
	you := &table.player2
	if offerer {
		me, you = you, me
	}

	doc := dom.GetWindow().Document()
	player1 := doc.GetElementByID("player1")
	player2 := doc.GetElementByID("player2")
	ball := doc.GetElementByID("ball")
	var left, right, quit bool
	doc.AddEventListener("keydown", false, func(event dom.Event) {
		e := event.(*dom.KeyboardEvent)
		switch e.Key {
		case "ArrowLeft":
			left = true
		case "ArrowRight":
			right = true
		case "Escape":
			quit = true
		}
	})
	doc.AddEventListener("keyup", false, func(event dom.Event) {
		e := event.(*dom.KeyboardEvent)
		switch e.Key {
		case "ArrowLeft":
			left = false
		case "ArrowRight":
			right = false
		}
	})

	for !quit {
		next := time.After(time.Second / fps)

		move := "nowhere"
		if left && !right {
			move = "left"
			me.dx -= 1
		}
		if !left && right {
			move = "right"
			me.dx += 1
		}
		dc.SendString(move)
		move = dc.Recv()
		if move == "left" {
			you.dx -= 1
		}
		if move == "right" {
			you.dx += 1
		}

		table.step()

		// TODO: send+recv table state, validate

		player1.SetAttribute("x1", fmt.Sprintf("%vcm", table.player1.x-1))
		player1.SetAttribute("x2", fmt.Sprintf("%vcm", table.player1.x+1))
		player2.SetAttribute("x1", fmt.Sprintf("%vcm", table.player2.x-1))
		player2.SetAttribute("x2", fmt.Sprintf("%vcm", table.player2.x+1))
		ball.SetAttribute("cx", fmt.Sprintf("%vcm", table.ball.x))
		ball.SetAttribute("cy", fmt.Sprintf("%vcm", table.ball.y))

		select {
		case <-next:
			fmt.Println("too slow")
		default:
			<-next
		}
	}
}

const (
	fps         = 60
	tableWidth  = 10
	tableHeight = 16
	paddleWidth = .5
	ballRadius  = .3
)

type table struct {
	player1, player2, ball object
}

func (t *table) reset() {
	t.player1.x = tableWidth / 2
	t.player2.x = tableWidth / 2
	t.ball.x = tableWidth / 2
	t.ball.y = tableHeight / 2
	t.ball.dx = 0
	t.ball.dy = 8
}

func (t *table) step() {
	t.player1.dx *= .9
	t.player2.dx *= .9
	t.player1.x += t.player1.dx / fps
	t.player2.x += t.player2.dx / fps
	if t.player1.x < 1 || t.player1.x > tableWidth-1 {
		t.player1.x = math.Max(1, math.Min(t.player1.x, tableWidth-1))
		t.player1.dx = -t.player1.dx
	}
	if t.player2.x < 1 || t.player2.x > tableWidth-1 {
		t.player2.x = math.Max(1, math.Min(t.player2.x, tableWidth-1))
		t.player2.dx = -t.player2.dx
	}

	t.ball.x += t.ball.dx / fps
	t.ball.y += t.ball.dy / fps
	if t.ball.x < ballRadius || t.ball.x > tableWidth-ballRadius {
		t.ball.dx = -t.ball.dx
	}
	if t.ball.y < paddleWidth+ballRadius && t.ball.x > t.player1.x-1 && t.ball.x < t.player1.x+1 ||
		t.ball.y > tableHeight-(paddleWidth+ballRadius) && t.ball.x > t.player2.x-1 && t.ball.x < t.player2.x+1 {
		t.ball.dy = -t.ball.dy
	}
	if t.ball.y < 0 || t.ball.y > tableHeight {
		t.reset()
	}
}

type object struct {
	x, y, dx, dy float64
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
