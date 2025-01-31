package main

import (
	"log/slog"
	"sync"
	"time"
)

var _ = time.Second //TODO: debugging

type LocalClient struct {
	gameCh     <-chan Game
	cmdCh      chan<- GameCommand
	quitCh     chan<- bool
	closeCliCh chan<- bool
	inputCh    <-chan string
	PlayerId   PlayerId
	d          Renderer
	p          *Position
}

func NewLocalClient(
	gameCh <-chan Game,
	cmdCh chan<- GameCommand,
	quitCh chan<- bool,
	inputCh <-chan string,
	closeCliCh chan<- bool,
	PlayerId PlayerId,
	d Renderer,
) LocalClient {
	cli := LocalClient{
		gameCh:     gameCh,
		cmdCh:      cmdCh,
		quitCh:     quitCh,
		inputCh:    inputCh,
		closeCliCh: closeCliCh,
		PlayerId:   PlayerId,
		d:          d,
		p:          &Position{},
	}

	return cli
}

func (c *LocalClient) Run() {
	// init game
	var g Game

	go func() {
	localClientInputLoop:
		for char := range c.inputCh {
			switch char {
			// move position
			case "h", "a": // ←
				c.p.addX(-1, g.Board.N)
				c.d.Render(&g, *c.p)
				continue localClientInputLoop
			case "l", "d": // →
				c.p.addX(1, g.Board.N)
				c.d.Render(&g, *c.p)
				continue localClientInputLoop
			case "j", "s": // ↓
				c.p.addY(1, g.Board.N)
				c.d.Render(&g, *c.p)
				continue localClientInputLoop
			case "k", "w": // ↑
				c.p.addY(-1, g.Board.N)
				c.d.Render(&g, *c.p)
				continue localClientInputLoop
			case "c": // quit
				go func() { c.quitCh <- true }()
				c.closeCliCh <- true
				logger.Info("Program finished.")
				break localClientInputLoop
			}

			if g.IsMyTurn(c.PlayerId) {
				switch char {
				case " ": // place
					cmd := GameCommand{CommandType: CommandPlace, Position: *c.p}
					go func() { c.cmdCh <- cmd }()
				}
				continue localClientInputLoop
			}

			if g.State == Finished {
				switch char {
				case "r":
					cmd := GameCommand{CommandType: CommandReplay}
					go func() { c.cmdCh <- cmd }()
				}
				continue localClientInputLoop
			}
		}
	}()

	// localClientLoop:
	for g = range c.gameCh {
		logger.Debug(
			"Game received",
			slog.String("PlayerId", c.PlayerId.String()),
			slog.String("g", g.State.String()),
		)

		if g.State == WaitingConnection &&
			g.GetPlayer(c.PlayerId).Ready == false {
			go func() {
				c.cmdCh <- GameCommand{CommandType: CommandConnectionCheck}
			}()
		}

		c.d.Render(&g, *c.p)
	}
}

type LocalMultiClient struct {
	gameCh1 <-chan Game
	cmdCh1  chan<- GameCommand
	quitCh1 chan<- bool
	gameCh2 <-chan Game
	cmdCh2  chan<- GameCommand
	quitCh2 chan<- bool
	inputCh <-chan string
	d       Renderer
	p       *Position
}

func NewLocalMultiClient(
	gameCh1 <-chan Game,
	cmdCh1 chan<- GameCommand,
	quitCh1 chan<- bool,
	gameCh2 <-chan Game,
	cmdCh2 chan<- GameCommand,
	quitCh2 chan<- bool,
	inputCh <-chan string,
	d Renderer,
) LocalMultiClient {
	cli := LocalMultiClient{
		gameCh1: gameCh1,
		cmdCh1:  cmdCh1,
		quitCh1: quitCh1,
		gameCh2: gameCh2,
		cmdCh2:  cmdCh2,
		quitCh2: quitCh2,
		inputCh: inputCh,
		d:       d,
		p:       &Position{},
	}

	return cli
}

func (c *LocalMultiClient) Run() {
	// init game
	var g Game

	input := make(chan string)
	defer close(input)

	go func() {
	localMultiClientLoop:
		for g = range c.gameCh1 {
			// discard gameCh2
			_ = <-c.gameCh2

			if g.State == WaitingConnection {
				c.cmdCh1 <- GameCommand{CommandType: CommandConnectionCheck}
				c.cmdCh2 <- GameCommand{CommandType: CommandConnectionCheck}
			}

			if g.State == Quit {
				break localMultiClientLoop
			}

			c.d.Render(&g, *c.p)
		}
	}()

localMultiClientInputLoop:
	for char := range c.inputCh {
		if g.State == Player1Turn || g.State == Player2Turn {
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
				if g.State == Player1Turn {
					go func() { c.cmdCh1 <- cmd }()
				} else {
					go func() { c.cmdCh2 <- cmd }()
				}
			}
		}

		if g.State == Finished {
			switch char {
			case "r":
				cmd := GameCommand{CommandType: CommandReplay}
				go func() { c.cmdCh1 <- cmd }()
			}
		}

		if char == "c" {
			c.quitCh1 <- true
			break localMultiClientInputLoop
		}
	}
}

type AiClient struct {
	gameCh   chan Game
	cmdCh    chan GameCommand
	quitCh   chan bool
	PlayerId PlayerId
	p        *AiPlayer
}

func NewAiClient(
	n int,
	gameCh chan Game,
	cmdCh chan GameCommand,
	quitCh chan bool,
	id PlayerId,
) AiClient {

	return AiClient{
		gameCh:   gameCh,
		cmdCh:    cmdCh,
		quitCh:   quitCh,
		PlayerId: id,
		p:        NewAiPlayer(n),
	}
}

func (c *AiClient) Run() {
AiClientLoop:
	for g := range c.gameCh {
		if g.IsMyTurn(c.PlayerId) {
			cmd := c.placeWithMinimumLength(&g)
			c.cmdCh <- cmd
		}

		if g.State == WaitingConnection {
			c.cmdCh <- GameCommand{CommandType: CommandConnectionCheck}
		}

		if g.State == Quit {
			break AiClientLoop
		}
	}
}

// AI's turn will take max(minimum length, position calculation time)
func (c *AiClient) placeWithMinimumLength(g *Game) GameCommand {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		time.Sleep(700 * time.Millisecond)
		wg.Done()
	}()

	p := c.p.getPosition(g.Board)

	cmd := GameCommand{CommandType: CommandPlace, Position: p}

	wg.Wait()

	return cmd
}
