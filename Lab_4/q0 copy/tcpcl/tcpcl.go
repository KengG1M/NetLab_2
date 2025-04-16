package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
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
	conn.Write([]byte(username))

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	conn.Write([]byte(password))

	// Read server response
	response, _ := server.ReadString('\n')
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
	key := response[keyStart : keyStart+keyEnd]

	// Game interaction loop
	// Modified message reading loop
	for {
		msg, err := server.ReadString('\n')
		if err != nil {
			fmt.Println("Disconnected:", err)
			return
		}

		// Print server messages except the prompt
		if !strings.HasPrefix(msg, "Guess a letter: ") {
			fmt.Print(msg)
		}

		if strings.Contains(msg, ">>> YOUR TURN!") {
			fmt.Print("Your guess: ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if len(input) > 0 {
				// Send only the first character in uppercase
				guess := strings.ToUpper(string(input[0]))
				fmt.Fprintf(conn, "%s_%s\n", key, guess)
			}
		}
	}
}
