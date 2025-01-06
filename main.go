package main

import (
	"flag"
	"fmt"
	"sync"
)

var _ string = fmt.Sprint("test")

const (
	DEFAULT_N = 3
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

	player1CmdCh, player2CmdCh, player1GameCh, player2GameCh := g.Start()

	cli1 := Client{
		stdin:     input,
		gameCh:    player1GameCh,
		gameCmdCh: player1CmdCh,
		d:         &d,
		p:         &Position{},
	}

	cli2 := AiClient{
		stdin:     nil,
		gameCh:    player2GameCh,
		gameCmdCh: player2CmdCh,
		d:         nil,
		p:         &Position{},
	}

	var wg sync.WaitGroup
	// finish if one of the players quit
	wg.Add(2)

	go func() {
		cli1.Run()
		wg.Done()
	}()

	go func() {
		cli2.Run()
		wg.Done()
	}()

	wg.Wait()

	close(player1CmdCh)
	close(player2CmdCh)
	close(player1GameCh)
	close(player2GameCh)
}
