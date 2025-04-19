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
	Email    string `json:"email"`
	Address  string `json:"address"`
}

type Player struct {
	conn   net.Conn
	name   string
	points int
	scan   *bufio.Scanner
}

type Word struct {
	Description string
	Word        string
}

var (
	users       []User
	playerQueue []*Player
	queueLock   sync.Mutex
	words       = []Word{
		{"A yellow fruit", "banana"},
		{"A programming language", "golang"},
		{"A network protocol", "socket"},
	}
)

var keyMap = make(map[string]net.Conn)

func main() {
	loadUsers("users.json")

	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is running at port 12345")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleLogin(conn)
	}
}

func loadUsers(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("No users file found:", err)
		os.Exit(1)
	}
	json.Unmarshal(data, &users)
}

func authenticate(username, encodedPassword string) bool {
	for _, u := range users {
		if u.Username == username && u.Password == encodedPassword {
			return true
		}
	}
	return false
}

func handleLogin(conn net.Conn) {
	scan := bufio.NewScanner(conn)

	fmt.Fprintln(conn, "Username:")
	scan.Scan()
	username := strings.TrimSpace(scan.Text())

	fmt.Fprintln(conn, "Password:")
	scan.Scan()
	password := strings.TrimSpace(scan.Text())

	encodedPw := base64.StdEncoding.EncodeToString([]byte(password))

	if !authenticate(username, encodedPw) {
		fmt.Fprintln(conn, "Authentication failed.")
		conn.Close()
		return
	}

	// generate key for each client
	rand.Seed(time.Now().UnixNano())
	key := fmt.Sprintf("%d", rand.Intn(1000)+100)
	keyMap[key] = conn

	fmt.Fprintf(conn, "Authenticated! Welcome "+key+"_%s\n", username)

	player := &Player{
		conn: conn,
		name: username,
		scan: scan,
	}

	queueLock.Lock()
	playerQueue = append(playerQueue, player)
	if len(playerQueue) >= 2 {
		p1 := playerQueue[0]
		p2 := playerQueue[1]
		playerQueue = playerQueue[2:] // bỏ 2 player đầu ở hàng đợi
		go startGame(p1, p2)          // bắt đầu game ở goroutine riêng
	}
	queueLock.Unlock()
}

func startGame(p1, p2 *Player) {
	players := []*Player{p1, p2}

	// random word from words
	word := words[rand.Intn(len(words))]

	// slice contain underscore char to hide guessing keyword
	hidden := make([]rune, len(word.Word))
	for i := range hidden {
		hidden[i] = '_'
	}

	broadcast := func(msg string) {
		for _, p := range players {
			fmt.Fprintln(p.conn, msg)
		}
	}

	broadcast("Game Start!")
	broadcast("Description: " + word.Description)
	broadcast("Word: " + string(hidden))

	turn := 0

	for strings.Contains(string(hidden), "_") {
		current := players[turn%2]
		fmt.Fprintf(current.conn, "Your turn! Guess a letter (30s):\n")

		timer := time.After(30 * time.Second)
		guessCh := make(chan string)

		go func() {
			if current.scan.Scan() {
				guessCh <- current.scan.Text()
			}
		}()

		select {
		case guess := <-guessCh:
			guess = strings.ToLower(guess)
			if len(guess) != 1 {
				fmt.Fprintln(current.conn, "Invalid input. Your turn is over.")
				turn++
				continue
			}
			correct := false
			count := 0
			for i, ch := range word.Word {
				if string(ch) == guess && hidden[i] == '_' {
					hidden[i] = ch
					correct = true
					count++
				}
			}
			if correct {
				current.points += count * 10
				broadcast(fmt.Sprintf("%s guessed '%s' correctly! Word: %s | %s: %d pts", current.name, guess, string(hidden), current.name, current.points))
			} else {
				fmt.Fprintf(current.conn, "Wrong guess. Your turn is over.\n")
				turn++
			}
		case <-timer:
			fmt.Fprintf(current.conn, "Time out! Your turn is over.\n")
			turn++
		}
	}

	winner := p1
	if p2.points > p1.points {
		winner = p2
	}
	broadcast(fmt.Sprintf("Game Over! Final word: %s", word.Word))
	broadcast(fmt.Sprintf("Winner is %s with %d points!", winner.name, winner.points))
	for _, p := range players {
		p.conn.Close()
	}
}
