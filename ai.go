package main

import (
	"log/slog"
	"math/rand"
)

type ScoreTable []map[Idx]int

type AiPlayer struct {
	N             int
	Colour        Turn
	depth         int
	evalCount     int
	ScoreTable    ScoreTable // store the pre-calculated score for each row
	EndScoreTable ScoreTable // store the pre-calculated score for each row
}

func NewAiPlayer(n int) *AiPlayer {
	ap := AiPlayer{N: n, depth: 7}

	ap.calcScoreTable()

	return &ap
}

func (ap *AiPlayer) calcScoreTable() {
	var cellScore [][]int

	scoreTable := make(ScoreTable, ap.N)
	endScoreTable := make(ScoreTable, ap.N)

	switch ap.N {
	case 3:
		cellScore = cellScore3
	case 4:
		cellScore = cellScore4
	case 5:
		cellScore = cellScore5
	case 6:
		cellScore = cellScore6
	case 7:
		cellScore = cellScore7
	case 8:
		cellScore = cellScore8
	}

	idxN := pow(3, ap.N)

	for row := 0; row < ap.N; row++ {
		scoreTable[row] = make(map[Idx]int)
		endScoreTable[row] = make(map[Idx]int)
		for idx := 0; idx < idxN; idx++ {
			score := 0
			endScore := 0

			for local := 0; local < ap.N; local++ {
				localState := idx / pow(3, local) % 3
				if localState == 1 {
					// add score if cell is black
					score += cellScore[row][local]
					endScore++
				} else if localState == 2 {
					// minus score if cell is white
					score -= cellScore[row][local]
					endScore--
				}
			}
			scoreTable[row][Idx{idx, ap.N}] = score
			endScoreTable[row][Idx{idx, ap.N}] = endScore
		}
	}

	ap.ScoreTable = scoreTable
	ap.EndScoreTable = endScoreTable
}

func (ap *AiPlayer) getPosition(b *Board) Position {
	ap.Colour = b.Turn
	return ap.getBest(b)
}

// place randomly

func (ap *AiPlayer) getBest(b *Board) Position {
	ap.evalCount = 0
	depth := ap.depth
	bestCell := 0
	alpha := -9999
	beta := 9999

	var score int

	canUseFinal := b.CountEmptyCells() < depth
	// canUseFinal = true

	for cell := 0; cell < b.CellN; cell++ {
		if !b.IsLegal(cell, b.Turn) {
			continue
		}

		// evaluate the current board
		b, _ := b.Place(ap.cellToPosition(cell))

		score = -ap.negMax(b, depth-1, -beta, -alpha, false, canUseFinal)

		logger.Debug("point ", slog.Any("cell", ap.cellToPosition(cell)), slog.Int("score", score))

		if score > alpha {
			bestCell = cell
			alpha = score
		}
	}

	logger.Debug("best  ", slog.Any("cell", ap.cellToPosition(bestCell)), slog.Int("score", alpha))

	logger.Debug("evaluated ", slog.Int("evalCount", ap.evalCount))
	return ap.cellToPosition(bestCell)
}

func (ap *AiPlayer) negMax(b *Board, depth, alpha, beta int, passed bool, canUseFinal bool) int {
	var evaluate func() int

	if canUseFinal {
		evaluate = func() int { return ap.evaluateFinalBoard(b) }
	} else {
		evaluate = func() int { return ap.evaluate(b) }
	}
	max := -9999

	var score int

	if depth == 0 {
		// evaluate the current board
		return evaluate()
	}

	if !b.HasLegalMove(b.Turn) {
		if passed {
			// game finished
			return evaluate()
		} else {
			b.SwitchTurn()
			return -ap.negMax(b, depth, -beta, -alpha, true, canUseFinal)
		}
	}

	for cell := 0; cell < b.CellN; cell++ {
		if !b.IsLegal(cell, b.Turn) {
			continue
		}

		// evaluate the current board
		b, _ := b.Place(ap.cellToPosition(cell))

		score = -ap.negMax(b, depth-1, -beta, -alpha, false, canUseFinal)

		if score >= beta {
			return score
		}

		if score > alpha {
			alpha = score
		}

		if score > max {
			max = score
		}
	}

	return max
}

