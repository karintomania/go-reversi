package main

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEndToEnd(t *testing.T) {
	// start game
	b := NewBoard(3)

	d := MockDisplay{}
	defer d.Close()

	player1InputCh := make(chan string)

	g := NewGame(b, Human, AI)

	player1CmdCh, hostCmdCh, player1GameCh, hostGameCh, player1QuitCh, hostQuitCh := g.Start()

	cli1 := NewLocalClient(
		player1GameCh,
		player1CmdCh,
		player1QuitCh,
		player1InputCh,
		Player1Id,
		&d,
	)

	hostConn := OnlineHostConnection{
		gameCh: hostGameCh,
		cmdCh:  hostCmdCh,
		quitCh: hostQuitCh,
		Port:   8089,
	}

	player2InputCh := make(chan string)

	id := Player2Id

	guestConn, player2GameCh, player2CmdCh, player2QuitCh := NewOnlineGuestConnection(id)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		cli1.Run()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		hostConn.Run()
		wg.Done()
	}()

	go func() {
		// wait for the host to start
		time.Sleep(10 * time.Millisecond)
		if err := guestConn.Run(); err != nil {
			t.Errorf("Can't connect 'http://localhost:8089'. Press 'c' to finish.")
		}
	}()

	cli2 := NewLocalClient(player2GameCh, player2CmdCh, player2QuitCh, player2InputCh, id, &d)

	wg.Add(1)
	go func() {
		cli2.Run()
		wg.Done()
	}()

	// wait for game to receive connection check
	time.Sleep(1500 * time.Millisecond)
	assert.Equal(t, Player1Turn.String(), g.State.String())

	t.Log(g.Board.String())

	player1InputCh <- "d"
	player1InputCh <- "d" // (2,0)
	player1InputCh <- " " // placee
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Player2Turn.String(), g.State.String())

	player2InputCh <- "d"
	player2InputCh <- "d"
	player2InputCh <- "s" // (2,1)
	player2InputCh <- " " // placee
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Player1Turn.String(), g.State.String())

	t.Log(g.Board.String())

	player1InputCh <- "s"
	player1InputCh <- "s" // (2,2)
	player1InputCh <- " " // placee
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Player1Turn.String(), g.State.String())

	t.Log(g.Board.String())
	player1InputCh <- "a"
	player1InputCh <- "a" // (0,2)
	player1InputCh <- " " // placee
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, Finished.String(), g.State.String())

	t.Log(g.Board.String())
	player1InputCh <- "r" // replay
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Player2Turn.String(), g.State.String())

	t.Log(g.Board.String())
	player2InputCh <- "w" // (2,0)
	player2InputCh <- " " // placee
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Player1Turn.String(), g.State.String())

	t.Log(g.Board.String())
	player1InputCh <- "d"
	player1InputCh <- "d"
	player1InputCh <- "w" // (2,1)
	player1InputCh <- " " // placee
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Player2Turn.String(), g.State.String())

	t.Log(g.Board.String())
	player2InputCh <- "s"
	player2InputCh <- "s" // (2,2)
	player2InputCh <- " " // placee
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Player2Turn.String(), g.State.String())

	t.Log(g.Board.String())

	player2InputCh <- "a"
	player2InputCh <- "a" // (2,2)
	player2InputCh <- " " // placee
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Finished.String(), g.State.String())

	t.Log(g.Board.String())

	// close cli1
	player1InputCh <- "c"
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Quit.String(), g.State.String())

	// close cli2
	player2InputCh <- "c"

	wg.Wait()
}