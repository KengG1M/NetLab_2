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
	msg, _ := server.ReadString(':')
	fmt.Print(msg)
	username, _ := reader.ReadString('\n')
	conn.Write([]byte(username))

	msg, _ = server.ReadString(':')
	fmt.Print(msg)
	password, _ := reader.ReadString('\n')
	conn.Write([]byte(password))

	result, _ := server.ReadString('\n')
	fmt.Print(result)

	if !strings.Contains(result, "key") {
		return
	}

	// Game interaction loop
	for {
		msg, err := server.ReadString('\n')
		if err != nil {
			fmt.Println("Disconnected from server.")
			return
		}

		fmt.Print(msg)

		if strings.Contains(msg, "Your turn") {
			fmt.Print("Enter guess: ")
			input, _ := reader.ReadString('\n')
			conn.Write([]byte(input))
		}
	}
}
