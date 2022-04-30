package usecase

import (
	"nocalendar/internal/app/errors"
	"nocalendar/internal/app/events"
	"nocalendar/internal/model"
	"nocalendar/internal/util"

	"github.com/sirupsen/logrus"
)

type EventsUsecase struct {
	repo   events.EventsRepository
	logger *logrus.Logger
}

func NewEventsUsecase(repo events.EventsRepository, logger *logrus.Logger) events.EventsUsecase {
	return &EventsUsecase{
		repo:   repo,
		logger: logger,
	}
}

func addAuthorToMembers(members []string, author string) []string {
	for _, member := range members {
		if member == author {
			return members
		}
	}
	return append(members, author)
}

func (eu *EventsUsecase) CreateEvent(event *model.Event, author string) (string, error) {
	event.Author = author
	event.Members = addAuthorToMembers(event.Members, author)
	event.ActiveMembers = event.Members
	event.Id = util.GenerateRandomString(32)

	_, err := eu.repo.InsertEvent(event, false)
	if err != nil {
		return "", err
	}

	return event.Id, nil
}

func copyEvent(old_event, new_event *model.Event, is_meta bool) {
	if new_event.Description == "" {
		new_event.Description = old_event.Description
	}

	if new_event.Title == "" {
		new_event.Title = old_event.Title
	}

	if new_event.Timestamp == 0 {
		new_event.Timestamp = old_event.Timestamp
	}

	if len(new_event.Members) == 0 {
		new_event.Members = old_event.Members
	}

	if len(new_event.ActiveMembers) == 0 {
		new_event.ActiveMembers = old_event.ActiveMembers
	}

	if is_meta && new_event.Delta == 0 {
		new_event.Delta = old_event.Delta
	}

	new_event.Author = old_event.Author
}

func mergeEvents(old_event *model.BsonEvent, new_event *model.Event, affectMeta bool) {
	if affectMeta {
		copyEvent(&old_event.Meta, new_event, true)
	} else {
		copyEvent(&old_event.Actual, new_event, false)
	}
}

func isParticipant(members []string, login string) bool {
	for _, member := range members {
		if member == login {
			return true
		}
	}
	return false
}

func (eu *EventsUsecase) EditEvent(event *model.Event, login string, affectMeta bool) (*model.BsonEvent, error) {
	old_event_version, err := eu.GetEvent(event.Id, login)
	if err != nil {
		return nil, err
	}

	if !isParticipant(old_event_version.Actual.Members, login) {
		return nil, errors.HasNoRights
	}

	mergeEvents(old_event_version, event, affectMeta)

	_, err = eu.repo.InsertEvent(event, !affectMeta)
	if err != nil {
		return nil, err
	}
	return eu.GetEvent(event.Id, login)
}

func (eu *EventsUsecase) GetEvent(eventId string, login string) (*model.BsonEvent, error) {
	return eu.repo.GetEvent(eventId)
}

func (eu *EventsUsecase) GetAllEvents(login string, from, to int64) (*model.JsonEvent, error) {
	eventIds, err := eu.repo.GetEventsIdsByLogin(login)
	if err != nil {
		return nil, err
	}

	events := &model.JsonEvent{}
	events.Events = make(map[string]model.BsonEvent, 0)
	for _, eventId := range eventIds {
		ev, err := eu.GetEvent(eventId, login)
		if err != nil {
			continue
		}

		if from > ev.Actual.Timestamp || ev.Actual.Timestamp > to {
			continue
		}

		events.Events[eventId] = *ev
	}
	return events, nil
}

func (eu *EventsUsecase) RemoveEvent(eventId, login string) error {
	event, err := eu.GetEvent(eventId, login)
	if err != nil {
		return err
	}

	if event.Meta.Author != login {
		return errors.HasNoRights
	}

	return eu.repo.RemoveEvent(eventId)
}
