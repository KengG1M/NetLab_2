package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server is listening on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		// go func() {
		// 	defer conn.Close()
		// 	reader := bufio.NewReader(conn)
		// 	peek, _ := reader.Peek(20) // xem trước vài byte
		// 	if strings.Contains(string(peek), "_") {
		// 		handleFileDownload(conn, "123")
		// 	} else {
		// 		handleConnection(conn)
		// 	}
		// }()
		go handleClient(conn, "123")
	}
}

// func handleConnection(conn net.Conn) {
// 	defer conn.Close()

// 	buffer := make([]byte, 1024)

// 	n, err := conn.Read(buffer)
// 	if err != nil {
// 		fmt.Println("Read error: ", err)
// 		return
// 	}

// 	message := strings.TrimSpace(string(buffer[:n]))
// 	fmt.Println("Received msg: ", message)

// 	if message == "exit" {
// 		fmt.Println("Shutting down sv")
// 		conn.Write([]byte("Sv is shutting down"))
// 		os.Exit(0)
// 	}
// }

// func handleFileDownload(conn net.Conn, validKey string) {
// 	defer conn.Close()
// 	reader := bufio.NewReader(conn)

// 	request, err := reader.ReadString('\n')
// 	if err != nil {
// 		fmt.Println("Error reading request:", err)
// 		return
// 	}
// 	request = strings.TrimSpace(request)

// 	parts := strings.SplitN(request, "_", 2)
// 	if len(parts) != 2 || parts[0] != validKey {
// 		conn.Write([]byte("Invalid key or format\n"))
// 		return
// 	}

// 	filename := parts[1]
// 	fmt.Println("Client requested file:", filename)

// 	content, err := os.ReadFile(filename)
// 	if err != nil {
// 		conn.Write([]byte("Error: File not found\n"))
// 		return
// 	}

// 	if len(content) > 10*1024*1024 {
// 		conn.Write([]byte("Error: File too large\n"))
// 		return
// 	}

// 	conn.Write([]byte("READY\n")) // Gửi tín hiệu sẵn sàng
// 	conn.Write(content)
// 	fmt.Println("File sent successfully.")
// }

func handleClient(conn net.Conn, validKey string) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected.")
			return
		}
		input = strings.TrimSpace(input)

		if input == "exit" {
			fmt.Println("Shutting down server...")
			conn.Write([]byte("Server is shutting down...\n"))
			os.Exit(0)
		}

		// Tách thành key và phần còn lại
		parts := strings.SplitN(input, "_", 2)
		if len(parts) != 2 || parts[0] != validKey {
			conn.Write([]byte("Invalid key or format\n"))
			continue
		}

		payload := parts[1]

		if strings.HasSuffix(payload, ".txt") {
			// Đây là yêu cầu tải file
			sendFile(conn, payload)
		} else {
			// Đây là message bình thường
			fmt.Println("[", validKey, "]:", payload)
			conn.Write([]byte("Server received: " + payload + "\n"))
		}
	}

}

func sendFile(conn net.Conn, filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		conn.Write([]byte("Error: File not found\n"))
		return
	}

	if len(content) > 10*1024*1024 {
		conn.Write([]byte("Error: File too large\n"))
		return
	}

	conn.Write([]byte("READY\n")) // Thông báo sẵn sàng
	conn.Write(content)
	fmt.Println("File", filename, "sent successfully.")
}
