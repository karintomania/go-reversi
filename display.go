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

	char := string(readBytes[0])

	fmt.Print(char)

	out <- char
}

func (d *Display) Close() {
	d.tm.Restore()
	d.tm.Close()
}
func (d *Display) Rendor(b *Board, state GameState) {
	p := b.Position

	n := len(b.Cells)

	for y := 0; y < n; y++ {
		fmt.Print("\r[")
		for x := 0; x < n; x++ {
			s := b.Cells[y][x]
			if y == p.Y && x == p.X {
				if s == HasNothing {
					fmt.Print(" ■")
				} else if s == HasBlack {
					fmt.Printf("|%s", BlackString)
				} else if s == HasWhite {
					fmt.Printf("|%s", WhiteString)
				}
			} else {
				if s == HasNothing {
					fmt.Print(" □")
				} else if s == HasBlack {
					fmt.Printf(" %s", BlackString)
				} else if s == HasWhite {
					fmt.Printf(" %s", WhiteString)
				}
			}
		}
		fmt.Print(" ]\n")
	}

	if state == Playing {
		fmt.Printf("\n\rNext: %s", b.Turn)
	} else {
		fmt.Print("\n\r\033[K")
	}
	fmt.Printf("\n\r\033[K%s\n", b.Message)

	if state == Playing {
		fmt.Printf("\r%s", "←↓↑→: a,s,w,d | <space> place | Quit: c")
	} else {
		fmt.Printf("\r\033[K%s", "Replay: r | Quit: c")
	}
	fmt.Printf("\033[%dA\r", n+3)
}
