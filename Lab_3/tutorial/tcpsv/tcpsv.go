package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Server is listening on port 8080...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}

	message := strings.TrimSpace(string(buffer[:n]))
	fmt.Println("Received:", message)

	if message == "exit" {
		fmt.Println("Shutting down server...")
		conn.Write([]byte("Server is shutting down..."))
		os.Exit(0)
	}

	conn.Write([]byte("Hello from server"))
}
