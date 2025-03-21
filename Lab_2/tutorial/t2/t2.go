package main

import "fmt"

func main() {
	var intSlice = []int{42, 17, 89, 5, 63, 38, 49, 92, 29}
	chOdd := make(chan int)
	chEven := make(chan int)

	go odd(chOdd)
	go even(chEven)

	for _, value := range intSlice {
		if value%2 != 0 {
			chOdd <- value
		}

	}
}

func odd(ch <-chan int) {
	for v := range ch {
		fmt.Println("ODD : ", v)
	}
}

func even(ch <-chan int) {
	for v := range ch {
		fmt.Println("EVEN : ", v)
	}

}
