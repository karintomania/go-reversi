package main

import "fmt"

type GameState int

const (
	Initialized GameState = iota
	Player1Turn
	Player2Turn
	Finished
	Quit
)

func (gs GameState) String() string {
	switch gs {
	case Initialized:
		return "Initialized"
	case Player1Turn:
		return "Player1Turn"
	case Player2Turn:
		return "Player2Turn"
	case Finished:
		return "Finished"
	case Quit:
		return "Quit"
	default:
		return "Not Defined"
	}
}

type Game struct {
	Board        *Board
	State        GameState
	Player1      Player
	Player2      Player
	Message      string
	DebugMessage string
	passCount    int
}

func NewGame(b *Board, type1, type2 PlayerType) Game {
	name1, name2 := "Player 1", "Player 2"
	if type1 == AI {
		name1 += " (AI)"
	}

	if type2 == AI {
		name2 += " (AI)"
	}

	return Game{
		b,
		Initialized,
		Player{name1, type1, Black},
		Player{name2, type2, White},
		"",
		"",
		0}
}

func (g *Game) Start() (chan GameCommand, chan GameCommand, chan Game, chan Game) {
	player1Cmd := make(chan GameCommand)
	player2Cmd := make(chan GameCommand)
	player1Out := make(chan Game)
	player2Out := make(chan Game)

	b := g.Board

	go func() {
	gameLoop:
		for {
			if g.State == Initialized {
				g.Message = g.getPlayerTypesMessage()
				g.updateTurnFromBoard()
				g.DebugMessage = fmt.Sprintf("init with state: %s", g.State)
				continue gameLoop
			}

			if g.State == Player1Turn || g.State == Player2Turn {
				if g.passCount >= 2 {
					g.finish()
					player1Out <- *g
					player2Out <- *g
					continue gameLoop
				}

				if b.HasPlayableCells() {
					g.updateTurnFromBoard()
					player1Out <- *g
					player2Out <- *g
				} else {
					g.pass()
					continue gameLoop
				}

				var cmd GameCommand
				if g.State == Player1Turn {
					g.DebugMessage = fmt.Sprint("waiting for 1")
					cmd = <-player1Cmd
				} else {
					g.DebugMessage = fmt.Sprint("waiting for 2")
					cmd = <-player2Cmd
				}

				switch cmd.CommandType {
				// place
				case CommandPlace:
					g.Place(cmd.Position)
					// quit
				case CommandQuit:
					g.State = Quit
					player1Out <- *g
					player2Out <- *g
					break gameLoop
				}
			}

			if g.State == Finished {
				cmd := <-player1Cmd

				switch cmd.CommandType {
				case CommandReplay:
					g.replay()
				case CommandQuit:
					g.State = Quit
					break gameLoop
				}
				player1Out <- *g
				player2Out <- *g
			}
		}
	}()

	return player1Cmd, player2Cmd, player1Out, player2Out
}

func (g *Game) Place(p Position) {
	err := g.Board.Place(p)
	if err != nil {
		g.Message = fmt.Sprintf("%s", err)
	}
}

func (g *Game) updateTurnFromBoard() {
	if g.Board.Turn == g.Player1.Colour {
		g.State = Player1Turn
	} else {
		g.State = Player2Turn
	}
}

func (g *Game) replay() {
	// swap player colour
	g.Player1.Colour, g.Player2.Colour = g.Player2.Colour, g.Player1.Colour

	g.Board.init(g.Board.N)
	g.passCount = 0

	g.State = Initialized
}

func (g *Game) finish() {
	g.Message = g.generateResultMessage()

	g.State = Finished
}

func (g *Game) generateResultMessage() string {
	totalB, totalW := g.Board.Finish()
	m := fmt.Sprintf("Black %d, White %d,", totalB, totalW)

	var playerB, playerW Player
	if g.Player1.Colour == Black {
		playerB = g.Player1
		playerW = g.Player2
	} else {
		playerB = g.Player2
		playerW = g.Player1
	}

	if totalB > totalW {
		m = fmt.Sprintf("%s %s won", m, playerB.Name)
	} else if totalB < totalW {
		m = fmt.Sprintf("%s %s won", m, playerW.Name)
	} else {
		m = fmt.Sprintf("%s Draw", m)
	}

	return m
}

func (g *Game) pass() {
	g.Message = fmt.Sprintf("Skipped %s", g.Board.Turn.String())
	g.passCount++
	g.Board.Pass()
}

func (g *Game) getCurrentPlayer() Player {
	turn := g.Board.Turn

	if g.Player1.Colour == turn {
		return g.Player1
	} else {
		return g.Player2
	}
}

func (g *Game) getPlayerTypesMessage() string {

	return fmt.Sprintf(
		"%s: %s, %s: %s",
		g.Player1.Name,
		g.Player1.Colour,
		g.Player2.Name,
		g.Player2.Colour,
	)
}

type Player struct {
	Name   string
	Type   PlayerType
	Colour Turn
}

type PlayerType int

const (
	Human PlayerType = iota
	AI
)

func (p PlayerType) String() string {
	switch p {
	case Human:
		return "Human"
	case AI:
		return "AI"
	default:
		return "Human"
	}
}

func GetPlayerTypeFromString(s string) PlayerType {
	switch s {
	case "Human":
		return Human
	case "AI":
		return AI
	}
	return Human
}

type CommandType int

const (
	CommandPlace CommandType = iota
	CommandReplay
	CommandQuit
)

type GameCommand struct {
	CommandType CommandType
	Position    Position
}
