package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"

	// "math/rand"
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

// var <Ten mang> []<type>
var Users []User

func main() {
	loadUsers("users.json")

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	defer ln.Close()
	fmt.Println("Server is listening on port 8080...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept err:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func loadUsers(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("No user file found. starting fresh")
		return
	}
	json.Unmarshal(data, &Users)

}

func isValid(username string, encrypted string) bool {
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

	conn.Write([]byte("input username:"))
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	conn.Write([]byte("input pass:"))
	pw, _ := reader.ReadString('\n')
	pw = strings.TrimSpace(pw)
	encrypted := base64.StdEncoding.EncodeToString([]byte(pw))

	if !isValid(username, encrypted) {
		fmt.Println("Failed!")
		conn.Write([]byte("Failed(sv)."))
		return
	}
	state := "Success(st)!"
	conn.Write([]byte(state))

	fmt.Println("Success")

}

func live(attempts int) {

}
