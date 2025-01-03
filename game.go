package main

type GameState int

const (
	Playing GameState = iota
	Finished
	Quit
)

type Game struct {
	Board       *Board
	State       GameState
	PlayerBlack Player
	PlayerWhite Player
	passCount   int
}

func NewGame(b *Board, typeB, typeW PlayerType) Game {
	return Game{b, Playing, Player{typeB}, Player{typeW}, 0}
}

func (g *Game) Progress(in <-chan string) {
	b := g.Board

	if g.State == Playing {
		if g.passCount >= 2 {
			g.finish()
			return
		}

		if !b.HasPlayableCells() {
			g.pass()
			return
		}

		if g.getPlayer().Type == AI {
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
			b.Place()

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
		// quit
		case "c":
			g.State = Quit
			return
		}
	}
}

func (g *Game) replay() {
	g.Board.init(MAX)
	g.passCount = 0
	g.State = Playing
}

func (g *Game) finish() {
	g.Board.Finish()
	g.State = Finished
}

func (g *Game) pass() {
	g.passCount++
	g.Board.Pass()
}

func (g *Game) getPlayer() Player {
	turn := g.Board.Turn
	if turn == Black {
		return g.PlayerBlack
	} else {
		return g.PlayerWhite
	}
}

type Player struct {
	Type PlayerType
}

type PlayerType int

const (
	Human PlayerType = iota
	AI
)
