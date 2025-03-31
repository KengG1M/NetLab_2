package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	inputReader := bufio.NewReader(os.Stdin)
	serverReader := bufio.NewReader(conn)

	for {
		// Đọc một dòng từ server (đến ký tự '\n')
		line, err := serverReader.ReadString('\n')
		if err != nil {
			fmt.Println("Server closed connection or error occurred:", err)
			return
		}
		line = strings.TrimSpace(line)

		// Dựa vào nội dung dòng nhận được để xử lý
		switch line {
		case "input username:":
			fmt.Print(line + " ")
			userInput, _ := inputReader.ReadString('\n')
			conn.Write([]byte(userInput)) // gửi đến server
			continue

		case "input pass:":
			fmt.Print(line + " ")
			passInput, _ := inputReader.ReadString('\n')
			conn.Write([]byte(passInput))
			continue

		case "Success(st).":
			// Đăng nhập thành công
			fmt.Println("Login success! ✅")
			break // thoát vòng for => sang phần xử lý tiếp

		case "Login failed.":
			// Sai 3 lần => kết thúc
			fmt.Println("❌ Too many failed attempts. Exiting...")
			return

		default:
			// Có thể là "Wrong credentials. 2 attempt(s) left."
			fmt.Println(line)
			// Chỉ in ra và loop tiếp => cho nhập lại
			continue
		}

		// Nếu rơi vào case "Success(st).", ta break vòng for
		break
	}

	// Tới đây => đăng nhập thành công => có thể chat hoặc làm gì đó
	for {
		fmt.Print("Enter msg to send (or 'exit'): ")
		userMsg, _ := inputReader.ReadString('\n')
		userMsg = strings.TrimSpace(userMsg)
		if userMsg == "exit" {
			fmt.Println("Bye!")
			return
		}

		conn.Write([]byte(userMsg + "\n"))

		// đọc reply
		reply, err := serverReader.ReadString('\n')
		if err != nil {
			fmt.Println("Server closed connection or error occurred:", err)
			return
		}
		fmt.Println("Server reply:", strings.TrimSpace(reply))
	}
}
