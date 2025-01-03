package main

import "fmt"

var _ string = fmt.Sprint("test")

const (
	MAX         = 4
	BlackString = "○"
	WhiteString = "●"
)

func main() {
	var b Board

	b.init(MAX)

	d := NewDisplay()
	defer d.Close()

	d.Rendor(b)

loop:
	for {
		if !b.HasPlayableCells() {
			b.Pass()
			if !b.HasPlayableCells() {
				// finish game if both player doesn't have cell to place
				b.Finish()
			}
			d.Rendor(b)
		}

		if b.GameState == Playing {
			if b.Turn == White {
				b.Position = b.GetPcPosition()
				b.Place()
				d.Rendor(b)
				continue loop
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
				break loop
			}
		}

		if b.GameState == Finished {
			char := d.Read()

			switch char {
			case "r":
				b.init(MAX)
			// quit
			case "c":
				break loop
			}
		}

		d.Rendor(b)

	}

}
