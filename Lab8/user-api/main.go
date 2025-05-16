package main

import "github.com/gin-gonic/gin"

func main() {
	InitDB() // ← Gọi từ db.go

	router := gin.Default()

	router.POST("/users", CreateUser)
	router.GET("/users", GetUsers)
	router.PUT("/users/:id", UpdateUser)
	router.DELETE("/users/:id", DeleteUser)

	router.Run(":8080")
}
