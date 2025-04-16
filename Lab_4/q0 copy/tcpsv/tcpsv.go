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

	// Modified handleConnection segment
	gameLock.Lock()
	if game == nil {
		initGame()
	}
	game.Players = append(game.Players, conn)
	game.Scores[username] = 0
	playerCount := len(game.Players)
	shouldStart := playerCount >= 2 && !game.GameStarted
	gameLock.Unlock()

	if shouldStart {
		gameLock.Lock()
		game.GameStarted = true
		game.CurrentPlayer = 0 // Explicitly set first player
		gameLock.Unlock()

		// Force newline-terminated broadcast
		broadcastGameState()
		time.Sleep(100 * time.Millisecond) // Allow broadcast to complete
	}

	// Handle client messages
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			removePlayer(username)
			return
		}
		msg = strings.TrimSpace(msg)

		parts := strings.SplitN(msg, "_", 2)
		if len(parts) != 2 {
			conn.Write([]byte("Invalid message format. Use key_guess\n"))
			continue
		}

		msgKey, guess := parts[0], parts[1]
		if msgKey != key {
			conn.Write([]byte("Invalid key\n"))
			continue
		}

		handleGameMessage(username, guess)
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

	// Clear previous output
	for _, conn := range game.Players {
		conn.Write([]byte("\033[2J\033[H")) // Clear terminal
	}

	displayWord := ""
	for i, c := range game.Word {
		if game.Revealed[i] {
			displayWord += string(c) + " "
		} else {
			displayWord += "_ "
		}
	}

	// Build scores string
	scores := new(strings.Builder)
	for username, score := range game.Scores {
		fmt.Fprintf(scores, "%s: %d points\n", username, score)
	}

	// Get current player username
	currentPlayerKey := ""
	for key, conn := range connMap {
		if conn == game.Players[game.CurrentPlayer] {
			currentPlayerKey = key
			break
		}
	}
	currentPlayerUsername := keyMap[currentPlayerKey]

	// Send game state to all players
	for i, conn := range game.Players {
		conn.Write([]byte(fmt.Sprintf(
			"=== HANGMAN ===\n"+
				"Hint: %s\n"+
				"Word: %s\n"+
				"Scores:\n%s\n",
			game.Hint, displayWord, scores.String())))

		if i == game.CurrentPlayer {
			conn.Write([]byte(fmt.Sprintf(
				">>> YOUR TURN! (30 seconds)\n" +
					"Guess a letter: ")))
		} else {
			conn.Write([]byte(fmt.Sprintf(
				"Waiting for %s to guess...\n",
				currentPlayerUsername)))
		}
	}
}

func handleGameMessage(username, guess string) {
	gameLock.Lock()
	defer gameLock.Unlock()

	// Verify it's this player's turn
	currentPlayerKey := ""
	for key, conn := range connMap {
		if conn == game.Players[game.CurrentPlayer] {
			currentPlayerKey = key
			break
		}
	}

	if keyMap[currentPlayerKey] != username {
		connMap[username].Write([]byte("It's not your turn!\n"))
		return
	}

	// Process guess
	guess = strings.ToUpper(strings.TrimSpace(guess))
	if len(guess) != 1 || guess[0] < 'A' || guess[0] > 'Z' {
		connMap[username].Write([]byte("Invalid guess! Enter a single letter A-Z\n"))
		return
	}

	found := false
	count := 0
	for i, c := range game.Word {
		if string(c) == guess && !game.Revealed[i] {
			game.Revealed[i] = true
			found = true
			count++
		}
	}

	if found {
		game.Scores[username] += 10 * count
		broadcastMessage(fmt.Sprintf("%s guessed '%s' correctly! +%d points\n", username, guess, 10*count))
	} else {
		broadcastMessage(fmt.Sprintf("%s guessed '%s' incorrectly\n", username, guess))
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
	} else {
		broadcastGameState()
	}
}

func broadcastMessage(msg string) {
	for _, conn := range game.Players {
		conn.Write([]byte(msg))
	}
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

	broadcastMessage(fmt.Sprintf(
		"\nGAME OVER! The word was: %s\n"+
			"Winner: %s with %d points\n\n",
		game.Word, winner, maxScore))
}

// func broadcastGameOver() {
// 	winner := ""
// 	maxScore := -1
// 	for username, score := range game.Scores {
// 		if score > maxScore {
// 			maxScore = score
// 			winner = username
// 		}
// 	}

// 	for _, conn := range game.Players {
// 		conn.Write([]byte(fmt.Sprintf("\nGame over! The word was: %s\n", game.Word)))
// 		conn.Write([]byte(fmt.Sprintf("Winner: %s with %d points\n", winner, maxScore)))
// 	}
// }

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
	gameLock.Lock()
	defer gameLock.Unlock()

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
