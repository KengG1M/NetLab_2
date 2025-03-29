package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, _ := net.Dial("tcp", "localhost:8080")
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	server := bufio.NewReader(conn)

	// Nhập username
	// reader server ở đây nó sẽ truyền data từ sv ->client cụ thể là những msg như "Username: "
	// Thì đó là ý khi mà đọc data từ sv về client
	msg, _ := server.ReadString(':') // đọc data từ sv: "Username: "
	fmt.Print(msg)                   // In ra Username: lên màn hình để user input
	username, _ := reader.ReadString('\n')
	conn.Write([]byte(username))

	// Nhập password
	msg, _ = server.ReadString(':')
	fmt.Print(msg)
	password, _ := reader.ReadString('\n')
	conn.Write([]byte(password))

	// Nhận key từ server
	result, _ := server.ReadString('\n')
	fmt.Print(result)

	if !strings.Contains(result, "key") {
		return
	}

	words := strings.Fields(result)
	key := words[len(words)-1] // lấy phần tử cuối là key

	fmt.Println("Extracted key:", key)

	for {
		fmt.Print("Send message (prefix will be added): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		full := key + "_" + input + "\n"
		conn.Write([]byte(full))
		reply, _ := server.ReadString('\n')
		fmt.Println("Server:", reply)
	}
}
