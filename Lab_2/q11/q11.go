package main

import (
	"fmt"
	"sync"
)

func main() {

	input := "I love you to the moon and back"

	numGors := 4

	result := charFrequency(input, numGors)

	for k, v := range result {
		charDisplay := string(k)

		if k == ' ' {
			charDisplay = "(blank)"
		}
		fmt.Printf("%s : %d \n", charDisplay, v)
	}
}

func countChars(chunk string) map[rune]int {
	freq := make(map[rune]int)
	for _, char := range chunk {
		freq[char]++
	}
	return freq
}

func mergeMaps(map1, map2 map[rune]int) {
	for k, v := range map2 {
		map1[k] += v
	}

}

func charFrequency(input string, numGors int) map[rune]int {
	length := len(input)

	if length == 0 || numGors < 1 {
		return nil
	}
	// calculate chunk size and remainder(if chia không đều)
	chunkSize := length / numGors
	remainder := length % numGors

	// Channel receive partial result of each supmap
	resultsChan := make(chan map[rune]int, numGors)

	var wg sync.WaitGroup
	start := 0

	// Create goroutines để đếm dựa vào số lượng goroutine muốn tạo
	// Ví dụ numGors ở đây là 4 thì vòng lặp 4 sẽ tạo ra bốn goroutines
	// để thực hiện quá trình đếm
	for i := 0; i < numGors; i++ {
		end := start + chunkSize

		if i < remainder {
			end++
		}
		if end > length {
			end = length
		}
		// Lấy chuỗi con
		chunk := input[start:end]
		start = end

		// Create goroutine to count
		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			freq := countChars(c)
			resultsChan <- freq
		}(chunk)
	}

	// Close channel after all of goroutine completed
	// use anonymous function
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Combine final result
	// sử dụng method mergeMaps() đã được define bên trên
	finalFreq := make(map[rune]int)
	for partialFreg := range resultsChan {
		mergeMaps(finalFreq, partialFreg)
	}
	return finalFreq // finalFreq là kết quả cuối cùng ở đây sau khi tách thành các chunk nhỏ và đếm riêng trên từng goroutine sau đó có những partialFreq là kết quả thu được sau khi đếm riềng từng chunk đó. Cuối cùng gộp nó lại bằng method mergeMaps()
}
