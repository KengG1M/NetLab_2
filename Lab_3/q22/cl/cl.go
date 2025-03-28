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

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter filename to download: ")
	filename, _ := reader.ReadString('\n')
	filename = strings.TrimSpace(filename)

	// Send filename to server
	conn.Write([]byte(filename + "\n"))

	// Read server response
	serverReader := bufio.NewReader(conn)
	status, _ := serverReader.ReadString('\n')
	status = strings.TrimSpace(status)

	if status != "READY" {
		fmt.Println("Server response:", status)
		return
	}

	// Create output file
	outFile, err := os.Create("downloaded_" + filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outFile.Close()

	// Read file content from server and save
	n, err := io.Copy(outFile, serverReader)
	if err != nil {
		fmt.Println("Download failed:", err)
		return
	}

	fmt.Printf("Downloaded %d bytes and saved as downloaded_%s\n", n, filename)
}
