package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"sync"
	"testing"
	"time"
)

var muTest sync.Mutex

func TestConnectionHostSendCommand(t *testing.T) {
	logger = NewLogger(slog.LevelInfo)

	muTest.Lock()
	defer func() {
		muTest.Unlock()
		// wait for server to shut down
		time.Sleep(50 * time.Millisecond)
	}()

	gameCh := make(chan Game)
	cmdCh := make(chan GameCommand)
	quitCh := make(chan bool)

	hostConn := NewOnlineHostConnection(gameCh, cmdCh, quitCh, DEFAULT_PORT)

	go hostConn.Run()

	// emulate game
	b := NewBoard(3)

	g := NewGame(b, Human, Human)
	g.State = Player1Turn

	gameCh <- g

	// mock guest conn
	time.Sleep(50 * time.Millisecond)

	url := fmt.Sprintf("ws://localhost:%d", DEFAULT_PORT)
	logger.Debug(url)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Errorf("Connection error: %v", err)
	}

	defer conn.Close()

	// test command sending
	cmd := GameCommand{CommandType: CommandPlace, Position: Position{1, 1}}
	conn.WriteJSON(cmd)

	got := <-cmdCh

	assert.Equal(t, cmd.CommandType, got.CommandType)
	assert.Equal(t, cmd.Position.X, got.Position.X)
	assert.Equal(t, cmd.Position.Y, got.Position.Y)

	// test receive game
	var receivedGame Game
	conn.ReadJSON(&receivedGame)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, Player1Turn.String(), receivedGame.State.String())

	// quit game
	cmd = GameCommand{Quit: true}
	conn.WriteJSON(cmd)

	gotQuit := <-quitCh

	assert.Equal(t, true, gotQuit)

	hostConn.Close()

	assert.Equal(t, true, true)
}
