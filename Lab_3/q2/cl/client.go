package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func requestFileDownload(conn net.Conn, key string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter filename to download: ")
	filename, _ := reader.ReadString('\n')
	filename = strings.TrimSpace(filename)

	// Gửi yêu cầu có prefix key
	request := key + "_" + filename + "\n"
	conn.Write([]byte(request))

	// Tạo file để lưu dữ liệu nhận được
	outFile, err := os.Create("downloaded_" + filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outFile.Close()

	// Đọc toàn bộ nội dung gửi về từ server
	n, err := io.Copy(outFile, conn)
	if err != nil {
		fmt.Println("Download error:", err)
		return
	}

	fmt.Printf("Downloaded %d bytes. File saved as downloaded_%s\n", n, filename)
}
