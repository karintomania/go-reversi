package main

import (
	"fmt"
	"log"

	"github.com/pkg/term"
)

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

func (d *Display) Read(out chan string) {
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
func (d *Display) Rendor(b *Board, state GameState, message string) {
	p := b.Position

	n := len(b.Cells)

	for y := 0; y < n; y++ {
		fmt.Print("\r[")
		for x := 0; x < n; x++ {
			s := b.Cells[y][x]
			if y == p.Y && x == p.X {
				fmt.Print(getFocusedCellContent(s))
			} else {
				fmt.Print(getCellContent(s))
			}
		}
		fmt.Print(" ]\n")
	}

	// print turn if playin
	if state == Playing {
		fmt.Printf("\n\rNext: %s", b.Turn)
	} else {
		fmt.Print("\n\r\033[K")
	}

	// print message
	fmt.Printf("\n\r\033[K%s\n", message)

	// print key bindings
	if state == Playing {
		fmt.Printf("\r%s", "←↓↑→: a,s,w,d | <space> place | Quit: c")
	} else {
		fmt.Printf("\r\033[K%s", "Replay: r | Quit: c")
	}

	// move curosr up
	fmt.Printf("\033[%dA\r", n+3)
}

func getFocusedCellContent(s State) string {
	if s == HasNothing {
		return " ■"
	} else if s == HasBlack {
		return fmt.Sprintf("|%s", BlackString)
	} else {
		return fmt.Sprintf("|%s", WhiteString)
	}
}

func getCellContent(s State) string {
	if s == HasNothing {
		return " □"
	} else if s == HasBlack {
		return fmt.Sprintf(" %s", BlackString)
	} else {
		return fmt.Sprintf(" %s", WhiteString)
	}
}
