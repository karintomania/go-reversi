package main

import (
	"fmt"
	"math"
	"strings"
)

// mobility[index][0:black/1:white][cell position in row] = [backward flip cells num, forward flip cells num]
type Mobility map[Idx]map[Turn][][]int
type Lines map[LineId]Idx
type LineId int

type Idx int // state of one line i.e., if there is no disk, Idx = 0

type LineForCell struct {
	LineId LineId
	Local  int // local position of the cell in the idx
}

type IndexedBoard struct {
	N     int // dimension of the board
	CellN int // number of cells (e.g., if N=8, CellN=64)
	IdxN  int // number of possible pattern for one index (if N=8, it's 3^8)
	// number of indexes
	//(e.g., if N=8, LineN=38=(8 rows + 8 cols + 11 left top + 11 right top)
	LineN int

	Lines Lines // indexes for each rows/columns/diagnal lines

	// store line ids (row/col/diagnal) where a specific cell is in
	LineForCells [][]LineForCell

	mobility Mobility

	Turn Turn
}

func NewIndexedBoard(n int) IndexedBoard {
	b := IndexedBoard{N: n, Turn: Black}
	b.init()
	return b
}

func (b *IndexedBoard) init() {
	b.CellN = b.N * b.N

	b.LineN = b.calcLineN(b.N)

	b.IdxN = pow(3, b.N)

	b.LineForCells = b.calcLineForCells(b.N)
	b.mobility = b.calcMobility()

	b.initLines()
}

func (b *IndexedBoard) initLines() {
	b.Lines = make(map[LineId]Idx, b.LineN)

	for i := 0; i < b.LineN; i++ {
		b.Lines[LineId(i)] = 0
	}

	middle1 := b.N/2 - 1
	middle2 := b.N / 2
	initB1 := b.N*(middle1) + middle1
	initB2 := b.N*(middle2) + middle2
	initW1 := b.N*(middle1) + middle2
	initW2 := b.N*(middle2) + middle1

	b.Place(initB1, Black)
	b.Place(initB2, Black)
	b.Place(initW1, White)
	b.Place(initW2, White)
}

func (b *IndexedBoard) calcLineN(n int) int {
	// 2*b.N = N rows + N cols
	// There are 2*(2*b.N-1) diagnol line from left top to right bottom and the other way
	// 8 of them only have 1 or 2 cells, so it can't have any legal cell
	return 2*n + 2*(2*n-1) - 8
}

func (b *IndexedBoard) calcLineForCells(n int) [][]LineForCell {
	cellN := n * n

	lineForCells := make([][]LineForCell, cellN)

	// add row indexes
	for cell := 0; cell < cellN; cell++ {
		rowLine := cell / n
		colLine := cell%n + n

		// colIdx is local position in row idx and vice versa
		rowLineForCell := LineForCell{LineId(rowLine), colLine - n}
		colLineForCell := LineForCell{LineId(colLine), rowLine}
		lineForCells[cell] = []LineForCell{rowLineForCell, colLineForCell}
	}

	// add right top diagnal lines
	for i := 0; i < n*2-5; i++ {
		local := 0
		middle := (n*2 - 5) / 2

		// right top lines starts from n*2
		line := n*2 + i

		var startCell int
		if i < middle {
			// start with 3rd lowest row, first column, going upper row
			startCell = n * (n - 3 - i)
		} else {
			// start with 1st cell, moving left
			startCell = i - middle
		}

		for cell := startCell; cell < cellN; cell += n + 1 {
			lineForCells[cell] = append(lineForCells[cell], LineForCell{LineId(line), local})
			local++
			// finish if cell reaches the right end column
			if cell != startCell && cell%n == n-1 {
				break
			}
		}
	}

	// add left top diagnal lines
	for i := 0; i < n*2-5; i++ {
		local := 0
		middle := (n*2 - 5) / 2

		// left top line starts from n*2 + n*2 - 5
		line := n*2 + n*2 - 5 + i

		var startCell int
		if i < middle {
			// start with 1st row, 2nd cell, moving 1 col
			startCell = i + 2
		} else {
			// start with the end cell of first row, moving to the next row ading n
			startCell = (n - 1) + n*(i-middle)
		}

		for cell := startCell; cell < cellN; cell += n - 1 {
			lineForCells[cell] = append(lineForCells[cell], LineForCell{LineId(line), local})
			local++
			// finish if cell reaches left end colum
			if cell != startCell && cell%n == 0 {
				break
			}
		}
	}

	return lineForCells
}

