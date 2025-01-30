package main

import (
	"fmt"
)

type ArrayBoard struct {
	N     int
	Cells [][]State
	Turn  Turn
}

func NewArrayBoard(n int) *ArrayBoard {
	b := &ArrayBoard{}
	b.init(n)
	return b
}

func (b *ArrayBoard) init(n int) {
	b.N = n

	b.Turn = Black

	cells := make([][]State, n)

	if n <= 2 {
		panic("Board dimension needs to be more than 2. Even numbers are recommended")
	}

	middle := n/2 - 1

	for y := 0; y < n; y++ {
		cells[y] = make([]State, n)
		for x := 0; x < n; x++ {
			if y == middle && x == middle {
				cells[y][x] = HasBlack
			} else if y == middle+1 && x == middle+1 {
				cells[y][x] = HasBlack
			} else if y == middle && x == middle+1 {
				cells[y][x] = HasWhite
			} else if y == middle+1 && x == middle {
				cells[y][x] = HasWhite
			} else {
				cells[y][x] = HasNothing
			}
		}
	}

	b.Cells = cells
}

func (b *ArrayBoard) SwitchTurn() {
	if b.Turn == Black {
		b.Turn = White
	} else {
		b.Turn = Black
	}
}

func (b *ArrayBoard) Pass() {
	// switch turn
	b.SwitchTurn()

}

func (b *ArrayBoard) Place(p Position) error {
	// check if the cell is taken
	isValid := b.isCellTaken(p)
	if !isValid {
		return fmt.Errorf("You can't place there.")
	}

	// get cells to flip
	cellsToFlip := b.GetCellsToFlip(p.X, p.Y)

	// error if no cells to flip
	if len(cellsToFlip) == 0 {
		return fmt.Errorf("You can't place there.")
	}

	for _, cell := range cellsToFlip {
		b.Cells[cell.Y][cell.X] = cell.NewState
	}

	// Place stone
	if b.Turn == Black {
		b.Cells[p.Y][p.X] = HasBlack
	} else {
		b.Cells[p.Y][p.X] = HasWhite
	}

	// switch turn
	b.SwitchTurn()

	return nil
}

func (b *ArrayBoard) HasLegalCells() bool {
	for x := 0; x < b.N; x++ {
		for y := 0; y < b.N; y++ {
			if b.Cells[y][x] != HasNothing {
				continue
			}

			result := b.GetCellsToFlip(x, y)

			if len(result) > 0 {
				return true
			}
		}
	}

	return false
}

func (b *ArrayBoard) isCellTaken(p Position) bool {
	state := b.Cells[p.Y][p.X]
	return state == HasNothing
}

