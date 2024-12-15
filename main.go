package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"regexp"
	"slices"

	"golang.org/x/net/websocket"
)

type Game struct {
	url            string
	conns          map[*websocket.Conn]bool
	host           *websocket.Conn
	currentPlayer  *websocket.Conn
	gameMatrix     [][]byte
	emptyPositions [][]int
}

type Server struct {
	games []*Game
	conns map[*websocket.Conn]bool
}

type Question struct {
	Question string `json:"question"`
	A        string `json:"A"`
	B        string `json:"B"`
	C        string `json:"C"`
	D        string `json:"D"`
	Answer   string `json:"answer"`
}

func newServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]bool),
	}
}

var questionArray []Question

func newGame() *Game {
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
		url:            "",
		conns:          make(map[*websocket.Conn]bool),
		host:           &websocket.Conn{},
		currentPlayer:  &websocket.Conn{},
		gameMatrix:     matrix,
		emptyPositions: positions,
	}
}

func (s *Server) TicTacToe(ws *websocket.Conn) {
	s.ReadLoop(ws, s)
}

func printGameMaxtrixToPlayers(game *Game) {
	for conn := range game.conns {
		for i := 0; i < 3; i++ {
			conn.Write([]byte(fmt.Sprintf(" %c | %c | %c ", game.gameMatrix[i][0], game.gameMatrix[i][1], game.gameMatrix[i][2])))
			if i != 2 {
				conn.Write([]byte("---+---+---"))
			}
		}
	}
}

func isMatrixWinner(matrix [][]byte, i, j int, symbol byte) bool {
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

func removeSliceElement(emptyPositions *[][]int, searchedPosition []int) {
	for i := 0; i < len(*emptyPositions); i++ {
		if slices.Equal((*emptyPositions)[i], searchedPosition) {
			*emptyPositions = append((*emptyPositions)[:i], (*emptyPositions)[i+1:]...)
		}
	}
}

func giveRandomValidMatrixPosition(matrix [][]int) (i, j int) {
	n := rand.IntN(len(matrix))
	return matrix[n][0] + 1, matrix[n][1] + 1
}

func loadQuestionsJSON() []Question {
	dat, err := os.Open("multipleAnswer.json")
	if err != nil {
		panic(err)
	}
	var questionArray []Question
	byteValue, _ := io.ReadAll(dat)
	errMarshal := json.Unmarshal(byteValue, &questionArray)
	if errMarshal != nil {
		fmt.Println("Erorare marshall")
	}
	return questionArray
}

func pickRandomQuestion(questionArray []Question) Question {
	n := len(questionArray)

	return questionArray[rand.IntN(n)]
}

// We need to verify that there are 2 players otherwise there will be made 2 goroutines for the host
// And this way the reading feature will get broken (host 2 goroutines, other person 1 goroutine)
// Will look to improve the logic behind this
func RunGame(game *Game, isTwoPlayers bool) {
	for v := range game.conns {
		v.Write([]byte("Bine ai venit la joc , mesajul asta il vad doar jucatorii"))
	}
	if isTwoPlayers {
		for conn := range game.conns {
			go handleConnectionGame(conn, game)
		}
	}
	select {}
}

func handleConnectionGame(conn *websocket.Conn, game *Game) {
	buf := make([]byte, 1024)
	conn.Write([]byte("Please pick a number between a row and a column (1-3) (1-3)"))

	const (
		WaitingForPosition = iota
		WaitingForAnswer
	)
	state := WaitingForPosition

	var (
		num1, num2  int
		question    Question
		otherPlayer *websocket.Conn
	)

	regexVal := regexp.MustCompile(`^[1-3] [1-3]$`)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected gracefully")
			} else {
				fmt.Println("Error reading message in game:", err)
			}
			return
		}

		// If it's not this player's turn, just notify and continue
		if game.currentPlayer != conn {
			conn.Write([]byte("It's not your turn, wait!"))
			continue
		}

		message := buf[:n]

		switch state {
		case WaitingForPosition:
			fmt.Println(state)
			if regexVal.Match(message) {
				otherPlayer, err = findTheOtherPlayer(game, conn)
				if err != nil {
					fmt.Println("Error changing turn:", err)
					continue
				}

				num1, num2 = int(message[0]-'0'), int(message[2]-'0')
				fmt.Printf("Selected position: Row=%d, Col=%d\n", num1, num2)

				if game.gameMatrix[num1-1][num2-1] != '.' {
					conn.Write([]byte("That position is already taken, please choose another."))
					continue
				}

				// Pick a random question
				question = pickRandomQuestion(questionArray)
				conn.Write([]byte(fmt.Sprintf("Question %s\n A. %s\n B. %s\n C. %s\n D. %s\n ", question.Question, question.A, question.B, question.C, question.D)))

				// Switch to waiting for the answer
				state = 1
				fmt.Println("AM SCHIMBAT STATEUL")
				continue
			} else {
				conn.Write([]byte("Please write a row and a column in the format: (1-3) (1-3)"))
			}

		case WaitingForAnswer:
			fmt.Println("Sunt in stateul nou!")
			userAnswer := string(message)
			if userAnswer != question.Answer {
				conn.Write([]byte("You answered WRONG!!!"))
				num1, num2 = giveRandomValidMatrixPosition(game.emptyPositions)
			} else {
				conn.Write([]byte("You answered GOOD!"))
			}

			var isWinner bool
			// Place the symbol depending on who is playing (Host = 'X', Guest = 'O')
			if game.currentPlayer == game.host {
				game.gameMatrix[num1-1][num2-1] = 'X'
				removeSliceElement(&game.emptyPositions, []int{num1 - 1, num2 - 1})
				isWinner = isMatrixWinner(game.gameMatrix, num1-1, num2-1, 'X')
			} else {
				game.gameMatrix[num1-1][num2-1] = 'O'
				removeSliceElement(&game.emptyPositions, []int{num1 - 1, num2 - 1})
				isWinner = isMatrixWinner(game.gameMatrix, num1-1, num2-1, 'O')
			}

			printGameMaxtrixToPlayers(game)

			if isWinner {
				conn.Write([]byte("You won! Congratulations!"))
				otherPlayer.Write([]byte("You lost! Better luck next time!"))
				// Here you might want to end the game or reset it
				return
			}

			// Switch turn to the other player
			game.currentPlayer = otherPlayer
			// Prompt the next player for their move
			otherPlayer.Write([]byte("Please pick a number between a row and a column (1-3) (1-3)"))
			state = WaitingForPosition
		}
	}
}

