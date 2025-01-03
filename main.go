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

	input := make(chan string)
	defer close(input)

	go func() {
		for {
			d.Read(input)
		}
	}()

	g := NewGame(&b, Human, Human)

	d.Rendor(g.Board, g.State)

	for g.State != Quit {
		g.Progress(input)
		d.Rendor(g.Board, g.State)
	}

}
