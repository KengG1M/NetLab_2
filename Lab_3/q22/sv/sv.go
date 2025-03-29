package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

type User struct {
	Username string `json:"username "`
	Password string `json:"password"`
	Fullname string `json:"fullname"`
	Emails   string `json:"email"`
	Address  string `json:"address"`
}

var Users []User
var keyMap = make(map[string]net.Conn)

func main() {

	loadUsers("users.json")

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
		// go handleFileRequest(conn)
		go handleClient(conn)
	}
}

func loadUsers(file string) {
	data, err := os.ReadFile(file)

	if err != nil {
		fmt.Println("No user found. Starting refresh")
		return
	}

	json.Unmarshal(data, &Users)
}

func checkAuthenticate(username string, encryptedpw string) bool {
	for _, u := range Users {
		if u.Username == username && u.Password == encryptedpw {
			return true
		}
	}
	return false
}

func handleClient(conn net.Conn) {
	// conn.Close() Close tcp connection when not neccessarily
	// defer sẽ hoãn thực thi conn.Close() cho đến khi hàm hiện tại kết thúc
	defer conn.Close()

	reader := bufio.NewReader(conn)

	conn.Write([]byte("Username: "))
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	conn.Write([]byte("Password: "))
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	encryptedpw := base64.StdEncoding.EncodeToString([]byte(password))

	// Check ko đúng sẽ return dừng chương trình
	if !checkAuthenticate(username, encryptedpw) {
		conn.Write([]byte("Failed\n"))
		return
	}

	rand.Seed(time.Now().UnixNano())
	key := fmt.Sprintf("%d", rand.Intn(100)+100)

	// Gán kết nối conn cho key user tương ứng
	// vd: conn#1 cho key 123; conn#2 cho key 456
	keyMap[key] = conn
	conn.Write([]byte("Login succes. Your key is " + key + "\n"))

	// Download file
	filename, err := reader.ReadString('\n')

	// Nếu có lỗi thì sẽ dừng chương trình
	if err != nil {
		conn.Write([]byte("Error reading filename\n"))
		return
	}

	filename = filename[:len(filename)-1] //Trim newline
	fmt.Println("Client request file: ", filename)

	content, err := os.ReadFile(filename)
	if err != nil {
		conn.Write([]byte("Error: 404 file not found\n"))
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

// func handleFileRequest(conn net.Conn) {
// 	defer conn.Close()
// 	reader := bufio.NewReader(conn)

// 	// Read file name
// 	filename, err := reader.ReadString('\n')
// 	if err != nil {
// 		conn.Write([]byte("Error reading filename\n"))
// 		return
// 	}
// 	filename = filename[:len(filename)-1] // Trim newline

// 	fmt.Println("Client requested file:", filename)

// 	// Read file
// 	content, err := os.ReadFile(filename)
// 	if err != nil {
// 		conn.Write([]byte("Error: File not found\n"))
// 		return
// 	}

// 	if len(content) > 10*1024*1024 {
// 		conn.Write([]byte("Error: File too large\n"))
// 		return
// 	}

// 	// Send READY and then content
// 	conn.Write([]byte("READY\n"))
// 	conn.Write(content)
// 	fmt.Println("File sent successfully.")
// 	os.Exit(0)
// }
