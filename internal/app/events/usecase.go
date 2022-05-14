package events

import "nocalendar/internal/model"

type EventsUsecase interface {
	CreateEvent(event *model.Event, author string) (string, error)
	EditEvent(event *model.Event, login string) (*model.Event, error)
	GetEvent(eventId string, login string) (*model.Event, error)
	GetAllEvents(login string, from, to int64) (*model.JsonEvents, error)
	RemoveEvent(eventId, login string) error

	AcceptInvite(event_id, login string) error
	GetInvites(cgi string, cgi_type string, login string) (*model.InviteJson, error)
}
