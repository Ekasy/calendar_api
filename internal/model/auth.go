package model

type Auth struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
