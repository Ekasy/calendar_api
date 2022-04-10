package delivery

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"nocalendar/internal/app/auth"
	"nocalendar/internal/app/errors"
	"nocalendar/internal/model"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type AuthDelivery struct {
	authUsecase auth.AuthUsecase
	logger      *logrus.Logger
}

func NewAuthDelivery(authDelivery auth.AuthUsecase, logger *logrus.Logger) *AuthDelivery {
	return &AuthDelivery{
		authUsecase: authDelivery,
		logger:      logger,
	}
}

func (ad *AuthDelivery) Routing(r *mux.Router) {
	r.HandleFunc("/auth", ad.Authorize).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/register", ad.Register).Methods(http.MethodPost, http.MethodOptions)
}

func (ad *AuthDelivery) Authorize(w http.ResponseWriter, r *http.Request) {
	authModel := &model.Auth{}
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ad.logger.Warnf("[Authorize] cannot convert body to bytes: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(buf, &authModel)
	if err != nil {
		ad.logger.Warnf("[Authorize] cannot convert body to bytes: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	usr, err := ad.authUsecase.GetUser(authModel)
	if err != nil {
		ad.logger.Warnf("[Authorize] user not authorized: %s", err.Error())
		switch err {
		case errors.BadPassword:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(errors.ErrorToBytes(err)))
		case errors.UserNotFound:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(errors.ErrorToBytes(err)))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Add("Authorize", usr.Token)
	w.WriteHeader(http.StatusOK)
	w.Write(model.ToBytes(usr.WithoutPassword()))
}

func (ad *AuthDelivery) Register(w http.ResponseWriter, r *http.Request) {
	usrModel := &model.User{}
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ad.logger.Warnf("[Authorize] cannot convert body to bytes: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(buf, &usrModel)
	if err != nil {
		ad.logger.Warnf("[Authorize] cannot convert body to bytes: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := ad.authUsecase.CreateUser(usrModel)
	if err != nil {
		ad.logger.Warnf("[Register] user not registered: %s", err.Error())
		switch err {
		case errors.LoginAlreadyExists, errors.EmailAlreadyExists:
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(errors.ErrorToBytes(err)))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Add("Authorize", token)
	w.WriteHeader(http.StatusOK)
}
