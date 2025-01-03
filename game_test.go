package main

import (
	"testing"
)

func TestQuitGame(t *testing.T) {
	var b Board

	b.init(4)

	g := NewGame(&b, Human, Human)

	ch := make(chan string)

	testStates := []GameState{Playing, Finished}

	for _, currentState := range testStates {
		g.State = currentState

		// quit
		go func() {
			ch <- "c"
		}()
		g.Progress(ch)

		if g.State != Quit {
			t.Errorf("Can't quit from %s, got %s", currentState, g.State)
		}
	}
}

func TestGameMovingPosition(t *testing.T) {
	var b Board

	b.init(4)

	g := NewGame(&b, Human, Human)

	ch := make(chan string)

	// move to playing
	g.Progress(ch)

	// go left
	go func() {
		ch <- "d"
	}()
	g.Progress(ch)

	if g.Board.Position.X != 1 || g.Board.Position.Y != 0 {
		t.Errorf("got %v, but want %v", g.Board.Position, Position{1, 0})
	}

	// go right
	go func() {
		ch <- "a"
	}()
	g.Progress(ch)

	if g.Board.Position.X != 0 || g.Board.Position.Y != 0 {
		t.Errorf("got %v, but want %v", g.Board.Position, Position{0, 0})
	}

	// go down
	go func() {
		ch <- "s"
	}()
	g.Progress(ch)

	if g.Board.Position.X != 0 || g.Board.Position.Y != 1 {
		t.Errorf("got %v, but want %v", g.Board.Position, Position{0, 1})
	}

	// go up
	go func() {
		ch <- "w"
	}()
	g.Progress(ch)

	if g.Board.Position.X != 0 || g.Board.Position.Y != 0 {
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
	// move to playing
	g.Progress(ch)

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
			g.Message,
		)
	}

	// pass White
	g.Progress(ch)

	if g.passCount != 1 || b.Turn != Black {
		t.Errorf("want 1, got %d, %v", g.passCount, b.Turn)
	}
	if g.Message != "Skipped ●" {
		t.Errorf("want 'Skipped ●', got %s", g.Message)
	}

	// pass Black
	g.Progress(ch)

	if g.passCount != 2 || b.Turn != White {
		t.Errorf("want 2, got %d, %v", g.passCount, b.Turn)
	}

	if g.Message != "Skipped ○" {
		t.Errorf("want 'Skipped ○', got %s", g.Message)
	}

	// Finish game
	g.Progress(ch)

	if g.State != Finished {
		t.Errorf("want Finished, got %v", g.State)
	}

	wantMsg := "Black 4, White 12, Player 2 won"
	if g.Message != wantMsg {
		t.Errorf("want '%s', got %s", wantMsg, g.Message)
	}
}

func TestRetry(t *testing.T) {
	var b Board

	b.init(4)

	b.FromStringCells(
		[][]string{
			{"b", "b", "b", "b"},
			{"b", "b", "b", "b"},
			{"b", "b", "b", "b"},
			{"b", "b", "b", "b"},
		},
	)

	g := NewGame(&b, Human, AI)

	ch := make(chan string)

	// move to playing
	g.Progress(ch)

	// pass black
	g.Progress(ch)

	// pass white
	g.Progress(ch)

	// finish game
	g.Progress(ch)

	if g.passCount != 2 || g.State != Finished {
		t.Errorf("needs to be finished, got %d, %v", g.passCount, g.State)
	}

	// press retry
	go func() {
		ch <- "r"
	}()

	g.Progress(ch)

	if g.passCount != 0 || g.State != Initialized {
		t.Errorf("needs to be restarted, got %d, %s", g.passCount, g.State)
	}

	if g.Player1.Type != Human ||
		g.Player1.Colour != White ||
		g.Player2.Type != AI ||
		g.Player2.Colour != Black {
		t.Errorf(
			"Player Type needs to be swapped after replay. %v, %v",
			g.Player2.Type,
			g.Player1.Type,
		)
	}
}

func TestAIPlayer(t *testing.T) {
	var b Board

	b.init(4)

	b.FromStringCells(
		[][]string{
			{"n", "w", "b", "w"},
			{"n", "b", "w", "w"},
			{"w", "w", "w", "w"},
			{"w", "w", "w", "w"},
		},
	)

	g := NewGame(&b, AI, AI)

	ch := make(chan string)
	// move to playing
	g.Progress(ch)

	// place black
	g.Progress(ch)

	if b.Cells[0][0] != HasBlack ||
		b.Cells[0][1] != HasBlack ||
		b.Turn != White {
		t.Errorf("got %v, %v, %v",
			b.Cells[0][0],
			b.Cells[0][1],
			b.Turn,
		)
	}

	// place white
	g.Progress(ch)

	if b.Cells[1][0] != HasWhite ||
		b.Cells[1][1] != HasWhite ||
		b.Turn != Black {
		t.Errorf("got %v, %v, %v",
			b.Cells[1][0],
			b.Cells[1][1],
			b.Turn,
		)
	}
}
