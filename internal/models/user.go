package models

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Created  string `json:"created"`
	Updated  string `json:"updated"`
}
