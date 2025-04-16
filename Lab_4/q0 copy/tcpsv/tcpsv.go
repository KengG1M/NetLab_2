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

// --- Cấu trúc dữ liệu ---

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

// Player lưu trữ thông tin của mỗi client sau khi xác thực thành công.
type Player struct {
	Username string
	Conn     net.Conn
	Key      string
}

// GameState lưu trữ trạng thái game hiện tại.
type GameState struct {
	Word          string
	Hint          string
	Revealed      []bool
	Players       []Player
	CurrentPlayer int
	Scores        map[string]int
	GameStarted   bool
	LastActivity  time.Time
	Mutex         sync.Mutex
}

var (
	Users    []User
	Words    []Word
	game     *GameState
	gameLock sync.Mutex
)

// --- Hàm main và load dữ liệu ---

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

// --- Xác thực và xử lý kết nối client ---

func handleConnection(conn net.Conn) {
	// Đặt deadline ban đầu cho kết nối (5 phút)
	conn.SetDeadline(time.Now().Add(5 * time.Minute))
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Xác thực người dùng
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

	// Sinh key phiên cho kết nối hiện tại
	rand.Seed(time.Now().UnixNano())
	key := fmt.Sprintf("%d", rand.Intn(9000)+1000)

	// Tạo đối tượng Player
	player := Player{
		Username: username,
		Conn:     conn,
		Key:      key,
	}

	// Thông báo xác thực thành công và gửi key cho client
	conn.Write([]byte(fmt.Sprintf("Authentication successful. Your key is %s\n", key)))

	// Thêm player vào game state
	gameLock.Lock()
	if game == nil {
		initGame()
	}
	game.Players = append(game.Players, player)
	game.Scores[username] = 0
	playerCount := len(game.Players)
	shouldStart := playerCount >= 2 && !game.GameStarted
	gameLock.Unlock()

	// Bắt đầu game nếu đủ 2 người chơi
	if shouldStart {
		gameLock.Lock()
		game.GameStarted = true
		game.CurrentPlayer = 0
		gameLock.Unlock()
		broadcastGameState()
	}

	// Vòng lặp lắng nghe message từ client
	for {
		// Thiết lập deadline 30s cho mỗi lượt đoán
		conn.SetDeadline(time.Now().Add(30 * time.Second))
		msg, err := reader.ReadString('\n')
		if err != nil {
			// Nếu hết hạn 30s, chuyển lượt nếu người chơi này đang có lượt.
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				gameLock.Lock()
				// Nếu chính xác người chơi này đang có lượt, báo timeout và chuyển lượt.
				if game.Players[game.CurrentPlayer].Username == username {
					conn.Write([]byte("Timeout! You lost your turn.\n"))
					game.CurrentPlayer = (game.CurrentPlayer + 1) % len(game.Players)
					broadcastGameState()
				}
				gameLock.Unlock()
				continue
			}
			removePlayer(username)
			return
		}

		msg = strings.TrimSpace(msg)
		// Message phải có định dạng key_guess, ví dụ: "125_A"
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

		// Xử lý lượt đoán trên game (nếu đến lượt)
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

// Khởi tạo game mới, chọn ngẫu nhiên một từ và reset trạng thái
func initGame() {
	wordObj := Words[rand.Intn(len(Words))]
	game = &GameState{
		Word:         strings.ToUpper(wordObj.Text),
		Hint:         wordObj.Hint,
		Revealed:     make([]bool, len(wordObj.Text)),
		Players:      make([]Player, 0),
		Scores:       make(map[string]int),
		GameStarted:  false,
		LastActivity: time.Now(),
	}
}

// --- Các hàm xử lý và thông báo game ---

// Gửi trạng thái game tới tất cả người chơi
func broadcastGameState() {
	gameLock.Lock()
	defer gameLock.Unlock()

	// Cập nhật thời gian hoạt động cuối cùng
	game.LastActivity = time.Now()

	// Xây dựng chuỗi hiển thị từ (chữ đã hé lộ hoặc dấu _)
	displayWord := ""
	for i, c := range game.Word {
		if game.Revealed[i] {
			displayWord += string(c) + " "
		} else {
			displayWord += "_ "
		}
	}

	// Xây dựng bảng điểm
	var scoresBuilder strings.Builder
	for username, score := range game.Scores {
		scoresBuilder.WriteString(fmt.Sprintf("%s: %d points\n", username, score))
	}

	currentPlayerUsername := ""
	if len(game.Players) > 0 {
		currentPlayerUsername = game.Players[game.CurrentPlayer].Username
	}

	// Gửi thông báo tới tất cả người chơi
	for i, player := range game.Players {
		player.Conn.Write([]byte("\033[2J\033[H")) // Clear terminal
		header := fmt.Sprintf("=== HANGMAN ===\nHint: %s\nWord: %s\nScores:\n%s\n",
			game.Hint, displayWord, scoresBuilder.String())
		player.Conn.Write([]byte(header))
		if i == game.CurrentPlayer {
			player.Conn.Write([]byte(">>> YOUR TURN! (30 seconds)\nGuess a letter: "))
		} else {
			player.Conn.Write([]byte(fmt.Sprintf("Waiting for %s to guess...\n", currentPlayerUsername)))
		}
	}
}

// Xử lý lượt đoán của người chơi
func handleGameMessage(username, guess string) {
	gameLock.Lock()
	defer gameLock.Unlock()

	// Kiểm tra lượt hiện tại: nếu người gửi không phải người chơi được chỉ định hiện tại
	if game.Players[game.CurrentPlayer].Username != username {
		// Tìm đối tượng player gửi và gửi thông báo lỗi
		for _, p := range game.Players {
			if p.Username == username {
				p.Conn.Write([]byte("It's not your turn!\n"))
				break
			}
		}
		return
	}

	guess = strings.ToUpper(strings.TrimSpace(guess))
	if len(guess) != 1 || guess[0] < 'A' || guess[0] > 'Z' {
		for _, p := range game.Players {
			if p.Username == username {
				p.Conn.Write([]byte("Invalid guess! Enter a single letter A-Z\n"))
				break
			}
		}
		return
	}

	// Xử lý đoán chữ: kiểm tra chữ đã có trong từ chưa và chưa được hé lộ
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
		// Nếu đoán sai, chuyển lượt cho người chơi kế tiếp
		game.CurrentPlayer = (game.CurrentPlayer + 1) % len(game.Players)
	}

	// Kiểm tra nếu toàn bộ chữ đã được hé lộ => game kết thúc
	gameOver := true
	for _, revealed := range game.Revealed {
		if !revealed {
			gameOver = false
			break
		}
	}

	if gameOver {
		// Xác định người chiến thắng (số điểm cao nhất)
		winner := ""
		maxScore := -1
		for uname, score := range game.Scores {
			if score > maxScore {
				maxScore = score
				winner = uname
			}
		}
		broadcastMessage(fmt.Sprintf("\nGAME OVER! The word was: %s\nWinner: %s with %d points\n\n", game.Word, winner, maxScore))
		// Khởi tạo game mới
		initGame()
		if len(game.Players) >= 2 {
			game.GameStarted = true
		}
	}
	broadcastGameState()
}

