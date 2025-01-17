package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var _ = time.Second //TODO: debugging

// host a game server
type OnlineHostConnection struct {
	gameCh       chan Game
	cmdCh        chan<- GameCommand
	quitCh       chan<- bool
	Port         int
	conn         *websocket.Conn
	server       *http.Server
	isConnActive bool
}

func NewOnlineHostConnection(
	gameCh chan Game,
	cmdCh chan<- GameCommand,
	quitCh chan<- bool,
	port int,
) OnlineHostConnection {
	conn := OnlineHostConnection{
		gameCh:       gameCh,
		cmdCh:        cmdCh,
		quitCh:       quitCh,
		Port:         port,
		isConnActive: false,
	}

	return conn
}

func (c *OnlineHostConnection) Run() error {
	// channel to check if connection is made
	connectedCh := make(chan bool)

	go func() {
		c.waitUntilConnIsReady(connectedCh)
	}()

	handler := func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Handler started")

		// establish websocket connection
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			// TODO: handle error properly
			logger.Error("Online host connection error", slog.Any("err", err))
		}

		c.conn = conn

		// Receive command from guest
		go c.handleReceive()

		connectedCh <- true

		// Send game info to guest
		c.handleSend()

	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	c.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Port),
		Handler: mux,
	}

	if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("Error starting server: %w", err)
	}

	return nil
}

func (c *OnlineHostConnection) waitUntilConnIsReady(connectedCh chan bool) {
	// consume game chan until connection is made
	// push back the last game to send via conn
	var lastGameSent Game

	for !c.isConnActive {
		select {
		case c.isConnActive = <-connectedCh:
			c.gameCh <- lastGameSent
			close(connectedCh)
		case g := <-c.gameCh:
			lastGameSent = g
			if g.State == Quit {
				c.isConnActive = false
			}
		}
	}

	logger.Debug("Host conn established")
}

func (c *OnlineHostConnection) handleReceive() {
	for {
		if !c.isConnActive {
			// quit listening if the conn is not active
			break
		}

		cmd := GameCommand{}
		if err := c.conn.ReadJSON(&cmd); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				logger.Error("Host WebSocket Error", slog.Any("err", err))
				break
			}
		}

		logger.Debug("Command received", slog.Any("cmd", cmd))

		if cmd.Quit {
			go func() { c.quitCh <- true }()
			logger.Debug("Quit sent")
		} else {
			go func() { c.cmdCh <- cmd }()
		}
	}
}

func (c *OnlineHostConnection) handleSend() {
	var mu sync.Mutex

	writeWithMutex := func(g Game) error {
		mu.Lock()
		defer mu.Unlock()
		return c.conn.WriteJSON(g)
	}

	for g := range c.gameCh {
		if !c.isConnActive {
			// if connection is not active, discard game
			continue
		}

		logger.Debug("Game received", slog.String("g", g.State.String()))
		if err := writeWithMutex(g); err != nil {
			logger.Error("Error on write", slog.Any("err", err))
			c.isConnActive = false
		}

		if g.State == Quit {
			c.closeWebsocket()
		}
	}
}

func (c *OnlineHostConnection) closeWebsocket() {
	if c.isConnActive {
		c.isConnActive = false

		if err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
			logger.Error("Error sending close message: %v", slog.Any("err", err))
		}

		c.conn.Close()
		logger.Debug("Closed Host websocket")
	}
}

func (c *OnlineHostConnection) Close() error {
	c.closeWebsocket()

	time.Sleep(500 * time.Millisecond)
	// gracefully shutdown the server
	err := c.server.Shutdown(context.TODO())
	if err != nil {
		return fmt.Errorf("Failed to shutdown server: %w", err)
	}

	logger.Debug("Closed host connection")

	return nil
}
