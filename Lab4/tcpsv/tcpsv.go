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
	Password string `json:"password"` // base64 encoded
}

type Player struct {
	conn   net.Conn
	name   string
	points int
	key    string
	scan   *bufio.Scanner
}

type Word struct {
	Description string
	Word        string
}

var (
	usersFile   = "users.json"
	users       []User
	playerQueue []*Player
	keyConnMap  = make(map[string]net.Conn)
	queueLock   sync.Mutex
	words       = []Word{
		{"A yellow fruit", "banana"},
		{"A programming language", "golang"},
		{"A network protocol", "socket"},
	}
)

func main() {
	loadUsers()

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

func loadUsers() {
	data, err := os.ReadFile(usersFile)
	if err != nil {
		fmt.Println("No users file found.")
		return
	}
	json.Unmarshal(data, &users)
}

func authenticate(username, password string) bool {
	encoded := base64.StdEncoding.EncodeToString([]byte(password))
	for _, u := range users {
		if u.Username == username && u.Password == encoded {
			return true
		}
	}
	return false
}

func generateKey() string {
	for {
		key := fmt.Sprintf("%d", rand.Intn(900)+100)
		if _, exists := keyConnMap[key]; !exists {
			return key
		}
	}
}

func handleLogin(conn net.Conn) {
	scan := bufio.NewScanner(conn)
	fmt.Fprintln(conn, "Username:")
	scan.Scan()
	username := scan.Text()

	fmt.Fprintln(conn, "Password:")
	scan.Scan()
	password := scan.Text()

	if !authenticate(username, password) {
		fmt.Fprintln(conn, "Authentication failed.")
		conn.Close()
		return
	}

	key := generateKey()
	keyConnMap[key] = conn
	fmt.Fprintf(conn, "Authenticated! Your key is: %s\n", key)

	player := &Player{
		conn: conn,
		name: username,
		key:  key,
		scan: scan,
	}

	queueLock.Lock()
	playerQueue = append(playerQueue, player)
	if len(playerQueue) >= 2 {
		p1 := playerQueue[0]
		p2 := playerQueue[1]
		playerQueue = playerQueue[2:]
		go startGame(p1, p2)
	}
	queueLock.Unlock()

	for scan.Scan() {
		msg := scan.Text()
		if strings.HasPrefix(msg, key+"_") {
			fmt.Println("["+key+"]:", msg)
			fmt.Fprintf(conn, "Received: %s\n", msg)
		} else {
			fmt.Fprintf(conn, "Invalid key prefix\n")
		}
	}
}

func startGame(p1, p2 *Player) {
	players := []*Player{p1, p2}
	word := words[rand.Intn(len(words))]
	hidden := make([]rune, len(word.Word))
	for i := range hidden {
		hidden[i] = '_'
	}

	broadcast := func(msg string) {
		for _, p := range players {
			fmt.Fprintf(p.conn, "%s_%s\n", p.key, msg)
		}
	}

	broadcast("Game Start!")
	broadcast("Description: " + word.Description)
	broadcast("Word: " + string(hidden))

	turn := 0
	for strings.Contains(string(hidden), "_") {
		current := players[turn%2]
		fmt.Fprintf(current.conn, "%s_Your turn! Guess a letter (30s):\n", current.key)

		timer := time.After(30 * time.Second)
		guessCh := make(chan string)

		go func() {
			if current.scan.Scan() {
				guessCh <- current.scan.Text()
			}
		}()

		select {
		case guess := <-guessCh:
			guess = strings.ToLower(strings.TrimSpace(guess))
			if len(guess) != 1 {
				fmt.Fprintf(current.conn, "%s_Invalid input. Your turn is over.\n", current.key)
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
				fmt.Fprintf(current.conn, "%s_Wrong guess. Your turn is over.\n", current.key)
				turn++
			}
		case <-timer:
			fmt.Fprintf(current.conn, "%s_Time out! Your turn is over.\n", current.key)
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
