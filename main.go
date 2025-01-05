package main

import (
	"flag"
	"fmt"
)

var _ string = fmt.Sprint("test")

const (
	DEFAULT_N = 8
)

func main() {

	n := flag.Int("n", DEFAULT_N, "Dimension of the board. (Default: 8)")
	playerNum := flag.Int("p", 1, "1 for Single Play, 2 for 2 Players. (Default: 1)")

	flag.Parse()

	var b Board

	b.init(*n)

	d := NewDisplay()
	defer d.Close()

	input := make(chan string)
	defer close(input)

	go func() {
		for {
			d.Read(input)
		}
	}()

	var bPlayerType, wPlayerType PlayerType
	switch *playerNum {
	case 2: // 2 players
		bPlayerType = Human
		wPlayerType = Human

	case 0: // mainly for debugging
		bPlayerType = AI
		wPlayerType = AI

	default: // 1 player
		bPlayerType = Human
		wPlayerType = AI
	}

	g := NewGame(&b, bPlayerType, wPlayerType)

	gameCmdCh, gameCh := g.Start()

	cli := Client{
		stdin:     input,
		gameCh:    gameCh,
		gameCmdCh: gameCmdCh,
		d:         &d,
		p:         &Position{},
	}

	cli.Run()

}
