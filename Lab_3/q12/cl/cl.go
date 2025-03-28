// =======================================
// ✅ CLIENT - main_client.go
// ✅ Hoàn thành tất cả yêu cầu trừ Câu 3 (file download)
// =======================================

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

	serverReader := bufio.NewReader(conn)
	inputReader := bufio.NewReader(os.Stdin)

	// Gửi username
	msg, _ := serverReader.ReadString(':')
	fmt.Print(msg)
	username, _ := inputReader.ReadString('\n')
	conn.Write([]byte(username))

	// Gửi password
	msg, _ = serverReader.ReadString(':')
	fmt.Print(msg)
	password, _ := inputReader.ReadString('\n')
	conn.Write([]byte(password))

	// Nhận key
	authMsg, _ := serverReader.ReadString('\n')
	fmt.Print(authMsg)
	if !strings.Contains(authMsg, "key") {
		return
	}
	key := strings.TrimSpace(strings.Split(authMsg, ":")[1])

	// Chơi game đoán số
	for {
		prompt, _ := serverReader.ReadString('\n')
		fmt.Print(prompt)
		guess, _ := inputReader.ReadString('\n')
		guess = strings.TrimSpace(guess)
		message := key + "_" + guess + "\n"
		conn.Write([]byte(message))
		response, _ := serverReader.ReadString('\n')
		fmt.Print("Server: ", response)
		if strings.Contains(response, "Correct") {
			break
		}
	}
}
