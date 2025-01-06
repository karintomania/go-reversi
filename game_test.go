package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGameStart(t *testing.T) {
	var b Board

	b.init(3)

	g := NewGame(&b, Human, Human)

	player1CmdCh, player2CmdCh, player1GameCh, player2GameCh := g.Start()

	mockSync(player1GameCh, player2GameCh)
	assert.Equal(t, Player1Turn, g.State)

	cmd := GameCommand{CommandPlace, Position{0, 2}}
	player1CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	assert.Equal(t, Player2Turn, g.State)

	cmd = GameCommand{CommandPlace, Position{1, 2}}
	player2CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	assert.Equal(t, Player1Turn, g.State)

}

// discard channel output
func mockSync(player1GameCh, player2GameCh chan Game) {
	_ = <-player1GameCh
	_ = <-player2GameCh
}
