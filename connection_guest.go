package main

import (
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var _ = time.Second //TODO: debugging

// guest client sends command to host
type OnlineGuestConnection struct {
	gameCh       chan<- Game
	cmdCh        <-chan GameCommand
	quitCh       <-chan bool
	closeConnCh  chan<- bool
	PlayerId     PlayerId
	Url          string
	conn         *websocket.Conn
	isConnActive bool
	mu           *sync.Mutex
}

func NewOnlineGuestConnection(id PlayerId, url string, port int) (*OnlineGuestConnection, chan Game, chan GameCommand, chan bool) {
	gameCh := make(chan Game)
	cmdCh := make(chan GameCommand)
	quitCh := make(chan bool)

	wsUrl := convertToWebSocketURL(url, port)

	var mu sync.Mutex

	conn := &OnlineGuestConnection{
		gameCh:       gameCh,
		cmdCh:        cmdCh,
		quitCh:       quitCh,
		PlayerId:     id,
		Url:          wsUrl,
		isConnActive: false,
		mu:           &mu,
	}

	return conn, gameCh, cmdCh, quitCh
}

func (c *OnlineGuestConnection) Run() error {
	logger.Debug("Guest started")

	// establish connection
	conn, _, err := websocket.DefaultDialer.Dial(c.Url, nil)
	if err != nil {
		return fmt.Errorf("Dial error: %v", err)
	}

	c.conn = conn
	c.isConnActive = true

	// listen to command
	go func() {
		for cmd := range c.cmdCh {
			logger.Debug("Send cmd to host", slog.Any("cmd", cmd))
			c.writeCmd(cmd)
		}
	}()

	// listen to quit
	go func() {
		for quit := range c.quitCh {
			if quit {
				logger.Debug("Sent quit")
				cmd := GameCommand{Quit: true}
				c.writeCmd(cmd)
			}
		}
	}()

	// init game
	var g Game

	// listen to server's game sync
	for {
		if err := c.conn.ReadJSON(&g); err != nil {
			// show error when websocket is closed unexpectedly
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				return fmt.Errorf("WebSocket Error: %v, %T", err, err)
			}
		}

		logger.Debug("Received game", slog.String("g", g.State.String()))

		c.gameCh <- g
	}
}

func (c *OnlineGuestConnection) Close() {
	logger.Debug("Close guest")

	if c.isConnActive {
		// send quit command
		cmd := GameCommand{Quit: true}
		c.writeCmd(cmd)

		c.conn.Close()

		c.isConnActive = false
	}
}

func (c *OnlineGuestConnection) writeCmd(cmd GameCommand) {
	if !c.isConnActive {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.WriteJSON(cmd)
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
