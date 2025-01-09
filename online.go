package main

import (
	// "context"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var _ = time.Second //TODO: debugging

// host a game server
type OnlineHostConnection struct {
	gameCh <-chan Game
	cmdCh  chan<- GameCommand
	quitCh chan<- bool
	Port   int
}

func (c *OnlineHostConnection) Run() {

	var wg sync.WaitGroup

	wg.Add(1)
	handler := func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("err %v", err)
			return
		}
		defer conn.Close()

		// Send game info to guest
		go func() {
			for g := range c.gameCh {
				err = conn.WriteJSON(g)
				if err != nil {
					fmt.Errorf("%v", err)
					break
				}
			}
		}()

		// Receive command from guest
		for {
			cmd := GameCommand{}
			err := conn.ReadJSON(&cmd)
			if err != nil {
				// show error when websocket is closed unexpectedly
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					fmt.Errorf("WebSocket Error %v", err)
				}
				break
			}

			if cmd.Quit {
				go func() { c.quitCh <- true }()
			} else {
				fmt.Println(cmd.CommandType)
				c.cmdCh <- cmd
			}

		}

		wg.Done()

	}

	// TODO: using mux to pass tests. Fix if it can be just http
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting server: %v\n", err)
		}
	}()

	wg.Wait()

	// gracefully shutdown the server
	err := server.Shutdown(context.TODO())
	if err != nil {
		fmt.Errorf("Failed to shutdown server: %v", err)
	}
	fmt.Println("server closed")
}

// guest client sends command to host
type OnlineGuestConnection struct {
	gameCh   chan<- Game
	cmdCh    <-chan GameCommand
	quitCh   <-chan bool
	PlayerId PlayerId
}

func NewOnlineGuestConnection(id PlayerId) (OnlineGuestConnection, chan Game, chan GameCommand, chan bool) {
	gameCh := make(chan Game)
	cmdCh := make(chan GameCommand)
	quitCh := make(chan bool)

	conn := OnlineGuestConnection{
		gameCh,
		cmdCh,
		quitCh,
		id,
	}

	return conn, gameCh, cmdCh, quitCh
}

func (c *OnlineGuestConnection) Run() error {
	// establish connection
	url := "ws://localhost:8089/"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("Dial error: %v", err)
	}

	defer conn.Close()

	// listen to command
	go func() {
		for cmd := range c.cmdCh {
			conn.WriteJSON(cmd)
		}
	}()

	// listen to quit
	go func() {
		for quit := range c.quitCh {
			if quit {
				cmd := GameCommand{Quit: true}
				conn.WriteJSON(cmd)
			}
		}
	}()

	// init game
	var g Game

	// listen to server's game sync
onlineGuestClientLoop:
	for {
		if err := conn.ReadJSON(&g); err != nil {
			return fmt.Errorf("Failed to read: %v", err)
		}

		c.gameCh <- g

		if g.State == WaitingConnection {
			conn.WriteJSON(GameCommand{CommandType: CommandConnectionCheck})
		}

		if g.State == Quit {
			break onlineGuestClientLoop
		}
	}

	return nil
}
