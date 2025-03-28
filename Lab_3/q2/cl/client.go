package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	defer conn.Close()

	key := "123" // giả sử client đã xác thực và có key này
	requestFileDownload(conn, key)

	// input exit
	// đọc data theo từ terminal/ NewReader(os.Stdin) đọc data từ terminal
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter exit to quit this fking program")
	// ReadString ở đây nó sẽ đọc data nhập từ bàn phím cho tới khi enter == \n
	text, _ := reader.ReadString('\n')

	// send data to sv
	conn.Write([]byte(text))

}

func requestFileDownload(conn net.Conn, key string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter filename to download: ")

	// Dung readstring data se record luon ca dau xuong dong \n
	filename, _ := reader.ReadString('\n')
	// -> nen vi the khi su dung trimspace thi no se xoa nhung dau cach tab hay \n
	filename = strings.TrimSpace(filename)

	request := key + "_" + filename + "\n"
	conn.Write([]byte(request))

	serverReader := bufio.NewReader(conn)
	status, _ := serverReader.ReadString('\n')

	if strings.TrimSpace(status) != "READY" {
		fmt.Println("Server response:", status)
		return
	}

	outFile, err := os.Create("downloaded_" + filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outFile.Close()

	n, err := io.Copy(outFile, serverReader)
	if err != nil {
		fmt.Println("Download error:", err)
		return
	}

	fmt.Printf("Downloaded %d bytes. File saved as downloaded_%s\n", n, filename)
}
