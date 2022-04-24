package events

import "nocalendar/internal/model"

type EventsUsecase interface {
	CreateEvent(event *model.Event, author string) (string, error)
	GetEvent(eventId string, login string) (*model.BsonEvent, error)
}
