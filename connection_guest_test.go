package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestWebGuestConnection(t *testing.T) {
	logger = NewLogger(slog.LevelInfo)

	muTest.Lock()
	defer func() {
		muTest.Unlock()
		// wait for server to shut down
		time.Sleep(50 * time.Millisecond)
	}()

	hostGameCh := make(chan Game)
	hostCmdCh := make(chan GameCommand)

	server := mockHost(t, hostGameCh, hostCmdCh)

	defer server.Close()

	// wait for server to start
	time.Sleep(50 * time.Millisecond)

	conn, gameCh, cmdCh, quitCh := NewOnlineGuestConnection(Player2Id, "ws://localhost", DEFAULT_PORT)

	go func() {
		if err := conn.Run(); err != nil {
			t.Errorf("Client run failed: %v", err)
		}
	}()

	g := NewGame(NewBoard(3), Human, Human)

	g.State = Player2Turn

	hostGameCh <- g
	gotGame := <-gameCh
	assert.Equal(t, Player2Turn, gotGame.State)

	wantCmd := GameCommand{CommandType: CommandPlace, Position: Position{0, 0}}
	cmdCh <- wantCmd
	gotCmd := <-hostCmdCh

	assert.Equal(t, wantCmd.CommandType, gotCmd.CommandType)
	assert.Equal(t, wantCmd.Position, gotCmd.Position)

	quitCh <- true
	gotCmd = <-hostCmdCh
	assert.Equal(t, true, gotCmd.Quit)

}

func mockHost(t *testing.T, hostGameCh chan Game, hostCmdCh chan GameCommand) *http.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("err %v", err)
			return
		}
		defer conn.Close()

		// send game
		go func() {
			for g := range hostGameCh {
				err = conn.WriteJSON(g)
				if err != nil {
					t.Errorf("%v", err)
					break
				}
			}
		}()

		for {
			cmd := GameCommand{}
			if err := conn.ReadJSON(&cmd); err != nil {
				t.Errorf("%v", err)
				break
			}
			hostCmdCh <- cmd
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", DEFAULT_PORT),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Error starting server:%v", err)
		}
	}()

	return server
}
