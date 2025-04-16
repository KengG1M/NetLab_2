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
	Username string   `json:"username"`
	Password string   `json:"password"`
	Fullname string   `json:"fullname"`
	Emails   []string `json:"emails"`
	Address  []string `json:"address"`
}

type Word struct {
	Text string `json:"text"`
	Hint string `json:"hint"`
}

type GameState struct {
	Word          string
	Hint          string
	Revealed      []bool
	Players       []net.Conn
	CurrentPlayer int
	Scores        map[string]int
	GameStarted   bool
	LastActivity  time.Time
	Mutex         sync.Mutex
}

var (
	Users    []User
	Words    []Word
	keyMap   = make(map[string]string)   // key -> username
	connMap  = make(map[string]net.Conn) // username -> conn
	game     *GameState
	gameLock sync.Mutex
)

func main() {
	loadUsers("users.json")
	loadWords("words.json")

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Server is listening on port 8080...")

	go gameTimer()

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
	err = json.Unmarshal(data, &Users)
	if err != nil {
		fmt.Println("Error loading users:", err)
	}
}

func loadWords(file string) {
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("No words file found. Using default words")
		Words = []Word{
			{Text: "hangman", Hint: "Popular word game"},
			{Text: "golang", Hint: "Programming language"},
		}
		return
	}
	err = json.Unmarshal(data, &Words)
	if err != nil {
		fmt.Println("Error loading words:", err)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Authentication
	conn.Write([]byte("Username: "))
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	conn.Write([]byte("Password: "))
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	encrypted := base64.StdEncoding.EncodeToString([]byte(password))

	if !checkAuthenticate(username, encrypted) {
		conn.Write([]byte("Authentication failed\n"))
		return
	}

	rand.Seed(time.Now().UnixNano())
	key := fmt.Sprintf("%d", rand.Intn(9000)+1000)
	keyMap[key] = username
	connMap[username] = conn

	conn.Write([]byte(fmt.Sprintf("Authentication successful. Your key is %s\n", key)))

	// Add player to game
	gameLock.Lock()
	if game == nil {
		initGame()
	}
	game.Players = append(game.Players, conn)
	game.Scores[username] = 0
	gameLock.Unlock()

	conn.Write([]byte(fmt.Sprintf("Welcome to Hangman! Waiting for players...\n")))

	// Start game if we have at least 2 players
	gameLock.Lock()
	if len(game.Players) >= 2 && !game.GameStarted {
		game.GameStarted = true
		broadcastGameState()
	}
	gameLock.Unlock()

	// Handle client messages
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			removePlayer(username)
			return
		}
		msg = strings.TrimSpace(msg)

		if strings.HasPrefix(msg, key+"_") {
			handleGameMessage(username, msg[len(key)+1:])
		} else {
			conn.Write([]byte("Invalid message format. Use key_message\n"))
		}
	}
}

func checkAuthenticate(username, encrypted string) bool {
	for _, u := range Users {
		if u.Username == username && u.Password == encrypted {
			return true
		}
	}
	return false
}

func initGame() {
	word := Words[rand.Intn(len(Words))]
	game = &GameState{
		Word:         strings.ToUpper(word.Text),
		Hint:         word.Hint,
		Revealed:     make([]bool, len(word.Text)),
		Players:      make([]net.Conn, 0),
		Scores:       make(map[string]int),
		GameStarted:  false,
		LastActivity: time.Now(),
	}
}

func broadcastGameState() {
	gameLock.Lock()
	defer gameLock.Unlock()

	game.LastActivity = time.Now()

	displayWord := ""
	for i, c := range game.Word {
		if game.Revealed[i] {
			displayWord += string(c) + " "
		} else {
			displayWord += "_ "
		}
	}

	scores := ""
	for username, score := range game.Scores {
		scores += fmt.Sprintf("%s: %d points\n", username, score)
	}

	for i, conn := range game.Players {
		conn.Write([]byte(fmt.Sprintf("\nHint: %s\n", game.Hint)))
		conn.Write([]byte(fmt.Sprintf("Word: %s\n", displayWord)))
		conn.Write([]byte(fmt.Sprintf("Scores:\n%s", scores)))
		if i == game.CurrentPlayer {
			conn.Write([]byte("It's your turn! Guess a letter: "))
		} else {
			conn.Write([]byte(fmt.Sprintf("Waiting for %s to guess...\n", keyMap[getKeyFromConn(conn)])))
		}
	}
}

func handleGameMessage(username, msg string) {
	gameLock.Lock()
	defer gameLock.Unlock()

	currentUsername := keyMap[getKeyFromConn(game.Players[game.CurrentPlayer])]
	if username != currentUsername {
		connMap[username].Write([]byte("It's not your turn!\n"))
		return
	}

	msg = strings.ToUpper(strings.TrimSpace(msg))
	if len(msg) != 1 || msg[0] < 'A' || msg[0] > 'Z' {
		connMap[username].Write([]byte("Please enter a single letter A-Z\n"))
		return
	}

	found := false
	count := 0
	for i, c := range game.Word {
		if string(c) == msg && !game.Revealed[i] {
			game.Revealed[i] = true
			found = true
			count++
		}
	}

	if found {
		game.Scores[username] += 10 * count
		connMap[username].Write([]byte(fmt.Sprintf("Correct! +%d points\n", 10*count)))
	} else {
		connMap[username].Write([]byte("Incorrect guess!\n"))
		game.CurrentPlayer = (game.CurrentPlayer + 1) % len(game.Players)
	}

	// Check if game is over
	gameOver := true
	for _, revealed := range game.Revealed {
		if !revealed {
			gameOver = false
			break
		}
	}

	if gameOver {
		broadcastGameOver()
		initGame()
		if len(game.Players) >= 2 {
			game.GameStarted = true
		}
	}

	broadcastGameState()
}

func broadcastGameOver() {
	winner := ""
	maxScore := -1
	for username, score := range game.Scores {
		if score > maxScore {
			maxScore = score
			winner = username
		}
	}

	for _, conn := range game.Players {
		conn.Write([]byte(fmt.Sprintf("\nGame over! The word was: %s\n", game.Word)))
		conn.Write([]byte(fmt.Sprintf("Winner: %s with %d points\n", winner, maxScore)))
	}
}

func removePlayer(username string) {
	gameLock.Lock()
	defer gameLock.Unlock()

	for i, conn := range game.Players {
		if keyMap[getKeyFromConn(conn)] == username {
			game.Players = append(game.Players[:i], game.Players[i+1:]...)
			delete(game.Scores, username)
			break
		}
	}

	if len(game.Players) < 2 && game.GameStarted {
		for _, conn := range game.Players {
			conn.Write([]byte("Not enough players. Game paused.\n"))
		}
		game.GameStarted = false
	}
}

func getKeyFromConn(conn net.Conn) string {
	for key, c := range connMap {
		if c == conn {
			return key
		}
	}
	return ""
}

func gameTimer() {
	for {
		time.Sleep(1 * time.Second)
		gameLock.Lock()
		if game != nil && game.GameStarted && time.Since(game.LastActivity) > 30*time.Second {
			// Timeout, switch player
			game.CurrentPlayer = (game.CurrentPlayer + 1) % len(game.Players)
			broadcastGameState()
		}
		gameLock.Unlock()
	}
}
