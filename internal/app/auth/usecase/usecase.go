package usecase

import (
	"nocalendar/internal/app/auth"
	"nocalendar/internal/app/errors"
	"nocalendar/internal/model"
	"nocalendar/internal/util"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	repo   auth.AuthRepository
	logger *logrus.Logger
}

func NewAuthUsecase(repo auth.AuthRepository, logger *logrus.Logger) auth.AuthUsecase {
	return &AuthUsecase{
		repo:   repo,
		logger: logger,
	}
}

func (au *AuthUsecase) CreateUser(usr *model.User) (string, error) {
	valid, err := au.repo.CheckUser(usr)
	if err != nil || !valid {
		return "", err
	}

	hash_, err := bcrypt.GenerateFromPassword([]byte(usr.Password), 4)
	if err != nil {
		return "", errors.InternalError
	}
	usr.Password = string(hash_)
	usr.Token = util.GenerateRandomString(32)

	usr, err = au.repo.Insert(usr)
	return usr.Token, err
}

func checkPassword(raw string, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw))
	if err != nil {
		return errors.BadPassword
	}
	return nil
}

func (au *AuthUsecase) GetUser(ausr *model.Auth) (*model.User, error) {
	usr, err := au.repo.GetUser(ausr.Login)
	if err != nil {
		return nil, err
	}

	err = checkPassword(ausr.Password, usr.Password)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

func (au *AuthUsecase) GetUserByToken(token string) (*model.User, error) {
	login, err := au.repo.GetLoginByToken(token)
	if err != nil {
		return nil, err
	}

	usr, err := au.repo.GetUser(login)
	return usr, err
}