// evaluate the finished game
func (ap *AiPlayer) evaluateFinalBoard(b *Board) int {
	score := 0

	for i := 0; i < ap.N; i++ {
		line := b.Lines[LineId(i)]
		score += ap.EndScoreTable[i][line]
	}

	if !b.Turn == Black {
		score = -score
	}

	ap.evalCount++
	return score
}

func (ap *AiPlayer) evaluate(b *Board) int {
	// score is positive when
	// - AI is white and depth is even
	// - AI is black and depth is odd
	score := 0

	for i := 0; i < ap.N; i++ {
		line := b.Lines[LineId(i)]
		score += ap.ScoreTable[i][line]
	}

	if !b.Turn == Black {
		score = -score
	}

	ap.evalCount++
	return score
}

// evalueate function is for black
// if this returns negative, use the score * -1
// func (ap *AiPlayer) evaluatePositive() bool {
// 	if ap.Colour == Black {
// 		return ap.depth%2 == 1
// 	} else {
// 		return ap.depth%2 == 0
// 	}
// }

func (ap *AiPlayer) cellToPosition(cell int) Position {
	return cellToPosition(ap.N, cell)
}

var cellScore3 = [][]int{
	{1, 1, 1},
	{1, 1, 1},
	{1, 1, 1},
}

var cellScore4 = [][]int{
	{10, -5, -5, 10},
	{-5, -1, -1, -5},
	{-5, -1, -1, -5},
	{10, -5, -5, 10},
}

var cellScore5 = [][]int{
	{30, -12, 0, -12, 30},
	{-12, -15, -3, -15, -12},
	{0, -3, 0, -3, 0},
	{-12, -15, -3, -15, -12},
	{30, -12, 0, -12, 30},
}

var cellScore6 = [][]int{
	{30, -12, 0, 0, -12, 30},
	{-12, -15, -3, -3, -15, -12},
	{0, -3, -1, -1, -3, 0},
	{0, -3, -1, -1, -3, 0},
	{-12, -15, -3, -3, -15, -12},
	{30, -12, 0, 0, -12, 30},
}

var cellScore7 = [][]int{
	{30, -12, 0, -1, 0, -12, 30},
	{-12, -15, -3, -3, -3, -15, -12},
	{0, -3, 0, -1, 0, -3, 0},
	{-1, -3, -1, -1, -1, -3, -1},
	{0, -3, 0, -1, 0, -3, 0},
	{-12, -15, -3, -3, -3, -15, -12},
	{30, -12, 0, -1, 0, -12, 30},
}

var cellScore8 = [][]int{
	{30, -12, 0, -1, -1, 0, -12, 30},
	{-12, -15, -3, -3, -3, -3, -15, -12},
	{0, -3, 0, -1, -1, 0, -3, 0},
	{-1, -3, -1, -1, -1, -1, -3, -1},
	{-1, -3, -1, -1, -1, -1, -3, -1},
	{0, -3, 0, -1, -1, 0, -3, 0},
	{-12, -15, -3, -3, -3, -3, -15, -12},
	{30, -12, 0, -1, -1, 0, -12, 30},
}

func (ap *AiPlayer) getRandom(b *Board) Position {
	availableCells := make([]int, 0, b.CellN)

	for cell := 0; cell < b.CellN; cell++ {
		if !b.IsLegal(cell, b.Turn) {
			continue
		}

		availableCells = append(availableCells, cell)
	}

	key := rand.Intn(len(availableCells))

	cell := availableCells[key]

	return ap.cellToPosition(cell)
}
