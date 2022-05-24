package delivery

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"nocalendar/internal/app/auth"
	"nocalendar/internal/app/errors"
	"nocalendar/internal/app/events"
	"nocalendar/internal/app/middleware"
	"nocalendar/internal/model"
	"strconv"
	"time"

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
	ev.HandleFunc("/one/{event_id:[\\w]+}", ed.GetEvent).Methods(http.MethodGet, http.MethodOptions)
	ev.HandleFunc("/edit", ed.EditEvent).Methods(http.MethodPost, http.MethodOptions)
	ev.HandleFunc("/all", ed.GetAllEvents).Methods(http.MethodGet, http.MethodOptions)
	ev.HandleFunc("/remove/{event_id:[\\w]+}", ed.RemoveEvent).Methods(http.MethodDelete, http.MethodOptions)

	ev.HandleFunc("/accept/{event_id:[\\w]+}", ed.AcceptInvite).Methods(http.MethodPost, http.MethodOptions)
	ev.HandleFunc("/invites", ed.GetInvites).Methods(http.MethodGet, http.MethodOptions)
	ev.HandleFunc("/reject/{event_id:[\\w]+}", ed.RejectInvite).Methods(http.MethodPost, http.MethodOptions)
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

func (ed *EventsDelivery) EditEvent(w http.ResponseWriter, r *http.Request) {
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
	event, err := ed.eventUsecase.EditEvent(eventModel, usr.Login)
	if err != nil {
		ed.logger.Warnf("[EditEvent] event not edited: %s", err.Error())
		switch err {
		case errors.EventNotFound:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(errors.ErrorToBytes(err)))
		case errors.HasNoRights:
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(errors.ErrorToBytes(err)))
		case errors.EventNotEdited:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(errors.ErrorToBytes(err)))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(200)
	w.Write(model.ToBytes(event.ToAnswer()))
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
	w.Write(model.ToBytes(event.ToAnswer()))
}

const DEFAULT_DAYS_INTERVAL = 30

func getFromToCgies(cgies url.Values) (int64, int64) {
	fromStr := cgies.Get("from")
	toStr := cgies.Get("to")
	from := int64(0)
	to := int64(0)
	var err error

	if fromStr == "" {
		from = time.Now().Unix() - DEFAULT_DAYS_INTERVAL*24*60*60
	} else {
		from, err = strconv.ParseInt(fromStr, 10, 64)
		if err != nil {
			return 0, 0
		}
	}

	if toStr == "" {
		to = time.Now().Unix() + DEFAULT_DAYS_INTERVAL*24*60*60
	} else {
		to, err = strconv.ParseInt(toStr, 10, 64)
		if err != nil {
			return 0, 0
		}
	}

	return from, to
}

func (ed *EventsDelivery) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	from, to := getFromToCgies(r.URL.Query())
	if from == 0 || to == 0 {
		ed.logger.Warnln("[GetAllEvents] could not parse from to cgies")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "could not parse from to cgies"}`))
		return
	}

	login := r.URL.Query().Get("login")
	if login == "" {
		login = r.Context().Value(middleware.ContextUserKey).(*model.User).Login
	}

	events, err := ed.eventUsecase.GetAllEvents(login, from, to)
	if err != nil {
		ed.logger.Warnf("[GetAllEvents] events not found: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(model.ToBytes(events.ToAnswer(true)))
}

func (ed *EventsDelivery) RemoveEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventId := vars["event_id"]
	usr := r.Context().Value(middleware.ContextUserKey).(*model.User)

	err := ed.eventUsecase.RemoveEvent(eventId, usr.Login)
	if err != nil {
		ed.logger.Warnf("[RemoveEvent] event not found: %s", err.Error())
		switch err {
		case errors.EventNotFound:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(errors.ErrorToBytes(err)))
		case errors.HasNoRights:
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(errors.ErrorToBytes(err)))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(`{"message": "ok", "event_id": "%s"}`, eventId)))
}

func (ed *EventsDelivery) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	eventId := mux.Vars(r)["event_id"]
	usr := r.Context().Value(middleware.ContextUserKey).(*model.User)
	err := ed.eventUsecase.AcceptInvite(eventId, usr.Login)
	if err != nil {
		ed.logger.Warnf("[AcceptInvite] event not found: %s", err.Error())
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
	w.Write([]byte(fmt.Sprintf(`{"message": "ok", "event_id": "%s"}`, eventId)))
}

func parseEventUserQuery(r *http.Request) (string, string) {
	cgis := mux.Vars(r)
	for k, v := range cgis {
		switch k {
		case string(model.EventCgi):
			return v, model.EventCgi
		}
	}
	return "", model.NilCgi
}

func (ed *EventsDelivery) GetInvites(w http.ResponseWriter, r *http.Request) {
	cgi, cgi_type := parseEventUserQuery(r)
	usr := r.Context().Value(middleware.ContextUserKey).(*model.User)

	invites, err := ed.eventUsecase.GetInvites(cgi, cgi_type, usr.Login)
	if err != nil {
		ed.logger.Warnf("[GetInvites] GetInvites: %s", err.Error())
		switch err {
		case errors.BadInviteCgi:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(errors.ErrorToBytes(err)))
		case errors.InviteNotFound:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "ok", "invites": []}`))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(200)
	w.Write(model.ToBytes(invites.ToAnswer()))
}

func (ed *EventsDelivery) RejectInvite(w http.ResponseWriter, r *http.Request) {
	eventId := mux.Vars(r)["event_id"]
	usr := r.Context().Value(middleware.ContextUserKey).(*model.User)
	err := ed.eventUsecase.RejectInvite(eventId, usr.Login)
	if err != nil {
		ed.logger.Warnf("[RejectInvite]: %s", err.Error())
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
	w.Write([]byte(`{"message": "ok"}`))
}
