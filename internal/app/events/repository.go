package events

import "nocalendar/internal/model"

type EventsRepository interface {
	InsertEvent(event *model.Event) (*model.Event, error)
	GetEvent(eventId string) (*model.BsonEvent, error)
	GetEventsIdsByLogin(login string) ([]string, error)
}
