package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server listening on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		go handleFileRequest(conn)
	}
}

func handleFileRequest(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Read file name
	filename, err := reader.ReadString('\n')
	if err != nil {
		conn.Write([]byte("Error reading filename\n"))
		return
	}
	filename = filename[:len(filename)-1] // Trim newline

	fmt.Println("Client requested file:", filename)

	// Read file
	content, err := os.ReadFile(filename)
	if err != nil {
		conn.Write([]byte("Error: File not found\n"))
		return
	}

	if len(content) > 10*1024*1024 {
		conn.Write([]byte("Error: File too large\n"))
		return
	}

	// Send READY and then content
	conn.Write([]byte("READY\n"))
	conn.Write(content)
	fmt.Println("File sent successfully.")
	os.Exit(0)
}
