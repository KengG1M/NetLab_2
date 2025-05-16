package main

type User struct {
	ID            int    `json:"id"`
	Username      string `json:"username"`
	Firstname     string `json:"firstname"`
	Lastname      string `json:"lastname"`
	Email         string `json:"email"`
	Avatar        string `json:"avatar"`
	Phone         string `json:"phone"`
	DOB           string `json:"dob"`
	Country       string `json:"country"`
	City          string `json:"city"`
	StreetName    string `json:"street_name"`
	StreetAddress string `json:"street_address"`
}
