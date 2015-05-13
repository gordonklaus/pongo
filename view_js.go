//+build js

package main

import (
	"fmt"

	"honnef.co/go/js/dom"
)

var doc = dom.GetWindow().Document()

func onKey(f keyFunc) {
	doc.AddEventListener("keydown", false, func(event dom.Event) {
		e := event.(*dom.KeyboardEvent)
		f(keyEvent{keyDown, e.KeyCode})
	})
	doc.AddEventListener("keyup", false, func(event dom.Event) {
		e := event.(*dom.KeyboardEvent)
		f(keyEvent{keyUp, e.KeyCode})
	})
}

type paddleView struct {
	line, left, right dom.Element
}

func newPaddleView(id string) *paddleView {
	s := "player" + id
	return &paddleView{
		line: doc.GetElementByID(s),
		left: doc.GetElementByID(s + "left"),
		right: doc.GetElementByID(s + "right"),
	}
}

func (v *paddleView) update(p paddle) {
	v.line.SetAttribute("x1", fmt.Sprintf("%vcm", p.left().x))
	v.line.SetAttribute("x2", fmt.Sprintf("%vcm", p.right().x))
	v.left.SetAttribute("cx", fmt.Sprintf("%vcm", p.left().x))
	v.right.SetAttribute("cx", fmt.Sprintf("%vcm", p.right().x))
}

type ballView struct {
	circle dom.Element
}

func newBallView() *ballView {
	return &ballView{
		circle: doc.GetElementByID("ball"),
	}
}

func (v *ballView) update(b ball) {
	v.circle.SetAttribute("cx", fmt.Sprintf("%vcm", b.pos.x))
	v.circle.SetAttribute("cy", fmt.Sprintf("%vcm", b.pos.y))
}