package main

import (
	"fmt"
	"log/slog"
	"time"
)

type HostStarter struct {
	d       Renderer
	g       Game
	inputCh chan string
}

func (hs *HostStarter) Start(n int, port int) {
	var b Board

	b.init(n)

	hs.g = NewGame(&b, Human, Human)

	player1CmdCh, hostCmdCh, player1GameCh, hostGameCh, player1QuitCh, hostQuitCh := hs.g.Start()

	closeCh := make(chan bool)

	cli1 := NewLocalClient(
		player1GameCh,
		player1CmdCh,
		player1QuitCh,
		hs.inputCh,
		closeCh,
		Player1Id,
		hs.d,
	)

	hostConn := NewOnlineHostConnection(
		hostGameCh,
		hostCmdCh,
		hostQuitCh,
		port,
	)

	go func() {
		cli1.Run()
		logger.Debug("Client quit")
	}()

	go func() {
		hostConn.Run()
		logger.Debug("Host closed")
	}()

	<-closeCh

	// wait for sending quit signal
	time.Sleep(500 * time.Millisecond)
	hostConn.Close()
}

type GuestStarter struct {
	d       Renderer
	inputCh chan string
}

func (gs *GuestStarter) Start(url string, port int) {

	id := Player2Id

	conn, gameCh, cmdCh, quitCh := NewOnlineGuestConnection(id, url, port)

	closeCh := make(chan bool)

	cli := NewLocalClient(gameCh, cmdCh, quitCh, gs.inputCh, closeCh, id, gs.d)

	go func() {
		fmt.Println("Trying to connect to the server...")
		cli.Run()
		logger.Debug("Client closed")
	}()

	go func() {
		logger.Debug("Starting online guest")
		err := conn.Run()
		if err != nil {
			fmt.Printf("\rCan't connect %s.\n", conn.Url)
			logger.Debug("Error on guest conn", slog.Any("err", err))
		}
		logger.Debug("Guest conn closed")
	}()

	<-closeCh

	// wait for sending quit signal
	time.Sleep(500 * time.Millisecond)
	conn.Close()
}
