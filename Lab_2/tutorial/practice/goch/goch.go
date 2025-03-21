package main

import (
	"fmt"
	"time"
)

func worker(n int, ch chan<- int) {
	time.Sleep(500 * time.Millisecond)
	ch <- n * n
}

func main() {
	ch := make(chan int)

	// Tạo  5 goroutine
	for i := 1; i <= 5; i++ {
		go worker(i, ch)
	}

	// Nhận 5 kết quả
	for i := 1; i <= 5; i++ {
		result := <-ch
		fmt.Println("result received:", result)
	}

	close(ch)
}
