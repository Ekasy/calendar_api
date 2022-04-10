package usecase

import (
	"fmt"
	"nocalendar/internal/app/auth"
	"nocalendar/internal/app/errors"
	"nocalendar/internal/model"
	"time"

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

func generateToken(login string) string {
	sec := time.Now().Unix()
	hash_, _ := bcrypt.GenerateFromPassword([]byte(fmt.Sprintf("%s-%d", login, sec)), 4)
	return string(hash_)
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
	usr.Token = generateToken(usr.Login)

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
