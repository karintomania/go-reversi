package main

import (
	"fmt"
	"log"

	"github.com/pkg/term"
)

const (
	BlackString     = "○"
	WhiteString     = "●"
	NothingString   = "_"
	LeftWallString  = "|"
	RightWallString = "|"
	CursorString    = "*"
	Spacer          = "    "
)

type Renderer interface {
	Render(g *Game, p Position)
	Close()
}

type Display struct {
	tm *term.Term
}

func NewDisplay() Display {
	tm, _ := term.Open("/dev/tty")
	err := term.RawMode(tm)
	if err != nil {
		log.Fatal(err)
	}

	d := Display{tm}

	return d
}

func (d *Display) Read(out chan<- string) {
	readBytes := make([]byte, 1)
	_, err := d.tm.Read(readBytes)
	if err != nil {
		log.Fatal(err)
	}

	var char string
	if readBytes[0] == 3 {
		char = "c"

	} else {
		char = string(readBytes[0])
	}

	out <- char
}

func (d *Display) Close() {
	d.tm.Restore()
	d.tm.Close()
}

func (d *Display) Render(g *Game, p Position) {

	// print message
	print("")
	print(fmt.Sprintf(" %s", g.GetInfo().Player1Info))
	print(fmt.Sprintf(" %s", g.GetInfo().Player2Info))
	print("")

	b := g.Board
	state := g.State
	n := b.N

	for y := 0; y < n; y++ {
		rowStr := RightWallString
		for x := 0; x < n; x++ {
			idx := b.Lines[LineId(y)]
			s := idx.GetLocalState(x)
			if y == p.Y && x == p.X { // on focus
				rowStr += getFocusedCellContent(s)
			} else {
				rowStr += getCellContent(s)
			}
		}
		rowStr += LeftWallString
		printWithSpacer(rowStr)
	}

	print("")

	// print message
	print(fmt.Sprintf("[Message] %s", g.Message))

	// print key bindings
	print("")

	switch state {
	case Quit, WaitingConnection:
		print("[Keys] Quit: c")
	case Finished:
		print("[Keys] Play Again: r | Quit: c")
	default:
		print("[Keys] ←↓↑→: a,s,w,d | Place: <space> | Quit: c")
	}

	// move curosr up
	fmt.Printf("\033[%dA\r", n+8)
}

func printWithSpacer(s string) {
	fmt.Printf("\r\033[K%s%s", Spacer, s)
	fmt.Print("\n")
}

func print(s string) {
	fmt.Printf("\r\033[K%s", s)
	fmt.Print("\n")
}

func getFocusedCellContent(s State) string {
	if s == HasNothing {
		return fmt.Sprintf(" %s", CursorString)
	} else if s == HasBlack {
		return fmt.Sprintf("|%s", BlackString)
	} else {
		return fmt.Sprintf("|%s", WhiteString)
	}
}

func getCellContent(s State) string {
	if s == HasNothing {
		return fmt.Sprintf(" %s", NothingString)
	} else if s == HasBlack {
		return fmt.Sprintf(" %s", BlackString)
	} else {
		return fmt.Sprintf(" %s", WhiteString)
	}
}
