package main

import (
	"math/rand"
)

func getAiPosition(b *Board) Position {
	return getRandom(b)
}

// place randomly
func getRandom(b *Board) Position {
	availableCells := make([]int, 0, b.CellN)

	for cell := 0; cell < b.CellN; cell++ {
		if !b.IsLegal(cell, b.Turn) {
			continue
		}

		availableCells = append(availableCells, cell)
	}

	key := rand.Intn(len(availableCells))

	cell := availableCells[key]

	return Position{cell % b.N, cell / b.N}
}
