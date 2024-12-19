package game

import (
	"fmt"

	"golang.org/x/net/websocket"
)

type Game struct {
	Url            string
	Conns          map[*websocket.Conn]bool
	Host           *websocket.Conn
	CurrentPlayer  *websocket.Conn
	GameMatrix     [][]byte
	EmptyPositions [][]int
}

func NewGame() *Game {
	matrix := make([][]byte, 3)
	for i := range matrix {
		matrix[i] = []byte{'.', '.', '.'}
	}
	positions := [][]int{
		{0, 0},
		{0, 1},
		{0, 2},
		{1, 0},
		{1, 1},
		{1, 2},
		{2, 0},
		{2, 1},
		{2, 2},
	}
	return &Game{
		Url:            "",
		Conns:          make(map[*websocket.Conn]bool),
		Host:           &websocket.Conn{},
		CurrentPlayer:  &websocket.Conn{},
		GameMatrix:     matrix,
		EmptyPositions: positions,
	}
}

func PrintGameMaxtrixToPlayers(game *Game) {
	for conn := range game.Conns {
		for i := 0; i < 3; i++ {
			conn.Write([]byte(fmt.Sprintf(" %c | %c | %c ", game.GameMatrix[i][0], game.GameMatrix[i][1], game.GameMatrix[i][2])))
			if i != 2 {
				conn.Write([]byte("---+---+---"))
			}
		}
	}
}

func ResetMatrix() [][]byte {
	matrix := make([][]byte, 3)
	for i := range matrix {
		matrix[i] = []byte{'.', '.', '.'}
	}
	return matrix
}

func ResetEmptyPositions() [][]int {
	positions := [][]int{
		{0, 0},
		{0, 1},
		{0, 2},
		{1, 0},
		{1, 1},
		{1, 2},
		{2, 0},
		{2, 1},
		{2, 2},
	}
	return positions
}

func IsMatrixWinner(matrix [][]byte, i, j int, symbol byte) bool {
	// Check row
	for col := 0; col < 3; col++ {
		if matrix[i][col] != symbol {
			break
		}
		if col == 2 { // Last column and all match
			return true
		}
	}
	// Check column
	for row := 0; row < 3; row++ {
		if matrix[row][j] != symbol {
			break
		}
		if row == 2 { // Last row and all match
			return true
		}
	}

	// Check aain diagonal
	if i == j { // Ensure it's a diagonal position
		for d := 0; d < 3; d++ {
			if matrix[d][d] != symbol {
				break
			}
			if d == 2 { // Last diagonal element and all match
				return true
			}
		}
	}

	// Check anti-diagonal
	if i+j == 2 { // Ensure it's an anti-diagonal position (diagonala secundara)
		for d := 0; d < 3; d++ {
			if matrix[d][2-d] != symbol {
				break
			}
			if d == 2 { // Last anti-diagonal element and all match
				return true
			}
		}
	}

	// No winner
	return false
}

func broadcast(currentGame *Game, message string) {
	for conn := range currentGame.Conns {
		conn.Write([]byte(message))
	}
}

func WonGame(currentGame *Game, winner *websocket.Conn, loser *websocket.Conn) {
	winner.Write([]byte("You won! Congratulations!"))
	loser.Write([]byte("You lost! Better luck next time!"))

	broadcast(currentGame, "Game has been reseted")
	currentGame.GameMatrix = ResetMatrix()
	currentGame.EmptyPositions = ResetEmptyPositions()
}

func DrawGame(currentGame *Game) {
	currentGame.GameMatrix = ResetMatrix()
	currentGame.EmptyPositions = ResetEmptyPositions()
	broadcast(currentGame, "Game was drawn! Game has been reseted")
}

func RemoveUserFromGameConn(currentGame *Game, userToRemove *websocket.Conn) {
	delete(currentGame.Conns, userToRemove)
}
