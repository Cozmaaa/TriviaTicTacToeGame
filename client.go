package main

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/net/websocket"
)

type Client struct {
	server *websocket.Conn
}

func joinWS(ipAddress, origin string) *Client {
	ws, err := websocket.Dial(ipAddress, "", origin)
	if err != nil {
		fmt.Println("Error connecting to the server", err)
		return nil
	}
	return &Client{
		server: ws,
	}
}

func (c *Client) handleWS(ws *websocket.Conn) {
	buf := make([]byte, 1024)

	go func() {
		for {
			n, err := ws.Read(buf)
			if err != nil {
				fmt.Println("Error reading message", err)
				return
			}
			msg := buf[:n]

			fmt.Println(string(msg))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		ws.Write([]byte(msg))
	}
}

func main() {
	fmt.Println("Please input the ip to the connection")
	// scanner := bufio.NewScanner(os.Stdin)
	var ipAddress string
	/*if scanner.Scan() {
		ipAddress = scanner.Text()
	}*/

	// ipAddress = "wss://04l4z8bl-3000.euw.devtunnels.ms/ws"
	// https://04l4z8bl-3000.euw.devtunnels.ms/
	// wsConn := joinWS(ipAddress, "https://04l4z8bl-3000.euw.devtunnels.ms/")

	ipAddress = "ws://localhost:3000/ws"
	wsConn := joinWS(ipAddress, "http://localhost/")

	defer wsConn.server.Close()
	wsConn.handleWS(wsConn.server)
}
