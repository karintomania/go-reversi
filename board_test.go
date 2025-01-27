package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayBoardIsCellTaken(t *testing.T) {
	cases := []struct {
		Input Position
		Want  bool
	}{
		{Position{0, 0}, true},
		{Position{1, 1}, false},
		{Position{2, 2}, false},
	}

	for _, c := range cases {
		want := c.Want
		b := ArrayBoard{}

		b.init(3)

		b.FromStringCells(
			[][]string{
				{"n", "n", "n"},
				{"b", "b", "b"},
				{"w", "w", "w"},
			},
		)

		got := b.isCellTaken(c.Input)

		assert.Equal(t, want, got)
	}
}

func TestArrayBoardGetCellsToFlip(t *testing.T) {
	b := ArrayBoard{}

	b.init(4)

	type input struct {
		CellsStr [][]string
		Position Position
		Turn     Turn
	}

	cases := []struct {
		Name  string
		Input input
		Want  []CellToFlip
	}{
		{
			Name: "Horizontal Right, black",
			Input: input{
				[][]string{
					{"n", "w", "b", "n"},
					{"n", "n", "n", "n"},
					{"n", "n", "n", "n"},
					{"n", "n", "n", "n"},
				},
				Position{0, 0},
				Black,
			},
			Want: []CellToFlip{
				{1, 0, HasBlack},
			},
		},
		{
			Name: "Horizontal Left, black",
			Input: input{
				[][]string{
					{"b", "w", "w", "n"},
					{"n", "n", "n", "n"},
					{"n", "n", "n", "n"},
					{"n", "n", "n", "n"},
				},
				Position{3, 0},
				Black,
			},
			Want: []CellToFlip{
				{2, 0, HasBlack},
				{1, 0, HasBlack},
			},
		},
		{
			Name: "Vertical to bottom, black",
			Input: input{
				[][]string{
					{"n", "n", "n", "n"},
					{"w", "n", "n", "n"},
					{"w", "n", "n", "n"},
					{"b", "n", "n", "n"},
				},
				Position{0, 0},
				Black,
			},
			Want: []CellToFlip{
				{0, 1, HasBlack},
				{0, 2, HasBlack},
			},
		},
		{
			Name: "Vertical to top, black",
			Input: input{
				[][]string{
					{"n", "b", "n", "n"},
					{"n", "w", "n", "n"},
					{"n", "w", "n", "n"},
					{"n", "n", "n", "n"},
				},
				Position{1, 3},
				Black,
			},
			Want: []CellToFlip{
				{1, 2, HasBlack},
				{1, 1, HasBlack},
			},
		},
		{
			Name: "Diagonal to bottom right, black",
			Input: input{
				[][]string{
					{"n", "n", "n", "n"},
					{"n", "w", "n", "n"},
					{"n", "n", "w", "n"},
					{"n", "n", "n", "b"},
				},
				Position{0, 0},
				Black,
			},
			Want: []CellToFlip{
				{1, 1, HasBlack},
				{2, 2, HasBlack},
			},
		},
		{
			Name: "Diagonal to bottom left, white",
			Input: input{
				[][]string{
					{"n", "n", "n", "n"},
					{"n", "n", "b", "b"},
					{"n", "b", "w", "n"},
					{"w", "n", "n", "w"},
				},
				Position{3, 0},
				White,
			},
			Want: []CellToFlip{
				{2, 1, HasWhite},
				{1, 2, HasWhite},
			},
		},
		{
			Name: "Diagonal to top right, white",
			Input: input{
				[][]string{
					{"n", "n", "w", "w"},
					{"n", "b", "b", "n"},
					{"n", "b", "n", "w"},
					{"n", "n", "n", "n"},
				},
				Position{0, 2},
				White,
			},
			Want: []CellToFlip{
				{1, 1, HasWhite},
			},
		},
		{
			Name: "Diagonal to top left, black",
			Input: input{
				[][]string{
					{"w", "n", "w", "w"},
					{"n", "b", "b", "n"},
					{"n", "b", "b", "w"},
					{"n", "n", "n", "n"},
				},
				Position{3, 3},
				White,
			},
			Want: []CellToFlip{
				{2, 2, HasWhite},
				{1, 1, HasWhite},
			},
		},
		{
			Name: "multiple directions 1",
			Input: input{
				[][]string{
					{"n", "w", "w", "b"},
					{"w", "w", "b", "w"},
					{"w", "b", "w", "n"},
					{"b", "n", "n", "b"},
				},
				Position{0, 0},
				Black,
			},
			Want: []CellToFlip{
				{1, 0, HasBlack},
				{2, 0, HasBlack},
				{0, 1, HasBlack},
				{0, 2, HasBlack},
				{1, 1, HasBlack},
				{2, 2, HasBlack},
			},
		},
		{
			Name: "multiple directions 2",
			Input: input{
				[][]string{
					{"w", "w", "w", "b"},
					{"b", "b", "b", "w"},
					{"n", "b", "w", "n"},
					{"w", "w", "n", "b"},
				},
				Position{0, 2},
				White,
			},
			Want: []CellToFlip{
				{1, 2, HasWhite},
				{0, 1, HasWhite},
				{1, 1, HasWhite},
			},
		},
		{
			Name: "Do nothing horizontal",
			Input: input{
				[][]string{
					{"n", "w", "w", "w"},
					{"n", "w", "b", "n"},
					{"n", "b", "w", "n"},
					{"w", "n", "n", "n"},
				},
				Position{0, 0},
				Black,
			},
			Want: []CellToFlip{},
		},
		{
			Name: "Do nothing vertical",
			Input: input{
				[][]string{
					{"n", "w", "b", "w"},
					{"n", "b", "w", "w"},
					{"w", "n", "n", "w"},
					{"w", "w", "w", "n"},
				},
				Position{3, 3},
				Black,
			},
			Want: []CellToFlip{},
		},
		{
			Name: "Do nothing diagnol",
			Input: input{
				[][]string{
					{"b", "n", "n", "n"},
					{"n", "w", "b", "n"},
					{"n", "b", "w", "n"},
					{"w", "n", "n", "n"},
				},
				Position{0, 1},
				White,
			},
			Want: []CellToFlip{},
		},
	}

	for _, c := range cases {
		cellsStr, position, turn := c.Input.CellsStr, c.Input.Position, c.Input.Turn

		b.FromStringCells(cellsStr)
		b.Turn = turn

		result := b.GetCellsToFlip(position.X, position.Y)

		if len(result) != len(c.Want) {
			t.Errorf("%s: Expected %d cell(s) to flip, got %d, %v", c.Name, len(c.Want), len(result), result)
			continue
		}

		for i, want := range c.Want {
			if result[i].X != want.X ||
				result[i].Y != want.Y ||
				result[i].NewState != want.NewState {
				t.Errorf("%s: Expected %v, got %v", c.Name, want, result[i])
			}

		}

	}
}
