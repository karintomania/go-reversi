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
	gameCh   chan Game
	cmdCh    chan GameCommand
	quitCh   chan bool
	PlayerId PlayerId
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

				if cmd.Quit {
					c.quitCh <- true
					wg.Done()
				} else {
					c.cmdCh <- cmd
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
	PlayerId PlayerId
	d        *Display
	p        *Position
}

func (c *OnlineGuestClient) Run() {
	// establish connection
	url := "ws://localhost:8089/"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Errorf("Dial error: %v", err)
	}

	// init game
	var g Game

	input := make(chan string)
	defer close(input)

	go func() {
		for {
			c.d.Read(input)
		}
	}()

	go func() {
		for char := range input {
			if g.IsMyTurn(c.PlayerId) {
				switch char {
				// move position
				case "h", "a": // ←
					c.p.addX(-1, g.Board.N)
					c.d.Render(&g, *c.p)
				case "l", "d": // →
					c.p.addX(1, g.Board.N)
					c.d.Render(&g, *c.p)
				case "j", "s": // ↓
					c.p.addY(1, g.Board.N)
					c.d.Render(&g, *c.p)
				case "k", "w": // ↑
					c.p.addY(-1, g.Board.N)
					c.d.Render(&g, *c.p)

				// place
				case " ":
					cmd := GameCommand{CommandType: CommandPlace, Position: *c.p}
					conn.WriteJSON(cmd)
				}
			}

			if g.State == Finished {
				switch char {
				case "r":
					cmd := GameCommand{CommandType: CommandReplay}
					conn.WriteJSON(cmd)
				}
			}

			if char == "c" {
				cmd := GameCommand{Quit: true}
				go conn.WriteJSON(cmd)
			}
		}
	}()

onlineGuestClientLoop:
	for {
		conn.ReadJSON(&g)

		if g.State == WaitingConnection {
			conn.WriteJSON(GameCommand{CommandType: CommandConnectionCheck})
		}

		if g.State == Quit {
			break onlineGuestClientLoop
		}

		c.d.Render(&g, *c.p)
	}
}
