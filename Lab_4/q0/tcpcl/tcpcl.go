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
	"sync"
	"time"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Fullname string `json:"fullname"`
	Emails   string `json:"email"`
	Address  string `json:"address"`
}

type Word struct {
	Text string
	Hint string
}

type GameSession struct {
	Word       string
	Hint       string
	Guessed    map[rune]bool
	PlayerKeys []string
	TurnIndex  int
	Scores     map[string]int
}

var Users []User
var keyMap = make(map[string]net.Conn)
var mu sync.Mutex

func main() {
	loadUsers("users.json")
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Server is listening on port 8080...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func loadUsers(file string) {
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("No user file found. Starting fresh")
		return
	}
	json.Unmarshal(data, &Users)
}

func checkAuthenticate(username, encrypted string) bool {
	for _, u := range Users {
		if u.Username == username && u.Password == encrypted {
			return true
		}
	}
	return false
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	conn.Write([]byte("Username: "))
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	conn.Write([]byte("Password: "))
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	encrypted := base64.StdEncoding.EncodeToString([]byte(password))

	if !checkAuthenticate(username, encrypted) {
		conn.Write([]byte("Failed\n"))
		return
	}

	key := fmt.Sprintf("%d", rand.Intn(1000)+100)
	mu.Lock()
	keyMap[key] = conn
	mu.Unlock()

	conn.Write([]byte("Auth success. Your key is " + key + "\n"))

	if len(keyMap) == 2 {
		go startGame()
	}

	for {
		time.Sleep(time.Second * 60)
	}
}

func startGame() {
	words := []Word{
		{"banana", "A tropical fruit"},
		{"python", "A programming language"},
		{"vietnam", "A country in Asia"},
	}
	rand.Seed(time.Now().UnixNano())
	selected := words[rand.Intn(len(words))]

	session := GameSession{
		Word:       selected.Text,
		Hint:       selected.Hint,
		Guessed:    make(map[rune]bool),
		PlayerKeys: getPlayerKeys(),
		Scores:     make(map[string]int),
	}

	for _, key := range session.PlayerKeys {
		keyMap[key].Write([]byte("Game started! Hint: " + session.Hint + "\n"))
		keyMap[key].Write([]byte("Word: " + displayWord(session.Word, session.Guessed) + "\n"))
	}

	for {
		currentKey := session.PlayerKeys[session.TurnIndex]
		conn := keyMap[currentKey]
		conn.Write([]byte("Your turn! Guess a letter or the full word:\n"))

		reader := bufio.NewReader(conn)
		guess, _ := reader.ReadString('\n')
		guess = strings.ToLower(strings.TrimSpace(guess))

		if len(guess) > 1 {
			if guess == session.Word {
				broadcast("Player "+currentKey+" wins! The word was: "+session.Word, session.PlayerKeys)
				return
			} else {
				conn.Write([]byte("Wrong guess. Turn passed.\n"))
				session.TurnIndex = (session.TurnIndex + 1) % len(session.PlayerKeys)
				continue
			}
		}

		letter := rune(guess[0])
		if session.Guessed[letter] {
			conn.Write([]byte("Already guessed.\n"))
			continue
		}
		session.Guessed[letter] = true
		count := strings.Count(session.Word, string(letter))
		if count > 0 {
			session.Scores[currentKey] += count * 10
			conn.Write([]byte(fmt.Sprintf("Correct! '%c' appears %d times. Score: %d\n", letter, count, session.Scores[currentKey])))
			if allGuessed(session.Word, session.Guessed) {
				broadcast("Player "+currentKey+" has revealed the word! It was: "+session.Word, session.PlayerKeys)
				return
			}
			// continue same player's turn
		} else {
			conn.Write([]byte("Wrong letter. Turn passed.\n"))
			session.TurnIndex = (session.TurnIndex + 1) % len(session.PlayerKeys)
		}

		for _, key := range session.PlayerKeys {
			keyMap[key].Write([]byte("Word: " + displayWord(session.Word, session.Guessed) + "\n"))
		}
	}
}

func displayWord(word string, guessed map[rune]bool) string {
	var display strings.Builder
	for _, c := range word {
		if guessed[c] {
			display.WriteRune(c)
			display.WriteRune(' ')
		} else {
			display.WriteString("_ ")
		}
	}
	return display.String()
}

func allGuessed(word string, guessed map[rune]bool) bool {
	for _, c := range word {
		if !guessed[c] {
			return false
		}
	}
	return true
}

func getPlayerKeys() []string {
	mu.Lock()
	defer mu.Unlock()
	var keys []string
	for k := range keyMap {
		keys = append(keys, k)
	}
	return keys
}

func broadcast(message string, keys []string) {
	for _, key := range keys {
		keyMap[key].Write([]byte(message + "\n"))
	}
}
