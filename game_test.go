package main

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGameStart(t *testing.T) {
	g, player1CmdCh, player2CmdCh, player1GameCh, player2GameCh, _, _ := gameTestInit(make([][]string, 0))

	mockSync(player1GameCh, player2GameCh)
	assert.Equal(t, WaitingConnection.String(), g.State.String())

	cmd := GameCommand{CommandType: CommandConnectionCheck}
	player1CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	player2CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	assert.Equal(t, Player1Turn, g.State)

	cmd = GameCommand{CommandType: CommandPlace, Position: Position{0, 2}}
	player1CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	assert.Equal(t, Player2Turn, g.State)

	cmd = GameCommand{CommandType: CommandPlace, Position: Position{1, 2}}
	player2CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	assert.Equal(t, Player1Turn, g.State)
}

func TestGamePass(t *testing.T) {
	g, player1CmdCh, player2CmdCh, player1GameCh, player2GameCh, _, _ := gameTestInit(
		[][]string{
			{"n", "n", "n"},
			{"w", "b", "b"},
			{"b", "w", "w"},
		},
	)

	// connection check
	mockSync(player1GameCh, player2GameCh)
	cmd := GameCommand{CommandType: CommandConnectionCheck}
	player1CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	player2CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	assert.Equal(t, Player1Turn, g.State)

	cmd = GameCommand{CommandType: CommandPlace, Position: Position{0, 0}}
	player1CmdCh <- cmd

	mockSync(player1GameCh, player2GameCh)
	assert.Equal(t, Player2Turn, g.State)
	assert.Equal(t, HasBlack, g.Board.GetCellState(Position{0, 0}))

	cmd = GameCommand{CommandType: CommandPlace, Position: Position{2, 0}}
	player2CmdCh <- cmd

	mockSync(player1GameCh, player2GameCh)

	// Player1 is skipped
	assert.Equal(t, Player2Turn, g.State)
	assert.Equal(t, HasWhite, g.Board.GetCellState(Position{0, 2}))

	cmd = GameCommand{CommandType: CommandPlace, Position: Position{1, 0}}
	player2CmdCh <- cmd

	mockSync(player1GameCh, player2GameCh)

	// game is finished as both player can't place
	assert.Equal(t, Finished, g.State)
	assert.Equal(t, fmt.Sprintf(messageWin, 3, 6, "Player 2"), g.Message)
	assert.Equal(t, HasWhite, g.Board.GetCellState(Position{0, 1}))
}

func TestReplay(t *testing.T) {
	g, player1CmdCh, player2CmdCh, player1GameCh, player2GameCh, _, _ := gameTestInit(
		[][]string{
			{"n", "w", "w"},
			{"w", "w", "w"},
			{"b", "w", "w"},
		},
	)

	// connection check
	mockSync(player1GameCh, player2GameCh)
	cmd := GameCommand{CommandType: CommandConnectionCheck}
	player1CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	player2CmdCh <- cmd
	mockSync(player1GameCh, player2GameCh)

	assert.Equal(t, Player1Turn, g.State)

	cmd = GameCommand{CommandType: CommandPlace, Position: Position{0, 0}}
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

	cmd = GameCommand{CommandType: CommandPlace, Position: Position{0, 0}}
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

func TestGameQuit(t *testing.T) {
	// var b Board

	// b.init(3)

	// // player 1
	// g := NewGame(&b, Human, Human)

	// _, _, _, _, player1QuitCh, _ := g.Start()

	g, _, _, _, _, player1QuitCh, _ := gameTestInit(make([][]string, 0))

	assert.Equal(t, Initialized, g.State)

	player1QuitCh <- true
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, Quit, g.State)

	// player 2
	g, _, _, _, _, _, player2QuitCh := gameTestInit(make([][]string, 0))

	assert.Equal(t, Initialized, g.State)

	player2QuitCh <- true
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, Quit, g.State)
}

func gameTestInit(initBoard [][]string) (*Game, chan GameCommand, chan GameCommand, chan Game, chan Game, chan bool, chan bool) {
	logger = NewLogger(slog.LevelInfo)

	b := NewBoard(3)

	if len(initBoard) > 0 {
		b.FromStringCells(initBoard)
	}

	g := NewGame(b, Human, Human)

	player1CmdCh, player2CmdCh, player1GameCh, player2GameCh, player1QuitCh, player2QuitCh := g.Start()
	return &g, player1CmdCh, player2CmdCh, player1GameCh, player2GameCh, player1QuitCh, player2QuitCh
}

// discard channel output
func mockSync(player1GameCh, player2GameCh chan Game) {
	_ = <-player1GameCh
	_ = <-player2GameCh
}
