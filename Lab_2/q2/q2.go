package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	LibraryCapacity = 30
	TotalStudents   = 100
	MaxStudyTime    = 4
)

type Student struct {
	ID       int
	Duration int
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var students []Student
	for i := 1; i <= TotalStudents; i++ {
		students = append(students, Student{
			ID:       rand.Intn(100) + 1,
			Duration: rand.Intn(MaxStudyTime) + 1,
		})
	}

	currentTime := 0

	libSeats := []Student{}

	fmt.Println("Library simulation started")

	for len(students) > 0 || len(libSeats) > 0 {
		for len(students) > 0 && len(libSeats) < LibraryCapacity {
			student := students[0]
			students = students[1:]
			libSeats = append(libSeats, student)
			fmt.Printf("Time %d: Student %d start reading at the lib\n", currentTime, student.ID)
		}

		if len(students) > 0 {
			fmt.Printf("Time %d: Student %d is waiting\n", currentTime, students[0].ID)
		}

		var remainingSeats []Student
		for _, student := range libSeats {
			if student.Duration == 1 {
				fmt.Printf("Time %d: Student %d is leaving. Spent %d hours reading\n", currentTime, student.ID, student.Duration)
			} else {
				student.Duration--
				remainingSeats = append(remainingSeats, student)
			}
		}
		libSeats = remainingSeats
		time.Sleep(1 * time.Second)
		currentTime++
	}

	fmt.Printf("Time %d: No more students. Let's call it a day\n", currentTime)
}
