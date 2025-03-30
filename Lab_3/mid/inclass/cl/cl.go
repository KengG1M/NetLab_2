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

	msg, _ = server.ReadString('.')
	fmt.Println(msg)

	// tạo var state để nhận data từ sv sau khi check valid
	// 0 fail  1 success
	state, _ := server.ReadString('!')
	fmt.Println("state: ", state, "cham het")

	// for {
	// 	// fmt.Println("send msg: ")
	// 	// input, _ := reader.ReadString('\n')
	// 	// conn.Write([]byte(input))

	// 	// reply, _ := server.ReadString('\n')
	// 	// fmt.Println("sv: ", reply)
	// 	if state == "Success(st)." {
	// 		fmt.Println("send msg: ")
	// 		input, _ := reader.ReadString('\n')
	// 		conn.Write([]byte(input))

	// 		reply, _ := server.ReadString('\n')
	// 		fmt.Println("sv: ", reply)
	// 	} else {
	// 		os.Exit(0)
	// 	}
	// }

	if state == "Success(st)." {
		fmt.Println("abc")
	}
}
