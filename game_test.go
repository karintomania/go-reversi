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
	assert.Equal(t, WaitingConnection, g.State)

	cmd := GameCommand{CommandType: CommandConnectionCheck}
	player1CmdCh <- cmd
	player2CmdCh <- cmd

	mockSync(player1GameCh, player2GameCh)

	assert.Equal(t, Player1Turn, g.State)

	cmd = GameCommand{CommandPlace, Position{0, 2}}
	player1CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	assert.Equal(t, Player2Turn, g.State)

	cmd = GameCommand{CommandPlace, Position{1, 2}}
	player2CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	assert.Equal(t, Player1Turn, g.State)

}

func TestGamePass(t *testing.T) {
	var b Board

	b.init(3)

	b.FromStringCells(
		[][]string{
			{"n", "n", "n"},
			{"w", "b", "b"},
			{"b", "w", "w"},
		},
	)

	g := NewGame(&b, Human, Human)

	player1CmdCh, player2CmdCh, player1GameCh, player2GameCh := g.Start()

	// connection check
	mockSync(player1GameCh, player2GameCh)
	cmd := GameCommand{CommandType: CommandConnectionCheck}
	player1CmdCh <- cmd
	player2CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	assert.Equal(t, Player1Turn, g.State)

	cmd = GameCommand{CommandPlace, Position{0, 0}}
	player1CmdCh <- cmd

	mockSync(player1GameCh, player2GameCh)
	assert.Equal(t, Player2Turn, g.State)
	assert.Equal(t, HasBlack, g.Board.Cells[0][0])

	cmd = GameCommand{CommandPlace, Position{2, 0}}
	player2CmdCh <- cmd

	mockSync(player1GameCh, player2GameCh)

	// Player1 is skipped
	assert.Equal(t, Player2Turn, g.State)
	assert.Equal(t, HasWhite, g.Board.Cells[0][2])

	cmd = GameCommand{CommandPlace, Position{1, 0}}
	player2CmdCh <- cmd

	mockSync(player1GameCh, player2GameCh)

	// game is finished as both player can't place
	assert.Equal(t, Finished, g.State)
	assert.Equal(t, "Black 3, White 6, Player 2 won", g.Message)
	assert.Equal(t, HasWhite, g.Board.Cells[0][1])
}

func TestReplay(t *testing.T) {
	var b Board

	b.init(3)

	b.FromStringCells(
		[][]string{
			{"n", "w", "w"},
			{"w", "w", "w"},
			{"b", "w", "w"},
		},
	)

	g := NewGame(&b, Human, Human)

	player1CmdCh, player2CmdCh, player1GameCh, player2GameCh := g.Start()

	// connection check
	mockSync(player1GameCh, player2GameCh)
	cmd := GameCommand{CommandType: CommandConnectionCheck}
	player1CmdCh <- cmd
	player2CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	assert.Equal(t, Player1Turn, g.State)

	cmd = GameCommand{CommandPlace, Position{0, 0}}
	player1CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)
	assert.Equal(t, Finished, g.State)

	// replay
	cmd = GameCommand{CommandType: CommandReplay}
	player1CmdCh <- cmd

	// the turn is swapped
	mockSync(player1GameCh, player2GameCh)
	assert.Equal(t, Player2Turn, g.State)

	g.Board.FromStringCells(
		[][]string{
			{"n", "w", "w"},
			{"w", "w", "w"},
			{"b", "w", "w"},
		},
	)

	cmd = GameCommand{CommandPlace, Position{0, 0}}
	player2CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	// finish the game
	assert.Equal(t, Finished, g.State)

	// replay
	cmd = GameCommand{CommandType: CommandReplay}
	player2CmdCh <- cmd

	// the turn is swapped
	mockSync(player1GameCh, player2GameCh)
	assert.Equal(t, Player1Turn, g.State)
}

// discard channel output
func mockSync(player1GameCh, player2GameCh chan Game) {
	_ = <-player1GameCh
	_ = <-player2GameCh
}
