package main

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexedBoard(t *testing.T) {
	logger = NewLogger(slog.LevelInfo)
	b := NewBoard(3)

	assert.Equal(t, 3, b.N)
}

func TestIndexedBoardCalcIdxN(t *testing.T) {
	logger = NewLogger(slog.LevelInfo)
	b := NewBoard(3)

	idxN := b.calcLineN(8)

	assert.Equal(t, 38, idxN)

	idxN = b.calcLineN(3)

	assert.Equal(t, 8, idxN)
}

func TestIndexedBoardCalcLinesForCell(t *testing.T) {
	logger = NewLogger(slog.LevelInfo)
	n := 4

	idxForCells := NewLineForCells(n)

	idxForCell5 := idxForCells[5]

	assert.Equal(t, LineForCell{1, 1, Row}, idxForCell5[0])
	assert.Equal(t, LineForCell{5, 1, Col}, idxForCell5[1])
	assert.Equal(t, LineForCell{9, 1, BackSlash}, idxForCell5[2])
	assert.Equal(t, LineForCell{11, 1, Slash}, idxForCell5[3])

	idxForCell9 := idxForCells[9]

	assert.Equal(t, LineForCell{2, 1, Row}, idxForCell9[0])
	assert.Equal(t, LineForCell{5, 2, Col}, idxForCell9[1])
	assert.Equal(t, LineForCell{8, 1, BackSlash}, idxForCell9[2])
	assert.Equal(t, LineForCell{12, 2, Slash}, idxForCell9[3])

	idxForCell10 := idxForCells[10]

	assert.Equal(t, LineForCell{2, 2, Row}, idxForCell10[0])
	assert.Equal(t, LineForCell{6, 2, Col}, idxForCell10[1])
	assert.Equal(t, LineForCell{9, 2, BackSlash}, idxForCell10[2])
	assert.Equal(t, LineForCell{13, 1, Slash}, idxForCell10[3])
}

func TestIndexedBoardCalcMobility(t *testing.T) {
	logger = NewLogger(slog.LevelInfo)
	n := 4

	m := NewMobility(n)

	assert.Equal(t, []int{0, 0}, m[Idx{0, n}][Black][0])
	assert.Equal(t, []int{1, 0}, m[Idx{5, n}][White][2])
	assert.Equal(t, []int{1, 0}, m[Idx{7, n}][Black][2])
	assert.Equal(t, []int{0, 1}, m[Idx{15, n}][Black][0])
	assert.Equal(t, []int{0, 1}, m[Idx{21, n}][White][0])
}

func TestIndexedBoardPlaceWithoutCheck(t *testing.T) {
	logger = NewLogger(slog.LevelInfo)

	n := 3

	b := NewBoard(n)

	// |1|2|0|
	// |2|1|0|
	// |0|0|0|
	assert.Equal(t, Idx{7, n}, b.Lines[LineId(0)])
	assert.Equal(t, Idx{5, n}, b.Lines[LineId(1)])
	assert.Equal(t, Idx{0, n}, b.Lines[LineId(2)])
	assert.Equal(t, Idx{7, n}, b.Lines[LineId(3)])
	assert.Equal(t, Idx{5, n}, b.Lines[LineId(4)])
	assert.Equal(t, Idx{0, n}, b.Lines[LineId(5)])
	assert.Equal(t, Idx{4, n}, b.Lines[LineId(6)])
	assert.Equal(t, Idx{3, n}, b.Lines[LineId(7)])

	b.PlaceWithoutCheck(2, Black)

	// |1|1|1|
	// |2|1|0|
	// |0|0|0|
	assert.Equal(t, Idx{13, n}, b.Lines[LineId(0)])
	assert.Equal(t, Idx{5, n}, b.Lines[LineId(1)])
	assert.Equal(t, Idx{0, n}, b.Lines[LineId(2)])
	assert.Equal(t, Idx{7, n}, b.Lines[LineId(3)])
	assert.Equal(t, Idx{4, n}, b.Lines[LineId(4)])
	assert.Equal(t, Idx{1, n}, b.Lines[LineId(5)])
	assert.Equal(t, Idx{4, n}, b.Lines[LineId(6)])
	assert.Equal(t, Idx{4, n}, b.Lines[LineId(7)])

	b.PlaceWithoutCheck(5, White)

	// |1|1|1|
	// |2|2|2|
	// |0|0|0|
	assert.Equal(t, Idx{13, n}, b.Lines[LineId(0)])
	assert.Equal(t, Idx{26, n}, b.Lines[LineId(1)])
	assert.Equal(t, Idx{0, n}, b.Lines[LineId(2)])
	assert.Equal(t, Idx{7, n}, b.Lines[LineId(3)])
	assert.Equal(t, Idx{7, n}, b.Lines[LineId(4)])
	assert.Equal(t, Idx{7, n}, b.Lines[LineId(5)])
	assert.Equal(t, Idx{7, n}, b.Lines[LineId(6)])
	assert.Equal(t, Idx{7, n}, b.Lines[LineId(7)])

	b.PlaceWithoutCheck(6, Black)

	// |1|1|1|
	// |1|1|2|
	// |1|0|0|
	assert.Equal(t, Idx{13, n}, b.Lines[LineId(0)])
	assert.Equal(t, Idx{22, n}, b.Lines[LineId(1)])
	assert.Equal(t, Idx{1, n}, b.Lines[LineId(2)])
	assert.Equal(t, Idx{13, n}, b.Lines[LineId(3)])
	assert.Equal(t, Idx{4, n}, b.Lines[LineId(4)])
	assert.Equal(t, Idx{7, n}, b.Lines[LineId(5)])
	assert.Equal(t, Idx{4, n}, b.Lines[LineId(6)])
	assert.Equal(t, Idx{13, n}, b.Lines[LineId(7)])

}

func TestIndexedBoardFromStringCells(t *testing.T) {
	logger = NewLogger(slog.LevelInfo)
	n := 3

	b := NewBoard(n)

	b.FromStringCells(
		[][]string{
			{"n", "n", "n"},
			{"b", "b", "b"},
			{"b", "w", "b"},
		},
	)

	assert.Equal(t, Idx{0, n}, b.Lines[LineId(0)])
	assert.Equal(t, Idx{13, n}, b.Lines[LineId(1)])
	assert.Equal(t, Idx{16, n}, b.Lines[LineId(2)])
	assert.Equal(t, Idx{12, n}, b.Lines[LineId(3)])
	assert.Equal(t, Idx{21, n}, b.Lines[LineId(4)])
	assert.Equal(t, Idx{12, n}, b.Lines[LineId(5)])
	assert.Equal(t, Idx{12, n}, b.Lines[LineId(6)])
	assert.Equal(t, Idx{12, n}, b.Lines[LineId(7)])
}

func TestIndexedBoardCount(t *testing.T) {
	logger = NewLogger(slog.LevelInfo)
	n := 3

	b := NewBoard(n)

	b.FromStringCells(
		[][]string{
			{"n", "n", "n"},
			{"b", "b", "b"},
			{"b", "w", "b"},
		},
	)

	totalB, totalW := b.Count()

	assert.Equal(t, 5, totalB)
	assert.Equal(t, 1, totalW)
}
