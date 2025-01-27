package main

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdxPlaceOnLocal(t *testing.T) {
	logger = NewLogger(slog.LevelDebug)

	idx := Idx{0, 4}

	// place Black at 0
	idx.PlaceOnLocal(0, Black)
	assert.Equal(t, 1, idx.Value)

	// flip
	idx.PlaceOnLocal(0, White)
	assert.Equal(t, 2, idx.Value)

	// place Black at 2
	idx.PlaceOnLocal(2, Black)
	assert.Equal(t, 11, idx.Value)

	// flip
	idx.PlaceOnLocal(2, White)
	assert.Equal(t, 20, idx.Value)
}
