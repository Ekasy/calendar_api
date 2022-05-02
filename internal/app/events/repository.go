package events

import "nocalendar/internal/model"

type EventsRepository interface {
	InsertEvent(event *model.Event, is_actial bool) (*model.Event, error)
	GetEvent(eventId string) (*model.BsonEvent, error)
	GetEventsIdsByLogin(login string) ([]string, error)
	RemoveEvent(eventId string) error
	GetAllMembers() (map[string][]string, error)
	RemoveEventIdFromMember(login, eventId string) error
	GetAllEventIds() ([]string, error)
}
