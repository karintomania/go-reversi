package main

import (
	"log/slog"
	// "sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEndToEnd(t *testing.T) {
	logger = NewLogger(slog.LevelInfo)

	// Initialize the display and player input channels
	d := MockDisplay{}
	defer d.Close()

	player1InputCh := make(chan string)
	player2InputCh := make(chan string)

	// Initialize HostStarter
	hostStarter := HostStarter{
		d:       &d,
		inputCh: player1InputCh,
	}

	// Start the host
	go hostStarter.Start(3, DEFAULT_PORT)

	time.Sleep(500 * time.Millisecond)

	// Initialize GuestStarter
	guestStarter := GuestStarter{
		d:       &d,
		inputCh: player2InputCh,
	}

	// Start the guest
	go guestStarter.Start("http://localhost", DEFAULT_PORT)

	// Allow some time for connections to establish
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, Player1Turn.String(), hostStarter.g.State.String())

	t.Log(hostStarter.g.Board.String())

	// Simulate game moves
	player1InputCh <- "d"
	player1InputCh <- "d" // (2,0)
	player1InputCh <- " " // place
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Player2Turn.String(), hostStarter.g.State.String())

	player2InputCh <- "d"
	player2InputCh <- "d"
	player2InputCh <- "s" // (2,1)
	player2InputCh <- " " // place
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Player1Turn.String(), hostStarter.g.State.String())

	t.Log(hostStarter.g.Board.String())

	player1InputCh <- "s"
	player1InputCh <- "s" // (2,2)
	player1InputCh <- " " // place
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Player1Turn.String(), hostStarter.g.State.String())

	t.Log(hostStarter.g.Board.String())
	player1InputCh <- "a"
	player1InputCh <- "a" // (0,2)
	player1InputCh <- " " // place
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, Finished.String(), hostStarter.g.State.String())

	t.Log(hostStarter.g.Board.String())
	player1InputCh <- "r" // replay
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Player2Turn.String(), hostStarter.g.State.String())

	t.Log(hostStarter.g.Board.String())
	player2InputCh <- "w" // (2,0)
	player2InputCh <- " " // place
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Player1Turn.String(), hostStarter.g.State.String())

	t.Log(hostStarter.g.Board.String())
	player1InputCh <- "d"
	player1InputCh <- "d"
	player1InputCh <- "w" // (2,1)
	player1InputCh <- " " // place
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Player2Turn.String(), hostStarter.g.State.String())

	t.Log(hostStarter.g.Board.String())
	player2InputCh <- "s"
	player2InputCh <- "s" // (2,2)
	player2InputCh <- " " // place
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Player2Turn.String(), hostStarter.g.State.String())

	t.Log(hostStarter.g.Board.String())

	player2InputCh <- "a"
	player2InputCh <- "a" // (2,2)
	player2InputCh <- " " // place
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Finished.String(), hostStarter.g.State.String())

	t.Log(hostStarter.g.Board.String())

	// Close clients
	player1InputCh <- "c"
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Quit.String(), hostStarter.g.State.String())

	player2InputCh <- "c"

}
