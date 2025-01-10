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

func (c *OnlineHostConnection) Run() error {
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(1)
	handler := func(w http.ResponseWriter, r *http.Request) {
		// establish websocket connection
		upgrader := websocket.Upgrader{}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			// TODO: handle error properly
			fmt.Errorf("Online Host Connection error %v", err)
		}

		defer func() {
			conn.Close()
			fmt.Println("\rHost: closed conn")
		}()

		// channel to detect if game is quit
		closeConnCh := make(chan bool)

		// Send game info to guest
		go func() {
			writeWithMutex := func(g Game) error {
				mu.Lock()
				defer mu.Unlock()
				return conn.WriteJSON(g)
			}

			for g := range c.gameCh {
				if err = writeWithMutex(g); err != nil {
					fmt.Errorf("%v", err)
					closeConnCh <- true
				}

				if g.State == Quit {
					closeConnCh <- true
				}
			}
		}()

		// Receive command from guest
		go func() {
			for {
				fmt.Printf("\rHost: waiting command\n")
				cmd := GameCommand{}
				err := conn.ReadJSON(&cmd)
				if err != nil {
					// show error when websocket is closed unexpectedly
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
						fmt.Printf("\rHost WebSocket Error %v\n", err)
					}
					closeConnCh <- true
				}
				fmt.Printf("\rHost: Command received: %v\n", cmd)

				if cmd.Quit {
					go func() { c.quitCh <- true }()
					fmt.Println("\rHost: quit sent to game")
				} else {
					go func() { c.cmdCh <- cmd }()
				}
			}
		}()

		// wait for closeConn signal
		<-closeConnCh
		wg.Done()
	}

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
		return fmt.Errorf("Failed to shutdown server: %v", err)
	}

	return nil
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
	var mu sync.Mutex
	// establish connection
	url := "ws://localhost:8089/"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("Dial error: %v", err)
	}

	defer func() {
		fmt.Println("\rGuest: closing guest conn")
		conn.Close()
	}()

	writeConnWithLock := func(cmd GameCommand) {
		mu.Lock()
		defer mu.Unlock()
		conn.WriteJSON(cmd)
	}

	// listen to command
	go func() {
		for cmd := range c.cmdCh {
			fmt.Println("\rGuest: sending command")
			writeConnWithLock(cmd)
		}
	}()

	// listen to quit
	go func() {
		for quit := range c.quitCh {
			if quit {
				cmd := GameCommand{Quit: true}
				writeConnWithLock(cmd)
				fmt.Println("\rGuest: sent quit to host")
			}
		}
	}()

	// init game
	var g Game

	// listen to server's game sync
onlineGuestClientLoop:
	for {
		if err := conn.ReadJSON(&g); err != nil {
			// show error when websocket is closed unexpectedly
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				return fmt.Errorf("\rWebSocket Error: %v, %T", err, err)
			}
			break onlineGuestClientLoop
		}

		c.gameCh <- g

		if g.State == Quit {
			break onlineGuestClientLoop
		}
	}

	return nil
}
