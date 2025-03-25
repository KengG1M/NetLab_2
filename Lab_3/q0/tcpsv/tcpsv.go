package main

import (
	"encoding/base64"
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

func main() {
	// init array
	// var Users []User

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

func encrypt64(input string) string {
	// TODO
	encryptedPw := base64.StdEncoding.EncodeToString([]byte(input))
	return encryptedPw
}

func decrypt64(input string) (string, error) {
	//TODO

	pw, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(pw), nil
}

func checkAuthenticate(username string, password string) {

}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)

	// n, err := conn.Read(buffer)
	// if err != nil {
	// 	fmt.Println("Read error:", err)
	// 	return
	// }

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Read error:", err)
			return
		}

		message := strings.TrimSpace(string(buffer[:n]))
		fmt.Println("Received:", message)

		if message == "exit" {
			fmt.Println("Shutting down server...")
			conn.Write([]byte("Server is shutting down..."))
			os.Exit(0)
		}

		conn.Write([]byte("Hello from server"))

	}

	// messageAuth1 := strings.TrimSpace(string(buffer[:n]))
	// fmt.Println("Received out loop:", messageAuth1)

}
