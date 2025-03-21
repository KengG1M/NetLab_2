package main

/*
Channel là cơ chế giao tiếp an toàn và đồng bộ giữa các goroutine

Bạn có thể hình dung Channel như một đường ống pipe mà qua ddos các goroutine trao đổi dữ liệu với nhau

Việc gửi và nhận dữ liệu trên channel có thể chặn (blocking)
goroutine cho đến khi dữ liệu sẵn sàng, giúp tránh tình trạng
race-condition(tranh chấp dữ liệu) mà không cần khóa (mutex) phức tạp
*/

import (
	"fmt"
	"time"
)

func main() {

	ch := make(chan string)
	go func() {
		time.Sleep(500 * time.Millisecond)
		ch <- "Hai baybi"
		ch <- "hahahahaha"
		close(ch)
	}()

	for msg := range ch {
		fmt.Println(msg)
	}

	fmt.Println("Closed channel, exit................")

}
