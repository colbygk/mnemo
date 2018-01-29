package models

type User struct {
	UUID     string `json:"uuid" form:"-"`
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
	Email    string `json:"email" form:"email"`
	CanAdmin bool   `json:"canadmin"`
	JWT      string
	JTW_age  int
}