// Gửi thông điệp tới tất cả người chơi
func broadcastMessage(msg string) {
	for _, player := range game.Players {
		player.Conn.Write([]byte(msg))
	}
}

// Xóa người chơi khi gặp lỗi hoặc ngắt kết nối
func removePlayer(username string) {
	gameLock.Lock()
	defer gameLock.Unlock()

	index := -1
	for i, player := range game.Players {
		if player.Username == username {
			index = i
			break
		}
	}
	if index != -1 {
		game.Players = append(game.Players[:index], game.Players[index+1:]...)
		delete(game.Scores, username)
	}

	if len(game.Players) < 2 && game.GameStarted {
		for _, player := range game.Players {
			player.Conn.Write([]byte("Not enough players. Game paused.\n"))
		}
		game.GameStarted = false
	}
}

// Goroutine kiểm tra timeout tổng thể: nếu không có hoạt động trong hơn 30 giây, chuyển lượt
func gameTimer() {
	for {
		time.Sleep(1 * time.Second)
		gameLock.Lock()
		if game != nil && game.GameStarted && time.Since(game.LastActivity) > 30*time.Second {
			game.CurrentPlayer = (game.CurrentPlayer + 1) % len(game.Players)
			broadcastGameState()
		}
		gameLock.Unlock()
	}
}
