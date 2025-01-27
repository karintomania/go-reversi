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
	b.mobility = NewMobility(b.N)

	b.initLines()
}

func (b *IndexedBoard) initLines() {
	b.Lines = make(map[LineId]Idx, b.LineN)

	for i := 0; i < b.LineN; i++ {
		b.Lines[LineId(i)] = Idx{0, b.N}
	}

	middle1 := b.N/2 - 1
	middle2 := b.N / 2
	initB1 := b.N*(middle1) + middle1
	initB2 := b.N*(middle2) + middle2
	initW1 := b.N*(middle1) + middle2
	initW2 := b.N*(middle2) + middle1

	b.PlaceWithoutCheck(initB1, Black)
	b.PlaceWithoutCheck(initB2, Black)
	b.PlaceWithoutCheck(initW1, White)
	b.PlaceWithoutCheck(initW2, White)
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

func NewMobility(n int) Mobility {

	idxN := pow(3, n)
	mobility := make(map[Idx]map[Turn][][]int, idxN)

	// initialise
	for i := 0; i < idxN; i++ {
		idx := Idx{i, n}

		blackLineFlipping := make([][]int, n)

		whiteLineFlipping := make([][]int, n)

		// init each indexes
		for j := 0; j < n; j++ {
			blackLineFlipping[j] = []int{0, 0}
			whiteLineFlipping[j] = []int{0, 0}
		}

		turnMap := make(map[Turn][][]int, 2)

		turnMap[Black] = blackLineFlipping
		turnMap[White] = whiteLineFlipping

		mobility[idx] = turnMap
	}

	for i := 0; i < idxN; i++ {
		idx := Idx{i, n}
	loopLocal:
		for local := 0; local < n; local++ {
			var backwardFlip, forwardFlip int

			if idx.GetLocalState(local) != HasNothing {
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

				backwardFlip, forwardFlip = getFlippingCells(idx, local, n, selfState, opponentState)

				mobility[idx][turn][local] = []int{backwardFlip, forwardFlip}
			}
		}
	}

	return mobility
}

func getFlippingCells(idx Idx, local, n int, selfState, opponentState State) (int, int) {
	var backwardFlip, forwardFlip int
	// backward
	if local >= 2 && idx.GetLocalState(local-1) == opponentState {
		backwardFlip++
	loopFlipBackward:
		for i := 2; i <= local; i++ {
			s := idx.GetLocalState(local - i)
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

	if local < n-2 && idx.GetLocalState(local+1) == opponentState {
		forwardFlip++
		// check flipping for white
	loopFlipForward:
		for i := 2; i+local < n; i++ {
			s := idx.GetLocalState(local + i)
			switch s {
			case opponentState:
				if i+local < n { // there is no ending disc
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

func (b *IndexedBoard) HasLegalMove(t Turn) bool {
	hasLegal := false
	for i := 0; i < b.CellN; i++ {
		hasLegal = hasLegal || b.IsLegal(i, t)
	}
	return hasLegal
}

func (b *IndexedBoard) Place(cell int, t Turn) {
	if b.HasLegalMove(t) {
		b.PlaceWithoutCheck(cell, t)

		b.SwitchTurn()
	}
}

// PlaceWithoutCheck only place the disk, without the legality or switching turn
func (b *IndexedBoard) PlaceWithoutCheck(cell int, t Turn) {
	lineForCells := b.LineForCells[cell]

	for _, lineForCell := range lineForCells {
		lineId, local := lineForCell.LineId, lineForCell.Local

		idx := b.Lines[lineId]

		m := b.mobility[idx][t][local]

		// flip / place the disk
		for i := -m[0]; i <= m[1]; i++ {
			idx.PlaceOnLocal(local+i, t)
		}

		b.Lines[lineId] = idx
	}
}

func (b *IndexedBoard) SwitchTurn() {
	b.Turn = Turn(!bool(b.Turn))
}

func (b *IndexedBoard) Count() (int, int) {
	var totalB, totalW int

	for i := 0; i < b.N; i++ {
		idx := b.Lines[LineId(i)]
		for local := 0; local < b.N; local++ {
			state := idx.GetLocalState(local)

			switch state {
			case HasBlack:
				totalB++
			case HasWhite:
				totalW++
			}
		}
	}

	return totalB, totalW
}

func (b *IndexedBoard) FromStringCells(cellsStr [][]string) {
	// reset lines
	for lineId := range b.Lines {
		b.Lines[lineId] = Idx{0, b.N}
	}

	// place according to the string
	for y, row := range cellsStr {
		for x, char := range row {
			cell := y*b.N + x

			for _, lineForCell := range b.LineForCells[cell] {
				lineId, local := lineForCell.LineId, lineForCell.Local

				idx := b.Lines[lineId]

				switch char {
				case "b":
					idx.PlaceOnLocal(local, Black)
				case "w":
					idx.PlaceOnLocal(local, White)
				}

				b.Lines[lineId] = idx
			}
		}
	}
}

func (b *IndexedBoard) String() string {
	var builder strings.Builder

	fmt.Fprintln(&builder, "")
	for i := 0; i < b.N; i++ {
		idx := b.Lines[LineId(i)]

		fmt.Fprintln(&builder, idx.String())
	}

	return builder.String()
}

func pow(x, y int) int {
	return int(math.Pow(float64(x), float64(y)))
}

// state of one line i.e., if there is no disk, Idx = 0
type Idx struct {
	Value int
	N     int
}

func (idx *Idx) GetLocalState(local int) State {
	ternary := idx.Value / pow(3, local) % idx.N

	return State(ternary)
}

func (idx *Idx) PlaceOnLocal(local int, turn Turn) {
	current := idx.GetLocalState(local)
	var diff int

	if turn == Black {
		diff = 1 - int(current)
	} else {
		diff = 2 - int(current)
	}

	newValue := idx.Value + diff*pow(3, local)

	idx.Value = newValue
}

func (idx *Idx) String() string {
	var builder strings.Builder

	for i := 0; i < idx.N; i++ {
		ternary := int(idx.Value) / pow(3, i) % idx.N

		fmt.Fprintf(&builder, "|%d", ternary)
	}
	fmt.Fprint(&builder, "|")

	return builder.String()
}
