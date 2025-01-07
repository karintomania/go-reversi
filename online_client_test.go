package main

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestOnlineHostClientSendCommand(t *testing.T) {

	gameCh := make(chan Game)
	gameCmdCh := make(chan GameCommand)

	client := OnlineHostClient{gameCh, gameCmdCh, Player2Id}

	go client.Run()

	// wait for server to start
	time.Sleep(50 * time.Millisecond)

	url := "ws://localhost:8089/"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Errorf("Dial error: %v", err)
	}

	defer conn.Close()

	cmd := GameCommand{CommandType: CommandPlace, Position: Position{1, 1}}
	conn.WriteJSON(cmd)

	got := <-gameCmdCh

	assert.Equal(t, cmd.CommandType, got.CommandType)
	assert.Equal(t, cmd.Position.X, got.Position.X)
	assert.Equal(t, cmd.Position.Y, got.Position.Y)

	// init game
	b := Board{}
	b.init(3)

	g := NewGame(&b, Human, Human)
	g.State = Player1Turn

	// send game
	gameCh <- g

	// receive game
	var receivedGame Game
	conn.ReadJSON(&receivedGame)

	assert.Equal(t, Player1Turn, receivedGame.State)

	// quit game
	cmd = GameCommand{CommandType: CommandQuit}
	conn.WriteJSON(cmd)

	got = <-gameCmdCh

	assert.Equal(t, cmd.CommandType, got.CommandType)
}
