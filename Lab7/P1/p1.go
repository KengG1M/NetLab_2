package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type FileInfo struct {
	FileName string
	FileSize int64
	Owner    string
}

var localCatalog = make(map[string]FileInfo)

func broadcastFile(fileName string, fileSize int64) {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 404,
	})
	if err != nil {
		fmt.Println("Broadcast error:", err)
		return
	}
	defer conn.Close()

	msg := fmt.Sprintf("%s|%d|%s", fileName, fileSize, getLocalIP())
	conn.Write([]byte(msg))
}

func listenForBroadcasts() {
	addr, _ := net.ResolveUDPAddr("udp", ":403")
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Listen broadcast error:", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, _, _ := conn.ReadFromUDP(buf)
		parts := strings.Split(string(buf[:n]), "|")
		if len(parts) == 3 {
			name := parts[0]
			size, _ := strconv.ParseInt(parts[1], 10, 64)
			ip := parts[2]
			fmt.Printf("Received broadcast: %s\n", string(buf[:n]))
			fmt.Printf("Catalog updated with: %s from %s\n", name, ip)
			localCatalog[name] = FileInfo{name, size, ip}
		}

	}
}

func startFileServer() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("File server error:", err)
		return
	}
	defer ln.Close()

	for {
		conn, _ := ln.Accept()
		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 256)
			n, _ := c.Read(buf)
			fileName := string(buf[:n])
			file, err := os.Open(fileName)
			if err != nil {
				fmt.Println("File open error:", err)
				return
			}
			defer file.Close()
			io.Copy(c, file)
		}(conn)
	}
}

func downloadFile(fileInfo FileInfo) {
	conn, err := net.Dial("tcp", fileInfo.Owner+":8080")
	if err != nil {
		fmt.Println("Download error:", err)
		return
	}
	defer conn.Close()

	conn.Write([]byte(fileInfo.FileName))
	file, _ := os.Create("downloaded_" + fileInfo.FileName)
	defer file.Close()
	io.Copy(file, conn)
	fmt.Println("Downloaded", fileInfo.FileName)
}

func broadcastSearch(query string) {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 9001,
	})
	if err != nil {
		fmt.Println("Search broadcast error:", err)
		return
	}
	defer conn.Close()

	conn.Write([]byte("SEARCH|" + query))
}

func listenForSearch() {
	addr, _ := net.ResolveUDPAddr("udp", ":9001")
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Search listen error:", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, remoteAddr, _ := conn.ReadFromUDP(buf)
		parts := strings.Split(string(buf[:n]), "|")
		if parts[0] == "SEARCH" {
			query := parts[1]
			for _, f := range localCatalog {
				if strings.Contains(f.FileName, query) {
					response := fmt.Sprintf("%s|%d|%s", f.FileName, f.FileSize, getLocalIP())
					conn.WriteToUDP([]byte(response), remoteAddr)
				}
			}
		}
	}
}

func getLocalIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func main() {
	go listenForBroadcasts()
	go startFileServer()
	go listenForSearch()

	for {
		var cmd string
		fmt.Print("Command (share/search/download/exit): ")
		fmt.Scanln(&cmd)

		switch cmd {
		case "share":
			fmt.Print("File name: ")
			var name string
			fmt.Scanln(&name)
			fi, err := os.Stat(name)
			if err != nil {
				fmt.Println("Invalid file")
				continue
			}
			broadcastFile(name, fi.Size())
			fmt.Println("File info broadcasted once.")

		case "search":
			fmt.Print("Enter keyword: ")
			var keyword string
			fmt.Scanln(&keyword)
			broadcastSearch(keyword)
		case "download":
			fmt.Print("Enter file name to download: ")
			var file string
			fmt.Scanln(&file)
			fmt.Println("Current catalog entries:")
			for k, v := range localCatalog {
				fmt.Printf(" - %s at %s\n", k, v.Owner)
			}

			if info, ok := localCatalog[file]; ok {
				downloadFile(info)
			} else {
				fmt.Println("File not found.")
			}
		case "exit":
			return
		}
	}
}
