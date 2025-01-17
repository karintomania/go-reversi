package main

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var _ = fmt.Sprint("")

type MockDisplay struct {
	g *Game
	p Position
}

func (m *MockDisplay) Render(g *Game, p Position) {
	m.g = g
	m.p = p
}

func (m *MockDisplay) Close() {
}

func TestLocalClientMovesPosition(t *testing.T) {
	gameCh, _, _, inputCh, _, d, client := localClientTestInitChannels()

	go client.Run()

	g := NewGame(NewBoard(3), Human, Human)
	g.State = Player1Turn

	gameCh <- g

	// move right
	inputCh <- "d"
	time.Sleep(5 * time.Millisecond)
	assert.Equal(t, Position{1, 0}, d.p)

	// move down
	inputCh <- "s"
	time.Sleep(5 * time.Millisecond)
	assert.Equal(t, Position{1, 1}, d.p)

	// move left
	inputCh <- "a"
	time.Sleep(5 * time.Millisecond)
	assert.Equal(t, Position{0, 1}, d.p)

	// move up
	inputCh <- "w"
	time.Sleep(5 * time.Millisecond)
	assert.Equal(t, Position{0, 0}, d.p)
}

func TestLocalClientPlaceDisk(t *testing.T) {
	gameCh, cmdCh, _, inputCh, _, _, client := localClientTestInitChannels()

	go client.Run()

	b := NewBoard(3)

	g := NewGame(b, Human, Human)
	g.State = Player1Turn

	gameCh <- g

	// place disk
	inputCh <- " "
	time.Sleep(100 * time.Millisecond)
	cmd := <-cmdCh

	assert.Equal(t, CommandPlace, cmd.CommandType)
	assert.Equal(t, Position{0, 0}, cmd.Position)
}

func TestLocalClientQuit(t *testing.T) {
	_, _, quitCh, inputCh, _, _, client := localClientTestInitChannels()

	go client.Run()

	// place disk
	inputCh <- "c"
	time.Sleep(100 * time.Millisecond)
	quit := <-quitCh

	assert.Equal(t, true, quit)
}

func localClientTestInitChannels() (
	chan Game,
	chan GameCommand,
	chan bool,
	chan string,
	chan bool,
	*MockDisplay, LocalClient) {
	logger = NewLogger(slog.LevelInfo)

	gameCh := make(chan Game)
	cmdCh := make(chan GameCommand)
	quitCh := make(chan bool)
	inputCh := make(chan string)
	closeCh := make(chan bool)

	d := MockDisplay{}

	client := NewLocalClient(gameCh, cmdCh, quitCh, inputCh, closeCh, Player1Id, &d)

	return gameCh, cmdCh, quitCh, inputCh, closeCh, &d, client
}
