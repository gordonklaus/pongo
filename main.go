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

	var err error
	initiator, err = sig.initiator()
	if err != nil {
		fmt.Println(err)
		sig.close()
		goto start
	}
	fmt.Println("initiator:", initiator)

	cfg := webrtc.Config{}
	c := webrtc.NewConn(cfg)
	dc := c.CreateDataChannel("gameState", webrtc.Negotiated(0))
	dcChan <- dc
	err = c.Negotiate(initiator, sig)
	sig.close()
	if err != nil {
		fmt.Println(err)
		dc.Close()
		c.Close()
		goto start
	}
	fmt.Println("negotiation succeeded")
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
	doc.AddEventListener("keydown", false, func(event dom.Event) {
		e := event.(*dom.KeyboardEvent)
		switch e.KeyCode {
		case 37:
			left = true
		case 39:
			right = true
		case 27:
			quit = true
		}
	})
	doc.AddEventListener("keyup", false, func(event dom.Event) {
		e := event.(*dom.KeyboardEvent)
		switch e.KeyCode {
		case 37:
			left = false
		case 39:
			right = false
		}
	})

	for frame := 0; !quit; frame++ {
		next := time.After(time.Second / framesPerSecond)

		if frame % framesPerControl == 0 {
			ddx := 60.0 / controlsPerSecond
			move := "nowhere"
			if left && !right {
				move = "left"
				me.dx -= ddx
			}
			if !left && right {
				move = "right"
				me.dx += ddx
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
				you.dx -= ddx
			}
			if move == "right" {
				you.dx += ddx
			}
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
	framesPerSecond   = 60
	controlsPerSecond = 10
	framesPerControl  = framesPerSecond / controlsPerSecond

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
	t.player1.x += t.player1.dx / framesPerSecond
	t.player2.x += t.player2.dx / framesPerSecond
	if t.player1.x < 1 || t.player1.x > tableWidth-1 {
		t.player1.x = math.Max(1, math.Min(t.player1.x, tableWidth-1))
		t.player1.dx = -t.player1.dx
	}
	if t.player2.x < 1 || t.player2.x > tableWidth-1 {
		t.player2.x = math.Max(1, math.Min(t.player2.x, tableWidth-1))
		t.player2.dx = -t.player2.dx
	}

	t.ball.x += t.ball.dx / framesPerSecond
	t.ball.y += t.ball.dy / framesPerSecond
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
