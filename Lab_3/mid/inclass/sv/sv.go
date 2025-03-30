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

var User []User

func main(){
	loadUsers("users.json")

	ln, err := net.Listen("tcp",":8080")
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

func loadUsers(filename string){

}

func checkAuthenticate

func handleConnection(conn net.Conn){

}
