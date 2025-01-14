package main

import (
	// "context"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var _ = time.Second //TODO: debugging

// host a game server
type OnlineHostConnection struct {
	gameCh chan Game
	cmdCh  chan<- GameCommand
	quitCh chan<- bool
	Port   int
}

func (c *OnlineHostConnection) Run() error {
	// channel to detect if game is quit
	closeConnCh := make(chan bool)

	// channel to check if connection is made
	connectedCh := make(chan bool)

	go func() {
		// consume game chan until connection is made
		connected := false
		var lastGameSent Game
		for !connected {
			select {
			case <-connectedCh:
				connected = true
				c.gameCh <- lastGameSent
				close(connectedCh)
			case g := <-c.gameCh:
				lastGameSent = g
				if g.State == Quit {
					closeConnCh <- true
				}
			}
		}
		logger.Debug("Host conn established")
	}()

	handler := func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Handler started")

		// establish websocket connection
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			// TODO: handle error properly
			logger.Error("Online host connection error", slog.Any("err", err))
			closeConnCh <- true
		}
		connectedCh <- true

		defer func() {
			conn.Close()
			logger.Debug("Closed conn")
		}()

		// Send game info to guest
		go c.handleSend(conn, closeConnCh)

		// Receive command from guest
		c.handleReceive(conn, closeConnCh)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Error starting server", slog.Any("err", err))
		}
	}()

	// wait for closeConn signal
	<-closeConnCh

	// gracefully shutdown the server
	err := server.Shutdown(context.TODO())
	if err != nil {
		return fmt.Errorf("Failed to shutdown server: %v", err)
	}

	return nil
}

func (c *OnlineHostConnection) handleReceive(conn *websocket.Conn, closeConnCh chan bool) {
	for {
		cmd := GameCommand{}

		if err := conn.ReadJSON(&cmd); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				logger.Error("Host WebSocket Error", slog.Any("err", err))
			}
			closeConnCh <- true
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

func (c *OnlineHostConnection) handleSend(conn *websocket.Conn, closeConnCh chan bool) {
	var mu sync.Mutex

	writeWithMutex := func(g Game) error {
		mu.Lock()
		defer mu.Unlock()
		return conn.WriteJSON(g)
	}

	for g := range c.gameCh {
		logger.Debug("Game received", slog.String("g", g.State.String()))
		if err := writeWithMutex(g); err != nil {
			logger.Error("Error on write", slog.Any("err", err))
			closeConnCh <- true
		}

		if g.State == Quit {
			closeConnCh <- true
		}
	}
}

// guest client sends command to host
type OnlineGuestConnection struct {
	gameCh   chan<- Game
	cmdCh    <-chan GameCommand
	quitCh   <-chan bool
	PlayerId PlayerId
	Url      string
}

func NewOnlineGuestConnection(id PlayerId, url string, port int) (OnlineGuestConnection, chan Game, chan GameCommand, chan bool) {
	gameCh := make(chan Game)
	cmdCh := make(chan GameCommand)
	quitCh := make(chan bool)

	wsUrl := convertToWebSocketURL(url, port)

	conn := OnlineGuestConnection{
		gameCh,
		cmdCh,
		quitCh,
		id,
		wsUrl,
	}

	return conn, gameCh, cmdCh, quitCh
}

func (c *OnlineGuestConnection) Run() error {
	logger.Debug("Guest started")

	var mu sync.Mutex
	// establish connection
	conn, _, err := websocket.DefaultDialer.Dial(c.Url, nil)
	if err != nil {
		return fmt.Errorf("Dial error: %v", err)
	}

	defer func() {
		logger.Debug("Close guest")
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
			logger.Debug("Send cmd to host", slog.Any("cmd", cmd))
			writeConnWithLock(cmd)
		}
	}()

	// listen to quit
	go func() {
		for quit := range c.quitCh {
			if quit {
				logger.Debug("Sent quit")
				cmd := GameCommand{Quit: true}
				writeConnWithLock(cmd)
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
				return fmt.Errorf("WebSocket Error: %v, %T", err, err)
			}
			break onlineGuestClientLoop
		}

		logger.Debug("Received game", slog.String("g", g.State.String()))

		c.gameCh <- g

		if g.State == Quit {
			break onlineGuestClientLoop
		}
	}

	return nil
}

// convertToWebSocketURL takes an HTTP/HTTPS URL and converts it to a WebSocket URL
func convertToWebSocketURL(inputURL string, port int) string {
	u, err := url.Parse(inputURL)
	if err != nil {
		logger.Error("Error on parsing URL", slog.Any("err", err))
	}

	// Replace the scheme with ws or wss
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
		// do nothing
	default:
		logger.Error("URL scheme must be http/https/ws/wss")
	}

	if u.Port() == "" {
		u.Host = fmt.Sprintf("%s:%d", u.Hostname(), port)
	}
	return u.String()
}
