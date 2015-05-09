//+build js

package main

import (
	"fmt"
	"math"
	"time"

	"github.com/gordonklaus/webrtc"
	"honnef.co/go/js/dom"
)

var (
	initiator bool
	dcChan = make(chan webrtc.DataChannel)
)

func main() {
	go negotiateSession()
	for {
		play()
	}
}

func negotiateSession() {
start:
	sig := openSignallingChannel()
	// defer ws.Close()

	initiator = sig.initiator

	fmt.Println("initiator:", initiator)

	cfg := webrtc.Config{}
	c := webrtc.NewConn(cfg)
	dc := c.CreateDataChannel("gameState", webrtc.Negotiated(0))
	dcChan <- dc

	if initiator {
		offer, err := c.CreateOffer()
		chk(err)
		err = c.SetLocalDescription(offer)
		chk(err)
		err = sig.enc.Encode(wsMessage{SessionDescription: &offer})
		chk(err)
	}
	localICECandidates := c.ICECandidates
	remoteICECandidates := sig.iceCandidates
	sessionDescriptions := sig.sessionDescriptions
	for {
		select {
		case ic := <-localICECandidates:
			if ic == nil {
				localICECandidates = nil
			}
			fmt.Println("sending", ic)
			err := sig.enc.Encode(wsMessage{ICECandidate: ic})
			chk(err)
		case ic, ok := <-remoteICECandidates:
			if !ok {
				remoteICECandidates = nil
				break
			}
			fmt.Println("received", ic)
			err := c.AddICECandidate(ic)
			chk(err)
		case sd := <-sessionDescriptions:
			if sd.Type != webrtc.ProvisionalAnswer {
				sessionDescriptions = nil
			}
			fmt.Println("received", sd.Type)
			err := c.SetRemoteDescription(sd)
			chk(err)
			if !initiator {
				answer, err := c.CreateAnswer()
				chk(err)
				err = c.SetLocalDescription(answer)
				chk(err)
				err = sig.enc.Encode(wsMessage{SessionDescription: &answer})
				chk(err)
			}
		case <-sig.closed:
			fmt.Println("signalling channel closed")
			dc.Close()
			c.Close()
			goto start
		case state := <-c.ICEConnectionState:
			fmt.Println("ICE connection state:", state)
			switch {
			case state.Completed():
				// sig.close()
			case state.Failed():
				// dc.Close()
				// c.Close()
				// goto start
			}
		}
	}
}

func play() {
	dc := <-dcChan

	var table table
	table.reset()
	me := &table.player1
	you := &table.player2
	if initiator {
		me, you = you, me
	}

	doc := dom.GetWindow().Document()
	player1 := doc.GetElementByID("player1")
	player2 := doc.GetElementByID("player2")
	ball := doc.GetElementByID("ball")
	var left, right, quit bool
	// TODO: key names not recognized on Chrome?
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
		err := dc.SendString(move)
		if err != nil {
			fmt.Println("DataChannel.SendString:", err)
			break
		}
		move, err = dc.Recv()
		if err != nil {
			fmt.Println("DataChannel.Recv:", err)
			break
		}
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
	tableHeight = 12
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
