package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"strings"
	"time"
)

var words = map[string]string{
	"golang":        "A programming language developed by Google",
	"hangman":       "A word guessing game",
	"developer":     "One who writes code",
	"computer":      "An electronic device",
	"dockercompose": "A tool for defining and running multi-container applications.",
}

type User struct {
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Fullname  string   `json:"fullname"`
	Emails    []string `json:"emails"`
	Addresses []string `json:"addresses"`
}

type Player struct {
	PlayerID int
	Score    int
}

func getRandomWord() (string, string) {
	index := rand.Intn(len(words))
	i := 0
	for word, desc := range words {
		if i == index {
			return word, desc
		}
		i++
	}
	return "", ""
}

func main() {
	ln, err := net.Listen("tcp", ":138")
	if err != nil {
		fmt.Println("Error starting the server:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Server started. Waiting for players...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	user_list, err := ioutil.ReadFile("user.json")
	if err != nil {
		fmt.Println(err.Error())
	}

	var users []User
	err2 := json.Unmarshal(user_list, &users)
	if err2 != nil {
		fmt.Println("Error JSON Unmarshalling")
		fmt.Println(err2.Error())
	}
	defer conn.Close()

	// count players
	count := 0
	//var players []Player

	// authentication
	authenticated := false
	for !authenticated {
		username, err := readString(conn)
		if err != nil {
			fmt.Println("Error reading username:", err)
			break
		}
		password, err := readString(conn)
		if err != nil {
			fmt.Println("Error reading password:", err)
			break
		}

		for _, user := range users {
			if user.Username == username && user.Password == password {
				authenticated = true
			}
		}
		if authenticated == true {
			count++
			conn.Write([]byte("Authentication successful!\n"))
			break
		} else {
			conn.Write([]byte("Invalid username or password.\n"))
		}
	}

	//players[count].PlayerID = count
	//players[count].Score = 0

	// send PlayerID
	//conn.Write([]byte("You are player " + fmt.Sprint(count) + "\n"))

	// receive starting signal
	ready, err := readString(conn)
	if err != nil {
		fmt.Println("Error reading username:", err)
	}
	fmt.Print(string(ready))

	// send signal to start the game
	conn.Write([]byte("Starts!\n"))

	for string(ready) == "Ready!\n" {
		word, desc := getRandomWord()
		data := word + "|" + desc + "\n"
		conn.Write([]byte(data))

		// game logic goes here
		guessedLetters := make(map[string]bool)
		revealedWord := strings.Repeat("_", len(word))
		// receive start guessing signal
		ready, err := readString(conn)
		if err != nil {
			fmt.Println("Error reading username:", err)
		}
		if string(ready) == "Guess!\n" {
			conn.Write([]byte(revealedWord + "\n"))
		}
		//score := 0

		for {
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Println("Player disconnected")
				return
			}

			guess := string(buf[:n])
			guess = strings.TrimSpace(guess)

			if guessedLetters[guess] {
				conn.Write([]byte("Letter already guessed. Try again.\n"))
				continue
			}

			guessedLetters[guess] = true
			if strings.Contains(word, guess) {
				for i, letter := range word {
					if string(letter) == guess {
						revealedWord = revealedWord[:i] + guess + revealedWord[i+1:]
						//score += 10 // increment score for correct guess
					}
				}
				if word != revealedWord {
					conn.Write([]byte("Correct guess!\n"))
					conn.Write([]byte(revealedWord + "\n"))
				}
			} else {
				conn.Write([]byte("Incorrect guess. Try again.\n"))
			}

			if revealedWord == word {
				conn.Write([]byte("Congratulations! You guessed the word: " + word + "\n"))
				//conn.Write([]byte("Your score: " + fmt.Sprint(score) + "\n"))
				return
			}
		}
	}
}

func readString(conn net.Conn) (string, error) {
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	return string(buffer[:n]), err
}
