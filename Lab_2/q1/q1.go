package main

import (
	"fmt"
	"sync"
)

// countChars: Hàm đếm ký tự trong một chuỗi chunk
func countChars(chunk string) map[rune]int {
	freq := make(map[rune]int)
	for _, char := range chunk {
		freq[char]++
	}
	return freq
}

// mergeMaps: Hàm gộp (cộng dồn) map2 vào map1
func mergeMaps(map1, map2 map[rune]int) {
	for k, v := range map2 {
		map1[k] += v
	}
}

// concurrentCharFrequency: Hàm chính, chia chuỗi và chạy các goroutine
func concurrentCharFrequency(input string, numGoroutines int) map[rune]int {
	length := len(input)
	if length == 0 || numGoroutines < 1 {
		return nil
	}

	// Tính toán kích thước chunk, và phần dư (nếu chia không đều)
	chunkSize := length / numGoroutines
	remainder := length % numGoroutines

	// Channel để lấy kết quả partial (từng map con)
	resultsChan := make(chan map[rune]int, numGoroutines)

	var wg sync.WaitGroup
	start := 0

	// Tạo các goroutine
	for i := 0; i < numGoroutines; i++ {
		end := start + chunkSize
		// Phân bổ thêm 1 ký tự cho các goroutine đầu tiên nếu có remainder
		if i < remainder {
			end++
		}
		if end > length {
			end = length
		}

		// Lấy chuỗi con
		chunk := input[start:end]
		start = end

		// Tạo goroutine để đếm
		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			freq := countChars(c)
			resultsChan <- freq
		}(chunk)
	}

	// Đóng channel sau khi tất cả goroutine hoàn thành
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Gộp kết quả cuối cùng
	finalFreq := make(map[rune]int)
	for partialFreq := range resultsChan {
		mergeMaps(finalFreq, partialFreq)
	}

	return finalFreq
}

func main() {
	// Ví dụ chuỗi
	input := "Hoàng Ngọc Quỳnh Anh"

	// Số lượng goroutine (số luồng logic)
	numGoroutines := 4

	// Thực thi đếm tần suất ký tự
	result := concurrentCharFrequency(input, numGoroutines)

	// In kết quả
	for k, v := range result {
		// Hiển thị "(blank)" thay cho khoảng trắng
		charDisplay := string(k)
		if k == ' ' {
			charDisplay = "(blank)"
		}
		fmt.Printf("%s: %d\n", charDisplay, v)
	}
}
