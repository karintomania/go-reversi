package main

import (
	"fmt"
)

type Board struct {
	N        int
	Cells    [][]State
	Turn     Turn
	Position Position
}

func (b *Board) init(n int) {
	b.N = n

	b.Turn = Black

	b.Position = Position{0, 0}

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

func (b *Board) SwitchTurn() {
	if b.Turn == Black {
		b.Turn = White
	} else {
		b.Turn = Black
	}
}

func (b *Board) Pass() {
	// switch turn
	b.SwitchTurn()

}

func (b *Board) PlaceByAi() {
	b.Position = getAiPosition(b)
	b.Place()
}

func (b *Board) Place() error {
	// validate
	isValid := b.isCellAvailable()

	if !isValid {
		return fmt.Errorf("You can't place there.")
	}

	// get cells to flip
	cellsToFlip := b.GetCellsToFlip(b.Position.X, b.Position.Y)

	if len(cellsToFlip) == 0 {
		return fmt.Errorf("You can't place there.")
	}

	for _, cell := range cellsToFlip {
		b.Cells[cell.Y][cell.X] = cell.NewState
	}

	// Place stone
	p := b.Position
	if b.Turn == Black {
		b.Cells[p.Y][p.X] = HasBlack
	} else {
		b.Cells[p.Y][p.X] = HasWhite
	}

	// switch turn
	b.SwitchTurn()

	return nil
}

func (b *Board) HasPlayableCells() bool {
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

func (b *Board) isCellAvailable() bool {
	state := b.Cells[b.Position.Y][b.Position.X]
	return state == HasNothing
}

func (b *Board) MovePositionX(n int) {
	b.Position.addX(n, b.N)
}

func (b *Board) MovePositionY(n int) {
	b.Position.addY(n, b.N)
}

func (b *Board) GetCellsToFlip(x, y int) []CellToFlip {
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

func (b *Board) Finish() (int, int) {
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

func (b *Board) FromStringCells(cellsStr [][]string) {
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

func (b *Board) String() string {
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
	HasNothing State = iota
	HasBlack
	HasWhite
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
