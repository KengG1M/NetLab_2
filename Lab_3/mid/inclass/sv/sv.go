package main

import (
	"bufio"
	"fmt"
	"net"
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

}

func isValid(username string, encrypted string) bool {
	for _, u := range Users {
		if username == u.Username && encrypted == u.Password {
			return true
		}
	}
	return false
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	conn.Write([]byte("input username"))
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	conn.Write([]byte("input pass"))
	pw, _ := reader.ReadString('\n')
	pw = strings.TrimSpace(pw)

	if !isValid(username, pw) {
		fmt.Println("Failed!")
		return
	}

	fmt.Println("Success")

}
