//+build js

package main

import (
	// "encoding/json"
	"fmt"
	// "io/ioutil"
	// "net/http"
	// "time"

	// "github.com/gopherjs/jquery"
	"github.com/gopherjs/websocket"
	"github.com/gordonklaus/webrtc"
)

func main() {
	conn, err := websocket.Dial("ws://localhost:12345/ws")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	buf := [1]byte{}
	n, err := conn.Read(buf[:])
	if n != 1 || err != nil {
		panic(fmt.Sprint(n, err))
	}
	offerer := buf[0] == 0
	pc := webrtc.NewConn()
	if offerer {
		pc.CreateOffer()
	} else {
		
	}

	// for {
	// 	resp, err := http.Get("http://localhost:12345/table")
	// 	if err != nil {
	// 		panic(err)
	// 	}
	//
	// 	tableBytes, err := ioutil.ReadAll(resp.Body)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	//
	// 	err = json.Unmarshal(tableBytes, &table)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	//
	// 	left.SetAttr("y1", fmt.Sprintf("%vcm", table.Left-1))
	// 	left.SetAttr("y2", fmt.Sprintf("%vcm", table.Left+1))
	// 	right.SetAttr("y1", fmt.Sprintf("%vcm", table.Right-1))
	// 	right.SetAttr("y2", fmt.Sprintf("%vcm", table.Right+1))
	// 	ball.SetAttr("cx", fmt.Sprintf("%vcm", table.BallX))
	// 	ball.SetAttr("cy", fmt.Sprintf("%vcm", table.BallY))
	// 	time.Sleep(time.Second / fps)
	// }
}

// type Table struct {
// 	Left, Right, BallX, BallY float64
// 	ballDX, ballDY            float64
// }
//
// func (t *Table) reset() {
// 	t.Left = tableHeight / 2
// 	t.Right = tableHeight / 2
// 	t.BallX = tableWidth / 2
// 	t.BallY = tableHeight / 2
// 	angle := 2 * math.Pi * rand.Float64()
// 	t.ballDX = 5 * math.Cos(angle)
// 	t.ballDY = 5 * math.Sin(angle)
// }
//
// var (
// 	jQuery = jquery.NewJQuery
//
// 	left  = jQuery("line#left")
// 	right = jQuery("line#right")
// 	ball  = jQuery("circle#ball")
// 	table Table
// )
//
// const (
// 	fps         = 30
// 	tableWidth  = 16
// 	tableHeight = 10
// 	paddleWidth = .5
// 	ballRadius  = .3
// )
//
// func animate() {
// 	for {
// 		table.Lock()
// 		table.BallX += table.ballDX / fps
// 		table.BallY += table.ballDY / fps
// 		if table.BallY < ballRadius || table.BallY > tableHeight-ballRadius {
// 			table.ballDY = -table.ballDY
// 		}
// 		if table.BallX < paddleWidth+ballRadius && table.BallY > table.Left-1 && table.BallY < table.Left+1 ||
// 			table.BallX > tableWidth-(paddleWidth+ballRadius) && table.BallY > table.Right-1 && table.BallY < table.Right+1 {
// 			table.ballDX = -table.ballDX
// 		}
// 		if table.BallX < 0 || table.BallX > tableWidth {
// 			table.reset()
// 		}
// 		table.Unlock()
//
// 		time.Sleep(time.Second / fps)
// 	}
// }
