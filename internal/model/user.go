package model

type User struct {
	Login    string `json:"login"`
	Name     string `json:"name" bson:"name"`
	Surname  string `json:"surname" bson:"surname"`
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
	Token    string `bson:"token"`
}

type UserWithoutPassword struct {
	Login   string `json:"login"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Email   string `json:"email"`
}

type JsonUser struct {
	Id    string          `json:"_id" bson:"_id"`
	Users map[string]User `json:"users" bson:"users"`
}

func (u *User) WithoutPassword() *UserWithoutPassword {
	return &UserWithoutPassword{
		Login:   u.Login,
		Name:    u.Name,
		Surname: u.Surname,
		Email:   u.Email,
	}
}

type JsonTokens struct {
	Id     string            `json:"_id" bson:"_id"`
	Tokens map[string]string `json:"tokens" bson:"tokens"`
}
