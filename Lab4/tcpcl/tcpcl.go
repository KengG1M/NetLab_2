package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
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

	var key string
	authDone := false

	// Listen to server messages
	go func() {
		for serverScanner.Scan() {
			text := serverScanner.Text()
			fmt.Println(text)

			if strings.HasPrefix(text, "Authenticated! Your key is: ") {
				parts := strings.Split(text, ": ")
				if len(parts) == 2 {
					key = strings.TrimSpace(parts[1])
					authDone = true
				}
			}
		}
	}()

	for inputScanner.Scan() {
		text := inputScanner.Text()
		if authDone && key != "" {
			fmt.Fprintf(conn, "%s_%s\n", key, text)
		} else {

			fmt.Fprintln(conn, text)
		}
	}
}
