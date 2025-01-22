package main

import (
	"fmt"
	"log/slog"
)

type GameState int

const (
	Initialized GameState = iota
	WaitingConnection
	Player1Turn
	Player2Turn
	Finished
	Quit
)

const (
	messageWaiting   string = "⏳  Waiting for another player..."
	messageGameStart string = "💫  Game Start!"
	messageTurn      string = "%s  %s's turn"
	messageSkipped   string = "🚨  %s  %s is skipped"
	messageWin       string = "Black %d, White %d, %s won ✨"
	messageDraw      string = "Black %d, White %d, Draw 👏"
	messageQuit      string = "%s left the game 🚪"
)

func (gs GameState) String() string {
	switch gs {
	case Initialized:
		return "Initialized"
	case WaitingConnection:
		return "WaitingConnection"
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
	Board   *Board
	State   GameState
	Player1 Player
	Player2 Player
	Info    GameInfo
	Message string
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
		Player{name1, false, type1, Black},
		Player{name2, false, type2, White},
		GameInfo{},
		"",
	}
}

func (g *Game) Start() (chan GameCommand, chan GameCommand, chan Game, chan Game, chan bool, chan bool) {
	player1Cmd := make(chan GameCommand)
	player2Cmd := make(chan GameCommand)

	player1Game := make(chan Game)
	player2Game := make(chan Game)

	player1Quit := make(chan bool)
	player2Quit := make(chan bool)

	// broadcast game status
	broadcast := func() {
		player1Game <- *g
		player2Game <- *g
	}

	// listen to quit chans
	go func() {
		quit := false
		quittingPlayer := g.Player1.Name

		select {
		case quit = <-player1Quit:
			logger.Debug("Quit received", slog.String("id", Player1Id.String()))
		case quit = <-player2Quit:
			quittingPlayer = g.Player2.Name
			logger.Debug("Quit received", slog.String("id", Player2Id.String()))
		}

		if quit {
			g.State = Quit
			g.Message = fmt.Sprintf(messageQuit, quittingPlayer)

			broadcast()
			logger.Debug("Quit is sent")
		}
	}()

	go func() {
	gameLoop:
		for {
			logger.Debug("State", slog.String("state", g.State.String()))
			switch g.State {
			case Initialized:
				g.Message = messageWaiting
				g.State = WaitingConnection

			case WaitingConnection:
				// make sure both clients are connected
				select {
				case cmd := <-player1Cmd:
					if cmd.CommandType == CommandConnectionCheck {
						g.Player1.Ready = true
					}
				case cmd := <-player2Cmd:
					if cmd.CommandType == CommandConnectionCheck {
						g.Player2.Ready = true
					}
				}

				if g.Player1.Ready && g.Player2.Ready {
					g.Message = messageGameStart
					g.updateTurnFromBoard()
				}

			case Player1Turn, Player2Turn:
				// waiting for players' input
				var cmd GameCommand
				if g.State == Player1Turn {
					cmd = <-player1Cmd
				} else {
					cmd = <-player2Cmd
				}

				switch cmd.CommandType {
				// place
				case CommandPlace:
					g.place(cmd.Position)
				}

			case Finished:
				// wait for input
				var cmd GameCommand
				select {
				case cmd = <-player1Cmd:
				case cmd = <-player2Cmd:
				}

				switch cmd.CommandType {
				case CommandReplay:
					g.replay()
					g.updateTurnFromBoard()
				}

			case Quit:
				break gameLoop
			}

			logger.Debug("Broadcast state", slog.String("state", g.State.String()))
			go broadcast()
		}
	}()

	return player1Cmd, player2Cmd, player1Game, player2Game, player1Quit, player2Quit
}

