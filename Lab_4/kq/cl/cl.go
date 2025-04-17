package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:138")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// authentication for the 1st time
	fmt.Println("Enter username:")
	var username string
	fmt.Scanln(&username)
	conn.Write([]byte(username))

	fmt.Println("Enter password:")
	var password string
	fmt.Scanln(&password)
	conn.Write([]byte(password))

	received := make([]byte, 1024)
	n, err := conn.Read(received)
	if err != nil {
		fmt.Println("Error reading data:", err)
		return
	}
	fmt.Print(string(received[:n])) // print authentication message

	// loop if authentication is incorrect
	for string(received[:n]) == "Invalid username or password.\n" {
		fmt.Println("Enter username:")
		fmt.Scanln(&username)
		conn.Write([]byte(username))

		fmt.Println("Enter password:")
		fmt.Scanln(&password)
		conn.Write([]byte(password))

		n, err = conn.Read(received)
		if err != nil {
			fmt.Println("Error reading data:", err)
			return
		}
		fmt.Print(string(received[:n]))
	}

	// receive PlayerID
	/*n, err = conn.Read(received)
	if err != nil {
		fmt.Println("Error reading data:", err)
		return
	}
	fmt.Print(string(received[:n])) */

	// send signal to start the game
	conn.Write([]byte("Ready!\n"))

	// receive starting signal
	n, err = conn.Read(received)
	if err != nil {
		fmt.Println("Error reading data:", err)
		return
	}
	fmt.Print(string(received[:n]))

	for string(received[:n]) == "Starts!\n" {
		reader := bufio.NewReader(os.Stdin)

		// receive word and description from the server
		data, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error receiving data from server:", err)
			return
		}

		parts := strings.Split(data, "|")
		word := parts[0] + "\n"
		desc := parts[1]

		fmt.Println("Welcome to the Hangman Game!")
		fmt.Println("Description: ", desc)

		// send signal to start guessing
		conn.Write([]byte("Guess!\n"))

		// receive original reveal string
		reveal, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error receiving data from server:", err)
			return
		}
		guess := ""

		for {
			fmt.Println("Word to guess: ", reveal)
			if reveal == word {
				break
			}
			fmt.Print("Enter your guess (a single letter): ")
			guess, _ = reader.ReadString('\n')
			guess = strings.TrimSpace(guess)

			_, err = conn.Write([]byte(guess + "\n"))
			if err != nil {
				fmt.Println("Error sending guess to server:", err)
				return
			}

			response, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println("Error receiving response from server:", err)
				return
			}

			fmt.Println(response)

			if strings.Contains(response, "Correct guess!") {
				// receive continued reveal string
				reveal, err = bufio.NewReader(conn).ReadString('\n')
				if err != nil {
					fmt.Println("Error receiving data from server:", err)
					return
				}
			}

			if strings.Contains(response, "Congratulations") {
				break
			}
		}
		return
	}
}
