package main

import "fmt"

type GameState int

const (
	Initialized GameState = iota
	Playing
	Finished
	Quit
)

func (gs GameState) String() string {
	switch gs {
	case Initialized:
		return "Initialized"
	case Playing:
		return "Playing"
	case Finished:
		return "Finished"
	case Quit:
		return "Quit"
	default:
		return "Not Defined"
	}
}

type Game struct {
	Board     *Board
	State     GameState
	Player1   Player
	Player2   Player
	Message   string
	passCount int
}

func NewGame(b *Board, type1, type2 PlayerType) Game {
	return Game{
		b,
		Initialized,
		Player{"Player 1", type1, Black},
		Player{"Player 2", type2, White},
		"",
		0}
}

func (g *Game) Progress(in <-chan string) {
	b := g.Board

	if g.State == Initialized {
		g.Message = g.getPlayerTypesMessage()
		g.State = Playing
		return
	}

	if g.State == Playing {
		if g.passCount >= 2 {
			g.finish()
			return
		}

		if !b.HasPlayableCells() {
			g.pass()
			return
		}

		// AI player
		if g.getCurrentPlayer().Type == AI {
			b.PlaceByAi()
			return
		}

		char := <-in

		switch char {
		// move position
		case "h", "a": // ←
			b.MovePositionX(-1)
		case "l", "d": // →
			b.MovePositionX(1)
		case "j", "s": // ↓
			b.MovePositionY(1)
		case "k", "w": // ↑
			b.MovePositionY(-1)

		// place
		case " ":
			err := b.Place()
			if err != nil {
				g.Message = fmt.Sprintf("%s", err)
			}
		// quit
		case "c":
			g.State = Quit
		}
	}

	if g.State == Finished {
		char := <-in

		switch char {
		case "r":
			g.replay()
		case "c": // quit
			g.State = Quit
			return
		}
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
	m := fmt.Sprintf("Black: %d, White %d,", totalB, totalW)

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
	player1Type, player2Type := "", ""

	if g.Player1.Type == AI {
		player1Type = " (AI)"
	}
	if g.Player2.Type == AI {
		player2Type = " (AI)"
	}

	return fmt.Sprintf(
		"%s%s: %s, %s%s: %s",
		g.Player1.Name,
		player1Type,
		g.Player1.Colour,
		g.Player2.Name,
		player2Type,
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