func (g *Game) place(p Position) {
	b := g.Board

	err := b.Place(p)
	if err != nil {
		g.Message = fmt.Sprintf("%s", err)
		return
	}

	// deal with pass
	passedCount := 0
	for !b.HasLegalCells() && passedCount <= 2 {
		g.pass()

		passedCount++
	}

	if passedCount >= 2 {
		g.finish()
		return
	}

	g.updateTurnFromBoard()

	// show skip message
	if passedCount > 0 {
		skipped := g.GetAnotherPlayer()
		g.Message = fmt.Sprintf(messageSkipped, skipped.Colour, skipped.Name)
	} else {
		playing := g.GetCurrentPlayer()
		g.Message = fmt.Sprintf(messageTurn, playing.Colour, playing.Name)
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
}

func (g *Game) finish() {
	g.Message = g.generateResultMessage()

	g.State = Finished
}

func (g *Game) generateResultMessage() string {
	totalB, totalW := g.Board.Count()

	var playerB, playerW Player
	if g.Player1.Colour == Black {
		playerB = g.Player1
		playerW = g.Player2
	} else {
		playerB = g.Player2
		playerW = g.Player1
	}

	var m string

	if totalB > totalW {
		m = fmt.Sprintf(messageWin, totalB, totalW, playerB.Name)
	} else if totalB < totalW {
		m = fmt.Sprintf(messageWin, totalB, totalW, playerW.Name)
	} else {
		m = fmt.Sprintf(messageDraw, totalB, totalW)
	}

	return m
}

func (g *Game) pass() {
	g.Board.Pass()
}

func (g *Game) GetPlayer(id PlayerId) Player {
	if id == Player1Id {
		return g.Player1
	} else {
		return g.Player2
	}
}

func (g *Game) IsMyTurn(id PlayerId) bool {
	if id == Player1Id {
		return g.State == Player1Turn
	} else {
		return g.State == Player2Turn
	}
}

func (g *Game) GetCurrentPlayer() *Player {
	if g.State == Player1Turn {
		return &g.Player1
	} else {
		return &g.Player2
	}
}

func (g *Game) GetAnotherPlayer() *Player {
	if g.State == Player1Turn {
		return &g.Player2
	} else {
		return &g.Player1
	}
}

type Player struct {
	Name   string
	Ready  bool
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

type GameInfo struct {
	Player1Info string
	Player2Info string
}

func (g *Game) GetInfo() *GameInfo {
	p1Name := fmt.Sprintf("%-15s", g.Player1.Name)
	p2Name := fmt.Sprintf("%-15s", g.Player2.Name)

	totalB, totalW := g.Board.Count()

	var p1Colour, p2Colour string

	if g.Player1.Colour == Black {
		p1Colour = fmt.Sprintf("%s x%d", BlackString, totalB)
		p2Colour = fmt.Sprintf("%s x%d", WhiteString, totalW)
	} else {
		p1Colour = fmt.Sprintf("%s x%d", WhiteString, totalW)
		p2Colour = fmt.Sprintf("%s x%d", BlackString, totalB)
	}

	p1 := fmt.Sprintf("%s %s", p1Name, p1Colour)
	p2 := fmt.Sprintf("%s %s", p2Name, p2Colour)

	if g.State == Player1Turn {
		p1 += " *"
	}

	if g.State == Player2Turn {
		p2 += " *"
	}

	return &GameInfo{p1, p2}
}

type PlayerId int

func (id PlayerId) String() string {
	switch id {
	case Player1Id:
		return "Player 1"
	case Player2Id:
		return "Player 2"
	default:
		return "Undefined PlayerId"
	}
}

const (
	Player1Id PlayerId = iota
	Player2Id
)

type CommandType int

const (
	CommandPlace CommandType = iota
	CommandConnectionCheck
	CommandReplay
)

func (c CommandType) String() string {
	switch c {
	case CommandPlace:
		return "CommandPlace"
	case CommandConnectionCheck:
		return "CommandConnectionCheck"
	case CommandReplay:
		return "CommandReplay"
	default:
		return "Unknown"
	}
}

type GameCommand struct {
	CommandType CommandType
	Position    Position
	Quit        bool
}
