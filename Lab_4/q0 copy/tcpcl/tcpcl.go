package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	server := bufio.NewReader(conn)

	// Login
	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	conn.Write([]byte(username + "\n"))

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	conn.Write([]byte(password + "\n"))

	// Read server response
	response, err := server.ReadString('\n')
	if err != nil {
		fmt.Println("Login error:", err)
		return
	}
	fmt.Print(response)

	if !strings.Contains(response, "key") {
		return
	}

	// Extract key from response
	keyStart := strings.Index(response, "key is ") + 7
	keyEnd := strings.Index(response[keyStart:], "\n")
	if keyEnd == -1 {
		keyEnd = len(response) - keyStart
	}
	key := strings.TrimSpace(response[keyStart : keyStart+keyEnd])

	// Game interaction loop
	go pingServer(conn, key) // Start keep-alive goroutine

	for {
		msg, err := server.ReadString('\n')
		if err != nil {
			fmt.Println("\nDisconnected from server:", err)
			return
		}

		msg = strings.TrimSpace(msg)
		fmt.Println(msg) // Print server message

		if strings.Contains(msg, ">>> YOUR TURN!") {
			handlePlayerTurn(conn, reader, key)
		}
	}
}

func handlePlayerTurn(conn net.Conn, reader *bufio.Reader, key string) {
	fmt.Print("Your guess: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Input error:", err)
		return
	}

	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return
	}

	// Send only first character as uppercase
	guess := strings.ToUpper(string(input[0]))
	_, err = fmt.Fprintf(conn, "%s_%s\n", key, guess)
	if err != nil {
		fmt.Println("Send error:", err)
		return
	}
}

func pingServer(conn net.Conn, key string) {
	for {
		time.Sleep(10 * time.Second)
		_, err := fmt.Fprintf(conn, "%s_PING\n", key)
		if err != nil {
			return
		}
	}
}
