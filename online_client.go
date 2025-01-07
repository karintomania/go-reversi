package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var _ = time.Second //TODO: debugging

// host a game server
type OnlineHostClient struct {
	gameCh    chan Game
	gameCmdCh chan GameCommand
	PlayerId  PlayerId
}

func (c *OnlineHostClient) Run() {

	var wg sync.WaitGroup

	upgrader := websocket.Upgrader{}

	handler := func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			fmt.Printf("err %v", err)
			return
		}
		defer conn.Close()

		// Read message from browser
		go func() {
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

				c.gameCmdCh <- cmd

				switch cmd.CommandType {
				case CommandQuit:
					wg.Done()
				}
			}
		}()

		for {
			// Write game
			g := <-c.gameCh

			err = conn.WriteJSON(g)
			if err != nil {
				break
			}
		}
	}

	http.HandleFunc("/", handler)

	server := &http.Server{
		Addr: ":8089",
	}

	wg.Add(1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Error starting server:", err)
		}
	}()

	wg.Wait()

	// gracefully shutdown the server
	err := server.Shutdown(context.TODO())
	if err != nil {
		fmt.Errorf("Failed to shutdown server: %v", err)
	}
}

// guest client which sends command to server
type OnlineGuestClient struct {
	gameCh    chan Game
	gameCmdCh chan GameCommand
	PlayerId  PlayerId
}

func (c *OnlineGuestClient) Run() {
AiClientLoop:
	for g := range c.gameCh {
		if g.IsMyTurn(c.PlayerId) {
			p := getAiPosition(g.Board)
			cmd := GameCommand{CommandPlace, p}
			// time.Sleep(5 * time.Second)
			c.gameCmdCh <- cmd
		}

		if g.State == Quit {
			break AiClientLoop
		}
	}
}
