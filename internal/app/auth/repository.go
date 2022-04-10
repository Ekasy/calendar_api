package auth

import (
	"nocalendar/internal/model"
)

type AuthRepository interface {
	Insert(usr *model.User) (*model.User, error)
	CheckUser(usr *model.User) (bool, error)
	GetUser(login string) (*model.User, error)
}
