package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connect: ", err)
		return
	}
	defer conn.Close()

	// input exit
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter message to send (type'exit' if you want to quit this facking program): ")
	text, _ := reader.ReadString('\n')

	// Send data to server
	conn.Write([]byte(text))

	// Receive response from sv
	buffer := make([]byte, 1024)

	n, _ := conn.Read(buffer)
	fmt.Println("Server reply:", string(buffer[:n]))
}
