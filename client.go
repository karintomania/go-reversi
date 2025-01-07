package main

import "time"

var _ = time.Second //TODO: debugging

type Client struct {
	gameCh    chan Game
	gameCmdCh chan GameCommand
	PlayerId  PlayerId
	d         *Display
	p         *Position
}

func (c *Client) Run() {
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
					c.d.Rendor(&g, *c.p)
				case "l", "d": // →
					c.p.addX(1, g.Board.N)
					c.d.Rendor(&g, *c.p)
				case "j", "s": // ↓
					c.p.addY(1, g.Board.N)
					c.d.Rendor(&g, *c.p)
				case "k", "w": // ↑
					c.p.addY(-1, g.Board.N)
					c.d.Rendor(&g, *c.p)

				// place
				case " ":
					cmd := GameCommand{CommandPlace, *c.p}
					c.gameCmdCh <- cmd
				// quit
				case "c":
					cmd := GameCommand{CommandType: CommandQuit}
					c.gameCmdCh <- cmd
				}
			}

			if g.State == Finished {
				switch char {
				case "r":
					cmd := GameCommand{CommandType: CommandReplay}
					c.gameCmdCh <- cmd
				case "c": // quit
					cmd := GameCommand{CommandType: CommandQuit}
					c.gameCmdCh <- cmd
				}
			}
		}
	}()

clientLoop:
	for g = range c.gameCh {
		if g.State == WaitingConnection {
			go func() {
				c.gameCmdCh <- GameCommand{CommandType: CommandConnectionCheck}
			}()
		}

		if g.State == Quit {
			break clientLoop
		}

		if c.PlayerId == Player1Id {
			c.d.Rendor(&g, *c.p)
		}
	}
}

type LocalMultiClient struct {
	gameCh1    chan Game
	gameCmdCh1 chan GameCommand
	gameCh2    chan Game
	gameCmdCh2 chan GameCommand
	d          *Display
	p          *Position
}

func (c *LocalMultiClient) Run() {
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
			if g.State == Player1Turn || g.State == Player2Turn {
				switch char {
				// move position
				case "h", "a": // ←
					c.p.addX(-1, g.Board.N)
					c.d.Rendor(&g, *c.p)
				case "l", "d": // →
					c.p.addX(1, g.Board.N)
					c.d.Rendor(&g, *c.p)
				case "j", "s": // ↓
					c.p.addY(1, g.Board.N)
					c.d.Rendor(&g, *c.p)
				case "k", "w": // ↑
					c.p.addY(-1, g.Board.N)
					c.d.Rendor(&g, *c.p)

				// place
				case " ":
					cmd := GameCommand{CommandPlace, *c.p}
					if g.State == Player1Turn {
						c.gameCmdCh1 <- cmd
					} else {
						c.gameCmdCh2 <- cmd
					}
				// quit
				case "c":
					cmd := GameCommand{CommandType: CommandQuit}
					c.gameCmdCh1 <- cmd
				}
			}

			if g.State == Finished {
				switch char {
				case "r":
					cmd := GameCommand{CommandType: CommandReplay}
					c.gameCmdCh1 <- cmd
				case "c": // quit
					cmd := GameCommand{CommandType: CommandQuit}
					c.gameCmdCh1 <- cmd
				}
			}
		}
	}()

localMultiClientLoop:
	for g = range c.gameCh1 {
		// discard gameCh2
		_ = <-c.gameCh2

		if g.State == WaitingConnection {
			c.gameCmdCh1 <- GameCommand{CommandType: CommandConnectionCheck}
			c.gameCmdCh2 <- GameCommand{CommandType: CommandConnectionCheck}
		}

		if g.State == Quit {
			break localMultiClientLoop
		}
		c.d.Rendor(&g, *c.p)

	}
}

type AiClient struct {
	gameCh    chan Game
	gameCmdCh chan GameCommand
	PlayerId  PlayerId
}

func (c *AiClient) Run() {
AiClientLoop:
	for g := range c.gameCh {
		if g.IsMyTurn(c.PlayerId) {
			p := getAiPosition(g.Board)
			cmd := GameCommand{CommandPlace, p}
			// time.Sleep(5 * time.Second)
			c.gameCmdCh <- cmd
		}

		if g.State == WaitingConnection {
			c.gameCmdCh <- GameCommand{CommandType: CommandConnectionCheck}
		}

		if g.State == Quit {
			break AiClientLoop
		}
	}
}
