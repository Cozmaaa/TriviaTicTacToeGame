package utils

import (
	"cozmaaa/xon/internal/game"
	"errors"
	"math/rand/v2"
	"slices"

	"golang.org/x/net/websocket"
)

func RemoveSliceElement(emptyPositions *[][]int, searchedPosition []int) {
	for i := 0; i < len(*emptyPositions); i++ {
		if slices.Equal((*emptyPositions)[i], searchedPosition) {
			*emptyPositions = append((*emptyPositions)[:i], (*emptyPositions)[i+1:]...)
		}
	}
}

func GiveRandomValidMatrixPosition(matrix [][]int) (i, j int) {
	n := rand.IntN(len(matrix))
	return matrix[n][0] + 1, matrix[n][1] + 1
}

func FindTheOtherPlayer(currentGame *game.Game, currUser *websocket.Conn) (*websocket.Conn, error) {
	for user := range currentGame.Conns {
		if user != currUser {
			return user, nil
		}
	}
	return currUser, errors.New("There is not other connection right now")
}
