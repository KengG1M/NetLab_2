package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connect: ", err)
		return
	}
	defer conn.Close()

	conn.Write([]byte("Hello from client"))

	buffer := make([]byte, 1024)

	n, _ := conn.Read(buffer)
	fmt.Println("Server reply:", string(buffer[:n]))
}
