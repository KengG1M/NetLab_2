package main

// Goroutine va channel are 2 core components supporting concurrent programming
// in Go language

/*
	Goroutine la mot luong thuc thi (execution thread) nhe trong Go.
	No nhe hơn rất nhiều so với luồng thread ở hầu hết ngôn ngữ khác
(execution thread != thread)

	Khi bạn gọi một hàm kèm từ khóa go phía trước, GO sẽ tạo ra một goroutine
	mới để chạy hàm đó một cách không đồng bộ (asynchronous)

	Goroutine dùng mô hình hợp tác (cooperative scheduling) thay vì mỗi
	goroutine chạy độc lập như thread hệ điều hành. NHiều goriutine
	có thể chạy trên một hay nhiều Hẹ điều hành, do Go runtime quản lý tự động

	Tạo ra hàng nghìn hay hàng triệu goroutine vẫn thường khả thi hơn là tạo
	cùng số lượng thread thực tieeps ở nhiều ngôn ngữ khác nhau (như java,
	C++) nhờ tính lightweight (nhẹ)

*/

import (
	"fmt"
	"time"
)

func doSomething() {
	fmt.Println("Hello from goroutine")
}

func main() {
	go doSomething()
	fmt.Println("Hello from main")

	time.Sleep(time.Second)
}
