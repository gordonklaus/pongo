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
	dcChan    = make(chan webrtc.DataChannel)
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
	player1left := doc.GetElementByID("player1left")
	player1right := doc.GetElementByID("player1right")
	player2 := doc.GetElementByID("player2")
	player2left := doc.GetElementByID("player2left")
	player2right := doc.GetElementByID("player2right")
	ball := doc.GetElementByID("ball")
	var left, right stickybool
	var quit bool
	doc.AddEventListener("keydown", false, func(event dom.Event) {
		e := event.(*dom.KeyboardEvent)
		switch e.KeyCode {
		case 37:
			left.set(true)
		case 39:
			right.set(true)
		case 27:
			quit = true
		}
	})
	doc.AddEventListener("keyup", false, func(event dom.Event) {
		e := event.(*dom.KeyboardEvent)
		switch e.KeyCode {
		case 37:
			left.set(false)
		case 39:
			right.set(false)
		}
	})

	for frame := 0; !quit; frame++ {
		next := time.After(time.Second / framesPerSecond)

		if frame%framesPerControl == 0 {
			ddx := 30.0 / controlsPerSecond
			move := "nowhere"
			left := left.get()
			right := right.get()
			if left && !right {
				move = "left"
				me.vel.x -= ddx
			}
			if !left && right {
				move = "right"
				me.vel.x += ddx
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
				you.vel.x -= ddx
			}
			if move == "right" {
				you.vel.x += ddx
			}
		}

		table.step()

		// TODO: send+recv table state, validate

		player1.SetAttribute("x1", fmt.Sprintf("%vcm", table.player1.left().x))
		player1.SetAttribute("x2", fmt.Sprintf("%vcm", table.player1.right().x))
		player1left.SetAttribute("cx", fmt.Sprintf("%vcm", table.player1.left().x))
		player1right.SetAttribute("cx", fmt.Sprintf("%vcm", table.player1.right().x))

		player2.SetAttribute("x1", fmt.Sprintf("%vcm", table.player2.left().x))
		player2.SetAttribute("x2", fmt.Sprintf("%vcm", table.player2.right().x))
		player2left.SetAttribute("cx", fmt.Sprintf("%vcm", table.player2.left().x))
		player2right.SetAttribute("cx", fmt.Sprintf("%vcm", table.player2.right().x))

		ball.SetAttribute("cx", fmt.Sprintf("%vcm", table.ball.pos.x))
		ball.SetAttribute("cy", fmt.Sprintf("%vcm", table.ball.pos.y))

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
	paddleRadius = .75
	ballRadius  = .3
)

type table struct {
	player1, player2 paddle
	ball             ball
}

func (t *table) reset() {
	t.player1.pos = vec2{tableWidth / 2, 0}
	t.player1.vel = vec2{}
	t.player1.width = 2

	t.player2.pos = vec2{tableWidth / 2, tableHeight}
	t.player2.vel = vec2{}
	t.player2.width = 2

	t.ball.pos = vec2{tableWidth / 2, tableHeight / 2}
	t.ball.vel = vec2{0, 8}
}

func (t *table) step() {
	t.player1.step()
	t.player2.step()

	t.ball.pos.x += t.ball.vel.x / framesPerSecond
	t.ball.pos.y += t.ball.vel.y / framesPerSecond
	if t.ball.pos.x < ballRadius || t.ball.pos.x > tableWidth-ballRadius {
		t.ball.vel.x = -t.ball.vel.x
	}
	collide(&t.ball, t.player1)
	collide(&t.ball, t.player2)
	if t.ball.pos.y < 0 || t.ball.pos.y > tableHeight {
		t.reset()
	}
}

func collide(ball *ball, player paddle) {
	const minDist = paddleRadius + ballRadius
	if math.Abs(ball.pos.y-player.pos.y) < minDist && ball.pos.x > player.left().x && ball.pos.x < player.right().x {
		ball.vel.y = -ball.vel.y
		return
	}
	for _, ppos := range []vec2{player.left(), player.right()} {
		d := ball.pos.sub(ppos)
		if d2 := d.len2(); d2 < minDist*minDist {
			ball.vel = ball.vel.sub(d.muls(2 * ball.vel.sub(player.vel).dot(d) / d2))
		}
	}
}

type paddle struct {
	pos, vel vec2
	width    float64
}

func (p paddle) left() vec2  { return vec2{p.pos.x - p.width/2, p.pos.y} }
func (p paddle) right() vec2 { return vec2{p.pos.x + p.width/2, p.pos.y} }

func (p *paddle) step() {
	p.vel.x -= 3 * p.vel.x / framesPerSecond
	p.pos.x += p.vel.x / framesPerSecond
	if p.left().x < paddleRadius || p.right().x > tableWidth-paddleRadius {
		p.vel.x = -p.vel.x
	}
}

type ball struct {
	pos, vel vec2
}

type vec2 struct {
	x, y float64
}

func (v vec2) sub(u vec2) vec2     { return vec2{v.x - u.x, v.y - u.y} }
func (v vec2) muls(s float64) vec2 { return vec2{v.x * s, v.y * s} }
func (v vec2) len2() float64       { return v.dot(v) }
func (v vec2) dot(u vec2) float64  { return v.x*u.x + v.y*u.y }

type stickybool struct {
	cur, next bool
}

func (s *stickybool) set(b bool) {
	if b {
		s.cur = true
	}
	s.next = b
}

func (s *stickybool) get() bool {
	b := s.cur
	s.cur = s.next
	return b
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
