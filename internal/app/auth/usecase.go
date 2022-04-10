package auth

import "nocalendar/internal/model"

type AuthUsecase interface {
	GetUser(usr *model.Auth) (*model.User, error)
	CreateUser(usr *model.User) (string, error)
}
