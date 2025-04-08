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
	Username string `json:"username"`
	Password string `json:"password"`
	Fullname string `json:"fullname"`
	Emails   string `json:"email"`
	Address  string `json:"address"`
}

var Users []User
var keyMap = make(map[string]net.Conn)

func main() {
	// init array

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

// func encrypt64(input string) string {
// 	// TODO
// 	encryptedPw := base64.StdEncoding.EncodeToString([]byte(input))
// 	return encryptedPw
// }

// func decrypt64(input string) (string, error) {
// 	//TODO

// 	pw, err := base64.StdEncoding.DecodeString(input)
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(pw), nil
// }

func checkAuthenticate(username, encrypted string) bool {
	for _, u := range Users {
		if u.Username == username && u.Password == encrypted {
			return true
		}
	}
	return false
}

func loadUsers(file string) {
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("No user file found. Starting fresh")
		return
	}
	json.Unmarshal(data, &Users)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	conn.Write([]byte("Username: "))
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	conn.Write([]byte("Password: "))
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	encrypted := base64.StdEncoding.EncodeToString([]byte(password))

	// authenticate
	if !checkAuthenticate(username, encrypted) {
		conn.Write([]byte("Failed\n"))
		return
	}

	rand.Seed(time.Now().UnixNano())
	key := fmt.Sprintf("%d", rand.Intn(1000)+100)
	keyMap[key] = conn

	conn.Write([]byte("Auth success. Your key is " + key + "\n"))

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		msg = strings.TrimSpace(msg)
		if strings.HasPrefix(msg, key+"_") {
			fmt.Println("[", key, "]:", msg)
			conn.Write([]byte("Server received: " + msg + "\n"))
		} else {
			conn.Write([]byte("Invalid prefix or key\n"))
		}
	}
}

type Word struct {
	text string
	hint string
}

func loadWord() {

}

func randomWord() {

}