func (b *ArrayBoard) GetCellsToFlip(x, y int) []CellToFlip {
	cells := make([]CellToFlip, 0, b.N)

	var selfState State
	var opponentState State

	if b.Turn == Black {
		selfState = HasBlack
		opponentState = HasWhite
	} else {
		selfState = HasWhite
		opponentState = HasBlack
	}

	// check horizontally to right
	tempCells := make([]CellToFlip, 0)

	if x < b.N-2 && b.Cells[y][x+1] == opponentState {
	loop1:
		for i := x + 1; i < b.N; i++ {
			switch b.Cells[y][i] {
			case HasNothing:
				break loop1
			case opponentState:
				tempCells = append(tempCells, CellToFlip{i, y, selfState})
			case selfState:
				cells = append(cells, tempCells...)
				break loop1
			}
		}
	}

	// check horizontally to left
	tempCells = make([]CellToFlip, 0)

	if x >= 2 && b.Cells[y][x-1] == opponentState {
	loopHorLeft:
		for i := x - 1; i >= 0; i-- {
			switch b.Cells[y][i] {
			case HasNothing:
				break loopHorLeft
			case opponentState:
				tempCells = append(tempCells, CellToFlip{i, y, selfState})
			case selfState:
				cells = append(cells, tempCells...)
				break loopHorLeft
			}
		}
	}

	// check vertically to bottom
	tempCells = make([]CellToFlip, 0)

	if y < b.N-2 && b.Cells[y+1][x] == opponentState {
	loopVerBottom:
		for i := y + 1; i < b.N; i++ {
			switch b.Cells[i][x] {
			case HasNothing:
				break loopVerBottom
			case opponentState:
				tempCells = append(tempCells, CellToFlip{x, i, selfState})
			case selfState:
				cells = append(cells, tempCells...)
				break loopVerBottom
			}
		}
	}

	// check vertically to top
	tempCells = make([]CellToFlip, 0)

	if y >= 2 && b.Cells[y-1][x] == opponentState {
	loopVerTop:
		for i := y - 1; i >= 0; i-- {
			switch b.Cells[i][x] {
			case HasNothing:
				break loopVerTop
			case opponentState:
				tempCells = append(tempCells, CellToFlip{x, i, selfState})
			case selfState:
				cells = append(cells, tempCells...)
				break loopVerTop
			}
		}
	}

	// check diagonally to bottom right
	tempCells = make([]CellToFlip, 0)

	if x < b.N-2 && y < b.N-2 && b.Cells[y+1][x+1] == opponentState {
	loopDiagBottomRight:
		for i := 1; y+i < b.N && x+i < b.N; i++ {
			switch b.Cells[y+i][x+i] {
			case HasNothing:
				break loopDiagBottomRight
			case opponentState:
				tempCells = append(tempCells, CellToFlip{x + i, y + i, selfState})
			case selfState:
				cells = append(cells, tempCells...)
				break loopDiagBottomRight
			}
		}
	}

	// check diagonally to bottom left
	tempCells = make([]CellToFlip, 0)

	if x >= 2 && y < b.N-2 && b.Cells[y+1][x-1] == opponentState {
	loopDiagBottomLeft:
		for i := 1; y+i < b.N && x-i >= 0; i++ {
			switch b.Cells[y+i][x-i] {
			case HasNothing:
				break loopDiagBottomLeft
			case opponentState:
				tempCells = append(tempCells, CellToFlip{x - i, y + i, selfState})
			case selfState:
				cells = append(cells, tempCells...)
				break loopDiagBottomLeft
			}
		}
	}

	// check diagonally to top right
	tempCells = make([]CellToFlip, 0)

	if x < b.N-2 && y >= 2 && b.Cells[y-1][x+1] == opponentState {
	loopDiagTopRight:
		for i := 1; y-i >= 0 && x+i < b.N; i++ {
			switch b.Cells[y-i][x+i] {
			case HasNothing:
				break loopDiagTopRight
			case opponentState:
				tempCells = append(tempCells, CellToFlip{x + i, y - i, selfState})
			case selfState:
				cells = append(cells, tempCells...)
				break loopDiagTopRight
			}
		}
	}

	// check diagonally to top left
	tempCells = make([]CellToFlip, 0)

	if x >= 2 && y >= 2 && b.Cells[y-1][x-1] == opponentState {
	loopDiagTopLeft:
		for i := 1; y-i >= 0 && x-i >= 0; i++ {
			switch b.Cells[y-i][x-i] {
			case HasNothing:
				break loopDiagTopLeft
			case opponentState:
				tempCells = append(tempCells, CellToFlip{x - i, y - i, selfState})
			case selfState:
				cells = append(cells, tempCells...)
				break loopDiagTopLeft
			}
		}
	}
	return cells
}

func (b *ArrayBoard) Count() (int, int) {
	var totalB, totalW int

	for x := 0; x < b.N; x++ {
		for y := 0; y < b.N; y++ {
			if b.Cells[y][x] == HasWhite {
				totalW++
			} else if b.Cells[y][x] == HasBlack {
				totalB++
			}
		}
	}

	return totalB, totalW
}

func (b *ArrayBoard) FromStringCells(cellsStr [][]string) {
	for y, row := range cellsStr {
		for x, s := range row {
			var state State

			if s == "n" {
				state = HasNothing
			} else if s == "b" {
				state = HasBlack
			} else {
				state = HasWhite
			}

			b.Cells[y][x] = state
		}
	}
}

func (b *ArrayBoard) String() string {
	var result string

	for _, row := range b.Cells {
		var rowStr string
		for _, s := range row {
			rowStr += s.String() + " "
		}
		result += fmt.Sprintf("\n%s", rowStr)
	}

	return result
}

type State int

const (
	HasNothing State = 0
	HasBlack   State = 1
	HasWhite   State = 2
)

func (s State) String() string {
	if s == HasNothing {
		return "n"
	} else if s == HasBlack {
		return "b"
	} else {
		return "w"
	}
}

type Position struct {
	X int
	Y int
}

func cellToPosition(n int, cell int) Position {
	return Position{cell % n, cell / n}
}

func (p *Position) addX(n int, maxN int) {
	p.X = min(max(p.X+n, 0), maxN-1)
}

func (p *Position) addY(n int, maxN int) {
	p.Y = min(max(p.Y+n, 0), maxN-1)
}

type Turn bool

const (
	Black Turn = false
	White Turn = true
)

func (t Turn) String() string {
	if t == Black {
		return "○"
	} else {
		return "●"
	}
}

type CellToFlip struct {
	X        int
	Y        int
	NewState State
}