func findTheOtherPlayer(game *Game, currUser *websocket.Conn) (*websocket.Conn, error) {
	for user := range game.conns {
		if user != currUser {
			return user, nil
		}
	}
	return currUser, errors.New("There is not other connection right now")
}

func (s *Server) ReadLoop(ws *websocket.Conn, server *Server) {
	buf := make([]byte, 1024)
	ws.Write([]byte("Would you want to create a lobby?(C) or join one?(J)"))
	for {

		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading from client", err)

			break

		}
		msg := buf[:n]

		if string(msg) == "C" {
			ws.Write([]byte("Please write a lobby id for your friend to join"))
			game := newGame()
			for {
				bytesRead, err := ws.Read(buf)
				if err != nil {
					break
				}
				urlName := buf[:bytesRead]
				game.conns[ws] = true
				game.url = string(urlName)
				game.host = ws
				game.currentPlayer = ws

				server.games = append(server.games, game)

				ws.Write([]byte(fmt.Sprintf("You have created a server with the url %s", game.url)))

				RunGame(game, false)
				return
			}
		} else if string(msg) == "J" {
			ws.Write([]byte("Please write the game url"))
			for {
				bytesRead, err := ws.Read(buf)
				if err != nil {
					break
				}
				var searchedGame *Game
				for _, v := range server.games {
					if v.url == string(buf[:bytesRead]) {
						searchedGame = v
						break
					}
				}
				if searchedGame.url == "" || len(searchedGame.conns) >= 2 {
					ws.Write([]byte("The server does not exist or there are too many players"))
				} else {
					searchedGame.conns[ws] = true
					ws.Write([]byte("Connected to the server!"))
					RunGame(searchedGame, true)
					return
				}
			}
		} else {
			ws.Write([]byte("Please type one of those characters"))
		}

	}
}

func (s *Server) broadcast(b []byte) {
	for ws := range s.conns {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(b); err != nil {
				fmt.Println("Error ", err)
			}
		}(ws)
	}
}

func main() {
	server := newServer()
	questionArray = loadQuestionsJSON()

	http.Handle("/ws", websocket.Handler(server.TicTacToe))
	http.ListenAndServe("127.0.0.1:3000", nil)
}
