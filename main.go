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

	switch *playerNum {
	case 2: // 2 players
		startLocalMultiGame(*n)

	default: // 1 player
		startLocalSingleGame(*n)
	}
}

func startLocalSingleGame(n int) {
	var b Board

	b.init(n)

	d := NewDisplay()
	defer d.Close()

	g := NewGame(&b, Human, AI)

	player1CmdCh, player2CmdCh, player1GameCh, player2GameCh := g.Start()

	p := &Position{}

	// local single
	cli1 := Client{
		gameCh:    player1GameCh,
		gameCmdCh: player1CmdCh,
		PlayerId:  Player1Id,
		d:         &d,
		p:         p,
	}

	cli2 := AiClient{
		gameCh:    player2GameCh,
		gameCmdCh: player2CmdCh,
		PlayerId:  Player2Id,
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		cli1.Run()
		wg.Done()
	}()

	wg.Add(1)
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

func startLocalMultiGame(n int) {
	var b Board

	b.init(n)

	d := NewDisplay()
	defer d.Close()

	g := NewGame(&b, Human, Human)

	player1CmdCh, player2CmdCh, player1GameCh, player2GameCh := g.Start()

	p := &Position{}

	cli := LocalMultiClient{
		gameCh1:    player1GameCh,
		gameCmdCh1: player1CmdCh,
		gameCh2:    player2GameCh,
		gameCmdCh2: player2CmdCh,
		d:          &d,
		p:          p,
	}

	var wg sync.WaitGroup
	// finish if one of the players quit
	wg.Add(1)

	go func() {
		cli.Run()
		wg.Done()
	}()

	wg.Wait()

	close(player1CmdCh)
	close(player2CmdCh)
	close(player1GameCh)
	close(player2GameCh)
}
