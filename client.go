package main

type Client struct {
	stdin     chan string
	gameCh    chan Game
	gameCmdCh chan GameCommand
	d         *Display
	p         *Position
}

func (c *Client) Run() {
	// init game
	g := <-c.gameCh

	c.d.Rendor(&g, *c.p)

clientLoop:
	for {
		select {
		case char := <-c.stdin:
			//

			if g.State == Player1Turn {
				switch char {
				// move position
				case "h", "a": // ←
					c.p.addX(-1, g.Board.N)
				case "l", "d": // →
					c.p.addX(1, g.Board.N)
				case "j", "s": // ↓
					c.p.addY(1, g.Board.N)
				case "k", "w": // ↑
					c.p.addY(-1, g.Board.N)

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
			c.d.Rendor(&g, *c.p)
		case g = <-c.gameCh:
			if g.State == Quit {
				break clientLoop
			}

			c.d.Rendor(&g, *c.p)
		}
	}
}

type AiClient struct {
	stdin     chan string
	gameCh    chan Game
	gameCmdCh chan GameCommand
	d         *Display
	p         *Position
}

func (c *AiClient) Run() {
AiClientLoop:
	for g := range c.gameCh {
		if g.State == Player2Turn {
			p := getAiPosition(g.Board)
			cmd := GameCommand{CommandPlace, p}
			c.gameCmdCh <- cmd
		}

		if g.State == Quit {
			break AiClientLoop
		}
	}
}
