package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	stmt, _ := DB.Prepare(`
		INSERT INTO users (username, firstname, lastname, email, avatar, phone, dob, country, city, street_name, street_address)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	result, err := stmt.Exec(user.Username, user.Firstname, user.Lastname, user.Email,
		user.Avatar, user.Phone, user.DOB, user.Country, user.City, user.StreetName, user.StreetAddress)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	user.ID = int(id)
	c.JSON(http.StatusCreated, user)
}

func GetUsers(c *gin.Context) {
	rows, err := DB.Query("SELECT * FROM users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		rows.Scan(&u.ID, &u.Username, &u.Firstname, &u.Lastname, &u.Email, &u.Avatar, &u.Phone, &u.DOB, &u.Country, &u.City, &u.StreetName, &u.StreetAddress)
		users = append(users, u)
	}
	c.JSON(http.StatusOK, users)
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	stmt, _ := DB.Prepare(`
		UPDATE users SET username=?, firstname=?, lastname=?, email=?, avatar=?, phone=?, dob=?, country=?, city=?, street_name=?, street_address=? WHERE id=?
	`)
	_, err := stmt.Exec(user.Username, user.Firstname, user.Lastname, user.Email,
		user.Avatar, user.Phone, user.DOB, user.Country, user.City, user.StreetName, user.StreetAddress, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	stmt, _ := DB.Prepare("DELETE FROM users WHERE id=?")
	_, err := stmt.Exec(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
