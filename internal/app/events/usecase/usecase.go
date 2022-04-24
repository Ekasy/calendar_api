package usecase

import (
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

	_, err := eu.repo.InsertEvent(event)
	if err != nil {
		return "", err
	}

	return event.Id, nil
}

func (eu *EventsUsecase) GetEvent(eventId string, login string) (*model.BsonEvent, error) {
	return eu.repo.GetEvent(eventId)
}