func (b *IndexedBoard) calcMobility() Mobility {
	mobility := make(map[Idx]map[Turn][][]int, b.IdxN)

	// initialise
	for i := 0; i < b.IdxN; i++ {
		idx := Idx(i)

		blackLineFlipping := make([][]int, b.N)

		whiteLineFlipping := make([][]int, b.N)

		// init each indexes
		for j := 0; j < b.N; j++ {
			blackLineFlipping[j] = []int{0, 0}
			whiteLineFlipping[j] = []int{0, 0}
		}

		turnMap := make(map[Turn][][]int, 2)

		turnMap[Black] = blackLineFlipping
		turnMap[White] = whiteLineFlipping

		mobility[idx] = turnMap
	}

	for i := 0; i < b.IdxN; i++ {
		idx := Idx(i)
	loopLocal:
		for local := 0; local < b.N; local++ {
			var backwardFlip, forwardFlip int

			if getCellState(idx, local) != HasNothing {
				// already taken
				mobility[idx][Black][local] = []int{0, 0}
				mobility[idx][White][local] = []int{0, 0}
				continue loopLocal
			}

			for _, turn := range []Turn{Black, White} {
				var selfState, opponentState State
				if turn == Black {
					selfState = HasBlack
					opponentState = HasWhite
				} else {
					selfState = HasWhite
					opponentState = HasBlack
				}

				backwardFlip, forwardFlip = b.getFlippingCells(idx, local, selfState, opponentState)

				mobility[idx][turn][local] = []int{backwardFlip, forwardFlip}
			}
		}
	}

	return mobility
}

func (b *IndexedBoard) getFlippingCells(idx Idx, local int, selfState, opponentState State) (int, int) {
	var backwardFlip, forwardFlip int
	// backward
	if local >= 2 && getCellState(idx, local-1) == opponentState {
		backwardFlip++
	loopFlipBackward:
		for i := 2; i <= local; i++ {
			s := getCellState(idx, local-i)
			switch s {
			case opponentState:
				if i == local { // there is no ending disc
					backwardFlip = 0
					break loopFlipBackward
				}
				backwardFlip++
				continue loopFlipBackward
			case selfState:
				break loopFlipBackward
			case HasNothing:
				backwardFlip = 0
				break loopFlipBackward
			}
		}
	}

	if local < b.N-2 && getCellState(idx, local+1) == opponentState {
		forwardFlip++
		// check flipping for white
	loopFlipForward:
		for i := 2; i+local < b.N; i++ {
			s := getCellState(idx, local+i)
			switch s {
			case opponentState:
				if i+local < b.N { // there is no ending disc
					forwardFlip = 0
					break loopFlipForward
				}
				forwardFlip++
				continue loopFlipForward
			case selfState:
				break loopFlipForward
			case HasNothing:
				forwardFlip = 0
				break loopFlipForward
			}
		}
	}

	return backwardFlip, forwardFlip
}

func (b *IndexedBoard) IsLegal(cell int, t Turn) bool {
	idxForCells := b.LineForCells[cell]

	flippingCellsNum := 0
	for _, idxForCell := range idxForCells {
		lineId, local := idxForCell.LineId, idxForCell.Local
		idx := b.Lines[lineId]
		m := b.mobility[idx][t][local]
		flippingCellsNum += m[0] + m[1]
	}

	return flippingCellsNum > 0
}

// Place is not responsible for legality check
func (b *IndexedBoard) Place(cell int, t Turn) {
	lineForCells := b.LineForCells[cell]

	for _, lineForCell := range lineForCells {
		lineId, local := lineForCell.LineId, lineForCell.Local

		idx := b.Lines[lineId]

		m := b.mobility[idx][t][local]

		updatedIdx := b.updateIdx(idx, local, m[0], m[1], t)

		b.Lines[lineId] = updatedIdx
	}
}

func (b *IndexedBoard) updateIdx(
	idx Idx,
	local, backwardFlip, forwardFlip int,
	turn Turn) Idx {

	idxInt := int(idx)

	var ternary int

	if turn == Black {
		ternary = 1
	} else {
		ternary = 2
	}

	// add local
	idxInt += pow(3, local) * ternary

	// flip backward
	for i := 0; i < backwardFlip; i++ {
		idxInt += pow(3, local-i-1) * (2*ternary - 3)
	}

	// add forward
	for i := 0; i < forwardFlip; i++ {
		idxInt += pow(3, local+i+1) * (2*ternary - 3)
	}

	return Idx(idxInt)
}

func (b *IndexedBoard) SwitchTurn() {
	if b.Turn == Black {
		b.Turn = White
	} else {
		b.Turn = Black
	}
}

func (b *IndexedBoard) FromStringCells(cellsStr [][]string) {

	// reset lines
	for lineId := range b.Lines {
		b.Lines[lineId] = Idx(0)
	}

	// place according to the string
	for y, row := range cellsStr {
		for x, char := range row {
			cell := y*b.N + x
			switch char {
			case "b":
				b.Place(cell, Black)
			case "w":
				b.Place(cell, White)
			}
		}
	}
}

func (b *IndexedBoard) String() string {
	var builder strings.Builder

	fmt.Fprintln(&builder, "")
	for i := 0; i < b.N; i++ {
		for j := 0; j < b.N; j++ {
			idx := b.Lines[LineId(i)]

			ternary := int(idx) / pow(3, j) % b.N

			fmt.Fprintf(&builder, "|%d", ternary)
		}
		fmt.Fprintln(&builder, "|")
	}

	return builder.String()
}

func pow(x, y int) int {
	return int(math.Pow(float64(x), float64(y)))
}

func getCellState(idx Idx, local int) State {
	ternary := int(idx) / pow(3, local) % 3

	return State(ternary)
}
