package main

import (
	"math/rand"
)

func getAiPosition(b *Board) Position {

	f := rand.Float32()

	if f < 0.4 {
		return getMinGain(b)
	} else if f < 0.7 {
		return getMaxGain(b)
	} else {
		return getRandom(b)
	}
}

// place where the number of flipped cells are minimum
func getMinGain(b *Board) Position {
	minCount := 9999
	tmpPosition := Position{0, 0}

	for x := 0; x < b.N; x++ {
		for y := 0; y < b.N; y++ {
			if b.Cells[y][x] != HasNothing {
				continue
			}

			result := b.GetCellsToFlip(x, y)

			gain := len(result)
			if gain > 0 && gain < minCount {
				minCount = gain
				tmpPosition.X = x
				tmpPosition.Y = y
			}
		}
	}

	return tmpPosition
}

// place where the number of flipped cells are maximum
func getMaxGain(b *Board) Position {
	maxCount := 0
	tmpPosition := Position{0, 0}

	for x := 0; x < b.N; x++ {
		for y := 0; y < b.N; y++ {
			if b.Cells[y][x] != HasNothing {
				continue
			}

			result := b.GetCellsToFlip(x, y)

			if len(result) > maxCount {
				maxCount = len(result)
				tmpPosition.X = x
				tmpPosition.Y = y
			}
		}
	}

	return tmpPosition
}

// place randomly
func getRandom(b *Board) Position {
	positions := make([]Position, 0, 100)

	for x := 0; x < b.N; x++ {
		for y := 0; y < b.N; y++ {
			if b.Cells[y][x] != HasNothing {
				continue
			}

			result := b.GetCellsToFlip(x, y)

			gain := len(result)
			if gain > 0 {
				positions = append(positions, Position{x, y})

			}
		}
	}

	key := rand.Intn(len(positions))

	return positions[key]
}
