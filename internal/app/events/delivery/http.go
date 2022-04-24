package delivery

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"nocalendar/internal/app/auth"
	"nocalendar/internal/app/errors"
	"nocalendar/internal/app/events"
	"nocalendar/internal/app/middleware"
	"nocalendar/internal/model"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type EventsDelivery struct {
	eventUsecase events.EventsUsecase
	authUsecase  auth.AuthUsecase
	logger       *logrus.Logger
}

func NewEventsDelivery(eventUsecase events.EventsUsecase, authUsecase auth.AuthUsecase, logger *logrus.Logger) *EventsDelivery {
	return &EventsDelivery{
		eventUsecase: eventUsecase,
		authUsecase:  authUsecase,
		logger:       logger,
	}
}

func (ed *EventsDelivery) Routing(r *mux.Router) {
	ev := r.PathPrefix("/event").Subrouter()
	am := middleware.NewAuthMiddleware(ed.authUsecase, ed.logger)
	ev.Use(am.TokenChecking)

	ev.HandleFunc("", ed.CreateEvent).Methods(http.MethodPost, http.MethodOptions)
	ev.HandleFunc("/{event_id:[\\w]+}", ed.GetEvent).Methods(http.MethodGet, http.MethodOptions)
}

func (ed *EventsDelivery) CreateEvent(w http.ResponseWriter, r *http.Request) {
	eventModel := &model.Event{}
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ed.logger.Warnf("[CreateEvent] cannot convert body to bytes: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(buf, &eventModel)
	if err != nil {
		ed.logger.Warnf("[CreateEvent] cannot unmarshal bytes: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	usr := r.Context().Value(middleware.ContextUserKey).(*model.User)
	eventId, err := ed.eventUsecase.CreateEvent(eventModel, usr.Login)
	if err != nil {
		ed.logger.Warnf("[CreateEvent] user not registered: %s", err.Error())
		switch err {
		case errors.LoginAlreadyExists, errors.EmailAlreadyExists:
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(errors.ErrorToBytes(err)))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(`{"message": "ok", "event_id": "%s"}`, eventId)))
}

func (ed *EventsDelivery) GetEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventId := vars["event_id"]
	usr := r.Context().Value(middleware.ContextUserKey).(*model.User)

	event, err := ed.eventUsecase.GetEvent(eventId, usr.Login)
	if err != nil {
		ed.logger.Warnf("[GetEvent] event not found: %s", err.Error())
		switch err {
		case errors.EventNotFound:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(errors.ErrorToBytes(err)))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(200)
	w.Write(model.ToBytes(event))
}
