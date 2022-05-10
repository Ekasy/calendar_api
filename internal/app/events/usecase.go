package events

import "nocalendar/internal/model"

type EventsUsecase interface {
	CreateEvent(event *model.Event, author string) (string, error)
	EditEvent(event *model.Event, login string, affectMeta bool) (*model.BsonEvent, error)
	GetEvent(eventId string, login string) (*model.BsonEvent, error)
	GetAllEvents(login string, from, to int64) (*model.JsonEvent, error)
	RemoveEvent(eventId, login string) error

	AcceptInvite(event_id, login string, is_meta bool) error
}
