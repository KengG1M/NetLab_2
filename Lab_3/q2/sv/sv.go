package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func handleFileDownload(conn net.Conn, validKey string) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Đọc yêu cầu từ client (VD: "125_file1.txt")
	request, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}

	request = strings.TrimSpace(request)

	// Kiểm tra prefix key (nếu dùng xác thực)
	parts := strings.SplitN(request, "_", 2)
	if len(parts) != 2 || parts[0] != validKey {
		conn.Write([]byte("Invalid key or format\n"))
		return
	}

	filename := parts[1]
	fmt.Println("Client requested file:", filename)

	// Đọc file
	content, err := os.ReadFile(filename)
	if err != nil {
		conn.Write([]byte("Error: File not found\n"))
		return
	}

	if len(content) > 10*1024*1024 {
		conn.Write([]byte("Error: File too large\n"))
		return
	}

	// Gửi nội dung file về client
	conn.Write(content)
	fmt.Println("File sent successfully.")
}
