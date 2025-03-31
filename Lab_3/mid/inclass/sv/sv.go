package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Fullname string `json:"fullname"`
	Emails   string `json:"email"`
	Address  string `json:"address"`
}

var Users []User

func main() {
	loadUsers("users.json")

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

func loadUsers(file string) {
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("No user file found. Starting fresh")
		return
	}
	json.Unmarshal(data, &Users)
}

func isValid(username, encrypted string) bool {
	for _, u := range Users {
		if u.Username == username && u.Password == encrypted {
			return true
		}
	}
	return false
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	const maxAttempts = 3

	for i := 1; i <= maxAttempts; i++ {
		// Hỏi username
		conn.Write([]byte("input username:\n")) // Gửi prompt có "\n"
		username, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading username:", err)
			return
		}
		username = strings.TrimSpace(username)

		// Hỏi password
		conn.Write([]byte("input pass:\n"))
		pw, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading password:", err)
			return
		}
		pw = strings.TrimSpace(pw)

		// Mã hoá password
		encrypted := base64.StdEncoding.EncodeToString([]byte(pw))

		// Kiểm tra đăng nhập
		if isValid(username, encrypted) {
			conn.Write([]byte("Success(st).\n"))
			fmt.Println("Login success:", username)
			return
		} else {
			attemptsLeft := maxAttempts - i
			if attemptsLeft > 0 {
				// Sai nhưng còn lượt
				conn.Write([]byte(
					fmt.Sprintf("Wrong credentials. %d attempt(s) left.\n", attemptsLeft),
				))
			} else {
				// Sai đủ 3 lần
				conn.Write([]byte("Login failed.\n"))
				fmt.Println("Too many failed attempts. Connection closed.")
				return
			}
		}
	}
}
