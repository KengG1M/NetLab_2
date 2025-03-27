package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, _ := net.Dial("tcp", "localhost:8080")
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	server := bufio.NewReader(conn)

	// Nhập username
	msg, _ := server.ReadString(':')
	fmt.Print(msg)
	username, _ := reader.ReadString('\n')
	conn.Write([]byte(username))

	// Nhập password
	msg, _ = server.ReadString(':')
	fmt.Print(msg)
	password, _ := reader.ReadString('\n')
	conn.Write([]byte(password))

	// Nhận key từ server
	result, _ := server.ReadString('\n')
	fmt.Print(result)

	if !strings.Contains(result, "key") {
		return
	}

	key := strings.TrimSpace(strings.Split(result, ":")[1])

	for {
		fmt.Print("Send message (prefix will be added): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		full := key + "_" + input + "\n"
		conn.Write([]byte(full))
		reply, _ := server.ReadString('\n')
		fmt.Println("Server:", reply)
	}
}
