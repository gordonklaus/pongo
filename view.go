package main

type keyFunc func(e keyEvent)

type keyEvent struct {
	action keyAction
	code   int
}

type keyAction uint8

const (
	keyDown keyAction = iota
	keyUp
)
