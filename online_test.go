package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestWebConnections(t *testing.T) {
	// if these tests run parallelly, it causes address conflict
	testWebHostConnectionSendCommand(t)

	// wait for server to close
	time.Sleep(50 * time.Millisecond)

	testWebGuestConnection(t)
}

func testWebHostConnectionSendCommand(t *testing.T) {
	gameCh := make(chan Game)
	cmdCh := make(chan GameCommand)
	quitCh := make(chan bool)

	client := OnlineHostConnection{gameCh, cmdCh, quitCh, 8089}

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

	got := <-cmdCh

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
	cmd = GameCommand{Quit: true}
	conn.WriteJSON(cmd)

	gotQuit := <-quitCh

	assert.Equal(t, true, gotQuit)
}

func testWebGuestConnection(t *testing.T) {
	hostGameCh := make(chan Game)
	hostCmdCh := make(chan GameCommand)

	server := mockHost(t, hostGameCh, hostCmdCh)

	defer server.Close()

	// wait for server to start
	time.Sleep(50 * time.Millisecond)

	conn, gameCh, cmdCh, quitCh := NewOnlineGuestConnection(Player2Id)

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
		Addr:    ":8089",
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Error starting server:%v", err)
		}
	}()

	return server
}
