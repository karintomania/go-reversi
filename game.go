package main

type Game struct {
	Board     Board
	State     GameState
	PassCount int
}

func (g *Game) Progress(d *Display) {
	b := g.Board

	if b.GameState == Playing {
		if g.PassCount >= 2 {
			g.State = Finished
			g.PassCount = 0
			return
		}

		if !b.HasPlayableCells() {
			b.Pass()
			g.PassCount++
			return
		}

		if b.Turn == White {
			b.Position = b.GetPcPosition()
			b.Place()
			return
		}

		char := d.Read()

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

		// pass
		case "p":
			b.Pass()

		// quit
		case "c":
			g.State = Quit
			return
		}
	}

	if b.GameState == Finished {
		char := d.Read()

		switch char {
		case "r":
			b.init(MAX)
		// quit
		case "c":
			g.State = Quit
			return
		}
	}

}
