package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, _ := net.Dial("tcp", "localhost:8080")

	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	server := bufio.NewReader(conn)

	msg, _ := server.ReadString(':')
	fmt.Println(msg)
	username, _ := reader.ReadString('\n')
	conn.Write([]byte(username))

	msg, _ = server.ReadString(':')
	fmt.Println(msg)
	pw, _ := reader.ReadString('\n')
	conn.Write([]byte(pw))

	for {
		fmt.Println("send msg: ")
		input, _ := reader.ReadString('\n')
		conn.Write([]byte(input))

		reply, _ := server.ReadString('\n')
		fmt.Println("sv: ", reply)
	}
}
