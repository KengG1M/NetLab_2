package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:12345")
	if err != nil {
		fmt.Println("Connection failed:", err)
		return
	}
	defer conn.Close()

	serverScanner := bufio.NewScanner(conn)
	inputScanner := bufio.NewScanner(os.Stdin)

	go func() {
		for serverScanner.Scan() {
			text := serverScanner.Text()
			fmt.Println(text)
		}
	}()

	for inputScanner.Scan() {
		text := inputScanner.Text()
		fmt.Fprintln(conn, text)
	}
}
