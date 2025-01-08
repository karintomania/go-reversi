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

type GameMode int

const (
	Single GameMode = iota
	LocalMulti
	OnlineHost
	OnlineGuest
)

func main() {
	n := flag.Int("n", DEFAULT_N, "Dimension of the board. (Default: 8)")
	playerNum := flag.Int("p", 1, "1 for Single Play, 2 for 2 Players. (Default: 1)")
	server := flag.Bool("s", false, "Start game with server")
	url := flag.String("url", "", "Specify game server url to connect")

	flag.Parse()

	gm := Single

	if *server {
		gm = OnlineHost
	} else if *url != "" {
		gm = OnlineGuest
	} else if *playerNum == 2 {
		gm = LocalMulti
	}

	switch gm {
	case Single: // 2 players
		startLocalSingleGame(*n)

	case LocalMulti: // 2 players
		startLocalMultiGame(*n)

	case OnlineHost:
		startHostClient(*n)

	case OnlineGuest:
		startGuestClient(*url)

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

	player1CmdCh, player2CmdCh, player1GameCh, player2GameCh, player1QuitCh, player2QuitCh := g.Start()

	inputCh := make(chan string)

	go func() {
		for {
			d.Read(inputCh)
		}
	}()

	// local single
	cli1 := NewLocalClient(
		player1GameCh,
		player1CmdCh,
		player1QuitCh,
		inputCh,
		Player1Id,
		&d,
	)

	cli2 := AiClient{
		gameCh:   player2GameCh,
		cmdCh:    player2CmdCh,
		quitCh:   player2QuitCh,
		PlayerId: Player2Id,
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

	inputCh := make(chan string)

	go func() {
		for {
			d.Read(inputCh)
		}
	}()

	g := NewGame(&b, Human, Human)

	player1CmdCh, player2CmdCh, player1GameCh, player2GameCh, player1QuitCh, player2QuitCh := g.Start()

	cli := NewLocalMultiClient(
		player1GameCh,
		player1CmdCh,
		player1QuitCh,
		player2GameCh,
		player2CmdCh,
		player2QuitCh,
		inputCh,
		&d,
	)

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
	close(player1QuitCh)
	close(player2QuitCh)
}

func startHostClient(n int) {
	var b Board

	b.init(n)

	d := NewDisplay()
	defer d.Close()

	g := NewGame(&b, Human, AI)

	player1CmdCh, player2CmdCh, player1GameCh, player2GameCh, player1QuitCh, player2QuitCh := g.Start()

	p := &Position{}

	cli1 := LocalClient{
		gameCh:   player1GameCh,
		cmdCh:    player1CmdCh,
		quitCh:   player1QuitCh,
		PlayerId: Player1Id,
		d:        &d,
		p:        p,
	}

	cli2 := OnlineHostClient{
		gameCh:   player2GameCh,
		cmdCh:    player2CmdCh,
		quitCh:   player2QuitCh,
		PlayerId: Player2Id,
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
	close(player1QuitCh)
	close(player2QuitCh)
	close(player1GameCh)
	close(player2GameCh)
}

func startGuestClient(url string) {
	_ = url
	d := NewDisplay()
	defer d.Close()

	cli := OnlineGuestClient{
		PlayerId: Player2Id,
		d:        &d,
		p:        &Position{},
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		cli.Run()
		wg.Done()
	}()

	wg.Wait()
}
