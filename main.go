package main

import (
	"cozmaaa/xon/internal/api"
	"cozmaaa/xon/internal/game"
	"cozmaaa/xon/internal/questions"
	"cozmaaa/xon/internal/utils"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"golang.org/x/net/websocket"
)

type Server struct {
	Games []*game.Game
	Conns map[*websocket.Conn]bool
}

func newServer() *Server {
	return &Server{
		Conns: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) removeGameFromServer(currentGame *game.Game) {
	var searchedIndex int = -1
	for i := 0; i < len(s.Games); i++ {
		if s.Games[i] == currentGame {
			searchedIndex = i
			break
		}
	}
	if searchedIndex == -1 {
		return
	}

	s.Games = append(s.Games[:searchedIndex], s.Games[searchedIndex+1:]...)
}

var questionArray []questions.Question

func (s *Server) TicTacToe(ws *websocket.Conn) {
	s.ReadLoop(ws, s)
}

// We need to verify that there are 2 players otherwise there will be made 2 goroutines for the host
// And this way the reading feature will get broken (host 2 goroutines, other person 1 goroutine)
// Will look to improve the logic behind this
func RunGame(game *game.Game, isTwoPlayers bool, server *Server) {
	for v := range game.Conns {
		v.Write([]byte("Bine ai venit la joc , mesajul asta il vad doar jucatorii"))
	}
	if isTwoPlayers {
		for conn := range game.Conns {
			go handleConnectionGame(conn, game, server)
		}
	}
	select {}
}

func handleConnectionGame(conn *websocket.Conn, currentGame *game.Game, server *Server) {
	buf := make([]byte, 1024)
	conn.Write([]byte("Please pick a number between a row and a column (1-3) (1-3)"))

	const (
		WaitingForPosition = iota
		WaitingForAnswer
	)
	state := WaitingForPosition

	var (
		num1, num2  int
		question    questions.Question
		otherPlayer *websocket.Conn
	)

	regexVal := regexp.MustCompile(`^[1-3] [1-3]$`)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected gracefully ")
				game.RemoveUserFromGameConn(currentGame, conn)
				if len(currentGame.Conns) == 0 {
					// Possible race condition , needs mutex
					server.removeGameFromServer(currentGame)
				}
			} else {
				fmt.Println("Error reading message in game:", err)
			}
			return
		}

		// If it's not this player's turn, just notify and continue
		if currentGame.CurrentPlayer != conn {
			conn.Write([]byte("It's not your turn, wait!"))
			continue
		}

		message := buf[:n]

		switch state {
		case WaitingForPosition:
			if regexVal.Match(message) {
				otherPlayer, err = utils.FindTheOtherPlayer(currentGame, conn)
				if err != nil {
					fmt.Println("Error changing turn:", err)
					continue
				}

				num1, num2 = int(message[0]-'0'), int(message[2]-'0')
				fmt.Printf("Selected position: Row=%d, Col=%d\n", num1, num2)

				if currentGame.GameMatrix[num1-1][num2-1] != '.' {
					conn.Write([]byte("That position is already taken, please choose another."))
					continue
				}

				// Pick a random question
				question = questions.PickRandomQuestion(questionArray)
				conn.Write([]byte(fmt.Sprintf("Question is : %s\n A. %s\n B. %s\n C. %s\n D. %s\n ", question.Question, question.A, question.B, question.C, question.D)))

				// Switch to waiting for the answer
				state = 1
				continue
			} else {
				conn.Write([]byte("Please write a row and a column in the format: (1-3) (1-3)"))
			}

		case WaitingForAnswer:
			userAnswer := string(message)

			var fullCorrectAnswer string
			switch question.Answer {
			case "A":
				fullCorrectAnswer = question.A
			case "B":
				fullCorrectAnswer = question.B
			case "C":
				fullCorrectAnswer = question.C
			case "D":
				fullCorrectAnswer = question.D

			}

			ch := make(chan string)

			go api.OpenAIAPICall(question.Answer, fullCorrectAnswer, userAnswer, ch)

			chatGPTResponse := <-ch
			if chatGPTResponse == "false" {
				conn.Write([]byte("You answered WRONG!!!"))
				num1, num2 = utils.GiveRandomValidMatrixPosition(currentGame.EmptyPositions)
			} else if chatGPTResponse == "true" {
				conn.Write([]byte("You answered GOOD!"))
			} else {
				conn.Write([]byte("There was an error getting the real answer, but we will make it correct for you"))
			}

			var isWinner bool
			// Place the symbol depending on who is playing (Host = 'X', Guest = 'O')
			if currentGame.CurrentPlayer == currentGame.Host {
				currentGame.GameMatrix[num1-1][num2-1] = 'X'
				utils.RemoveSliceElement(&currentGame.EmptyPositions, []int{num1 - 1, num2 - 1})
				isWinner = game.IsMatrixWinner(currentGame.GameMatrix, num1-1, num2-1, 'X')
			} else {
				currentGame.GameMatrix[num1-1][num2-1] = 'O'
				utils.RemoveSliceElement(&currentGame.EmptyPositions, []int{num1 - 1, num2 - 1})
				isWinner = game.IsMatrixWinner(currentGame.GameMatrix, num1-1, num2-1, 'O')
			}

			game.PrintGameMaxtrixToPlayers(currentGame)

			if isWinner {
				game.WonGame(currentGame, currentGame.CurrentPlayer, otherPlayer)
			} else if !isWinner && len(currentGame.EmptyPositions) == 0 {
				game.DrawGame(currentGame)
			}
			// Switch turn to the other player
			currentGame.CurrentPlayer = otherPlayer
			// Prompt the next player for their move
			otherPlayer.Write([]byte("Please pick a number between a row and a column (1-3) (1-3)"))
			state = WaitingForPosition
		}
	}
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

		if string(msg) == "C" || string(msg) == "c" {
			ws.Write([]byte("Please write a lobby id for your friend to join"))
			game := game.NewGame()
			for {
				bytesRead, err := ws.Read(buf)
				if err != nil {
					break
				}
				urlName := buf[:bytesRead]
				game.Conns[ws] = true
				game.Url = string(urlName)
				game.Host = ws
				game.CurrentPlayer = ws

				server.Games = append(server.Games, game)

				ws.Write([]byte(fmt.Sprintf("You have created a server with the url %s", game.Url)))

				RunGame(game, false, server)
				return
			}
		} else if string(msg) == "J" {
			ws.Write([]byte("Please write the game url"))
			for {
				bytesRead, err := ws.Read(buf)
				if err != nil {
					break
				}
				var searchedGame *game.Game
				for _, v := range server.Games {
					if v.Url == string(buf[:bytesRead]) {
						searchedGame = v
						break
					}
				}
				if searchedGame == nil || len(searchedGame.Conns) >= 2 {
					ws.Write([]byte("The server does not exist or there are too many players"))
				} else {
					searchedGame.Conns[ws] = true
					ws.Write([]byte("Connected to the server!"))
					RunGame(searchedGame, true, server)
					return
				}
			}
		} else {
			ws.Write([]byte("Please type one of those characters"))
		}

	}
}

func main() {
	server := newServer()
	questionArray = questions.LoadQuestionsJSON()

	http.Handle("/ws", websocket.Handler(server.TicTacToe))
	http.ListenAndServe("127.0.0.1:3000", nil)
}
