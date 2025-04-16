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
	// Kết nối tới server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	defer conn.Close()

	serverReader := bufio.NewReader(conn)
	consoleReader := bufio.NewReader(os.Stdin)

	// Nhận prompt và nhập Username
	prompt, err := serverReader.ReadString(':')
	if err != nil {
		fmt.Println("Error reading username prompt:", err)
		return
	}
	fmt.Print(prompt)
	username, err := consoleReader.ReadString('\n')
	if err != nil {
		fmt.Println("Input error:", err)
		return
	}
	username = strings.TrimSpace(username)
	conn.Write([]byte(username + "\n"))

	// Nhận prompt và nhập Password
	prompt, err = serverReader.ReadString(':')
	if err != nil {
		fmt.Println("Error reading password prompt:", err)
		return
	}
	fmt.Print(prompt)
	password, err := consoleReader.ReadString('\n')
	if err != nil {
		fmt.Println("Input error:", err)
		return
	}
	password = strings.TrimSpace(password)
	conn.Write([]byte(password + "\n"))

	// Đọc phản hồi xác thực từ server (chứa key phiên)
	authResponse, err := serverReader.ReadString('\n')
	if err != nil {
		fmt.Println("Authentication error:", err)
		return
	}
	fmt.Print(authResponse)
	if !strings.Contains(authResponse, "key is") {
		return
	}

	// Trích xuất key từ phản hồi
	keyIndex := strings.Index(authResponse, "key is")
	if keyIndex == -1 {
		fmt.Println("Không thể trích xuất key")
		return
	}
	// Giả định định dạng: "Authentication successful. Your key is <key>"
	keyPart := authResponse[keyIndex:]
	fields := strings.Fields(keyPart)
	if len(fields) < 3 {
		fmt.Println("Key không hợp lệ")
		return
	}
	key := fields[2]

	// Goroutine đọc liên tục dữ liệu từ server
	go func() {
		for {
			msg, err := serverReader.ReadString('\n')
			if err != nil {
				fmt.Println("\nDisconnected from server:", err)
				os.Exit(0)
			}
			// In ra mọi thông báo từ server
			fmt.Print(msg)
		}
	}()

	// Goroutine gửi PING đến server mỗi 10s
	go func() {
		for {
			time.Sleep(10 * time.Second)
			_, err := fmt.Fprintf(conn, "%s_PING\n", key)
			if err != nil {
				return
			}
		}
	}()

	// Vòng lặp đọc input từ console và gửi cho server
	for {
		// Hỏi người chơi nhập đoán (chữ cái)
		fmt.Print("Your guess (enter a letter): ")
		guessInput, err := consoleReader.ReadString('\n')
		if err != nil {
			fmt.Println("Input error:", err)
			break
		}
		guessInput = strings.TrimSpace(guessInput)
		if guessInput == "" {
			continue
		}
		// Chỉ lấy ký tự đầu tiên và chuyển sang chữ in hoa
		guess := strings.ToUpper(string(guessInput[0]))
		// Gửi thông điệp theo định dạng: key_guess (ví dụ: "9676_A")
		msg := fmt.Sprintf("%s_%s\n", key, guess)
		_, err = fmt.Fprintf(conn, msg)
		if err != nil {
			fmt.Println("Send error:", err)
			break
		}
	}
}
