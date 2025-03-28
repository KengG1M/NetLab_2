// =======================================
// ✅ SERVER - main_server.go
// ✅ Hoàn thành tất cả yêu cầu trừ Câu 3 (file download)
// =======================================

package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	FullName  string   `json:"full_name"`
	Emails    []string `json:"emails"`
	Addresses []string `json:"address"`
}

var users []User
var activeKeys = make(map[string]net.Conn)

func main() {
	loadUsers("users.json")
	listener, _ := net.Listen("tcp", ":8080")
	defer listener.Close()
	fmt.Println("Server is running on port 8080...")

	for {
		conn, _ := listener.Accept()
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	conn.Write([]byte("Username: "))
	username, _ := r.ReadString('\n')
	username = strings.TrimSpace(username)

	conn.Write([]byte("Password: "))
	password, _ := r.ReadString('\n')
	password = strings.TrimSpace(password)
	encrypted := base64.StdEncoding.EncodeToString([]byte(password))

	if !authenticate(username, encrypted) {
		conn.Write([]byte("Authentication failed\n"))
		return
	}

	// Generate unique key
	rand.Seed(time.Now().UnixNano())
	key := fmt.Sprintf("%d", rand.Intn(900)+100)
	activeKeys[key] = conn
	conn.Write([]byte("Authenticated. Your key is: " + key + "\n"))

	// Start guessing game
	runGuessingGame(r, conn, key)
}

func runGuessingGame(r *bufio.Reader, conn net.Conn, key string) {
	target := rand.Intn(100) + 1
	for {
		conn.Write([]byte("Guess a number (1–100): \n"))
		input, _ := r.ReadString('\n')
		input = strings.TrimSpace(input)

		parts := strings.SplitN(input, "_", 2)
		if len(parts) != 2 || parts[0] != key {
			conn.Write([]byte("Invalid key or format\n"))
			continue
		}
		guessStr := parts[1]
		guess, err := strconv.Atoi(guessStr)
		if err != nil {
			conn.Write([]byte("Invalid number\n"))
			continue
		}

		if guess < target {
			conn.Write([]byte(key + "_Too low\n"))
		} else if guess > target {
			conn.Write([]byte(key + "_Too high\n"))
		} else {
			conn.Write([]byte(key + "_Correct! You win!\n"))
			break
		}
	}
}

func authenticate(username, password string) bool {
	for _, user := range users {
		if user.Username == username && user.Password == password {
			return true
		}
	}
	return false
}

func loadUsers(filename string) {
	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("No user file found.")
		return
	}
	json.Unmarshal(file, &users)
}
