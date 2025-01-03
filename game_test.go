package main

import "testing"

func TestGameMovingPosition(t *testing.T) {
	var b Board

	b.init(2)

	g := NewGame(&b, Human, Human)

	ch := make(chan string)

	// go left
	go func() {
		ch <- "d"
	}()

	g.Progress(ch)

	if g.Board.Position.X != 1 && g.Board.Position.Y != 0 {
		t.Errorf("got %v, but want %v", g.Board.Position, Position{1, 0})
	}

	// go right
	go func() {
		ch <- "a"
	}()

	g.Progress(ch)

	if g.Board.Position.X != 0 && g.Board.Position.Y != 0 {
		t.Errorf("got %v, but want %v", g.Board.Position, Position{0, 0})
	}

	// go down
	go func() {
		ch <- "s"
	}()

	g.Progress(ch)

	if g.Board.Position.X != 0 && g.Board.Position.Y != 1 {
		t.Errorf("got %v, but want %v", g.Board.Position, Position{0, 1})
	}

	// go up
	go func() {
		ch <- "w"
	}()

	g.Progress(ch)

	if g.Board.Position.X != 0 && g.Board.Position.Y != 0 {
		t.Errorf("got %v, but want %v", g.Board.Position, Position{0, 0})
	}
}

func TestPassAndFinishGame(t *testing.T) {
	var b Board

	b.init(4)

	b.FromStringCells(
		[][]string{
			{"n", "w", "w", "b"},
			{"w", "w", "w", "w"},
			{"w", "w", "w", "w"},
			{"w", "w", "w", "w"},
		},
	)

	g := NewGame(&b, Human, Human)

	ch := make(chan string)

	// place black
	go func() {
		ch <- " "
	}()

	g.Progress(ch)

	if b.Cells[0][0] != HasBlack ||
		b.Cells[0][1] != HasBlack ||
		b.Cells[0][2] != HasBlack ||
		b.Turn != White {
		t.Errorf(
			"Place failed, got %v, %v, %v, %v, \"%v\"",
			b.Cells[0][0],
			b.Cells[0][1],
			b.Cells[0][2],
			b.Turn,
			b.Message,
		)
	}

	// pass White
	g.Progress(ch)

	if g.passCount != 1 || b.Turn != Black {
		t.Errorf("want 1, got %d, %v", g.passCount, b.Turn)
	}
	if b.Message != "Skipped ●" {
		t.Errorf("want 'Skipped ●', got %s", b.Message)
	}

	// pass Black
	g.Progress(ch)

	if g.passCount != 2 || b.Turn != White {
		t.Errorf("want 2, got %d, %v", g.passCount, b.Turn)
	}

	if b.Message != "Skipped ○" {
		t.Errorf("want 'Skipped ○', got %s", b.Message)
	}

	// Finish game
	g.Progress(ch)

	if g.State != Finished {
		t.Errorf("want Finished, got %v", g.State)
	}
}
