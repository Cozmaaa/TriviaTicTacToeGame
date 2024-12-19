# Tic Tac Toe with Trivia

Welcome to **Tic Tac Toe with Trivia**, a fun twist on the classic Tic Tac Toe game! This version not only challenges your strategic skills but also tests your trivia knowledge along the way. Built with **Go**, it ensures smooth and efficient gameplay.

## Features

- **Classic Tic Tac Toe Gameplay**: Enjoy the timeless 3x3 grid battle.
- **Trivia Challenges**: Answer trivia questions correctly to earn your move.
- **Multiplayer**: Play online with a friend.
- **Customizable Trivia Categories**: Choose from a variety of topics for your trivia questions.

## Installation

To run this game, ensure you have Go installed on your machine. Follow these steps to get started:

1. Clone the repository:
   ```bash
   git clone https://github.com/Cozmaaa/TriviaTicTacToeGame.git
   ```
2. Create a `.env` file and insert your `OPENAI_API_KEY`.

3. Build the project:
   The project contains two files: `main.go` (server) and `client.go` (client).
   ```bash
   go build -o main.go
   go build -o client.go
   ```

4. Run the game:
   Start the server first, then run the client. Users will be able to join the game.
   ```bash
   ./main
   ./client
   ```

## How to Play

1. Start the game and choose to create or join a game:
   - `C`: Create a game.
   - `J`: Join a game.

2. The host will input a URL to create a lobby. The other player must join using the same URL.

3. Trivia questions will appear after you input your position choice:
   - Answer correctly to place your mark (X or O).
   - Incorrect answers will place the mark randomly.

4. Win the game by aligning three of your marks horizontally, vertically, or diagonally.

## Trivia Categories

You can customize the trivia questions by editing the provided JSON file (`multipleAnswer.json`). Add or modify categories and questions to make the game even more engaging!

## Requirements

- Go 1.18 or later
- Terminal or command-line interface


Have fun and test your wits with **Tic Tac Toe with Trivia**! 🎉
