package events

import "nocalendar/internal/model"

type EventsRepository interface {
	InsertRegularEvent(event *model.RegularEvent, mode string) error
	InsertSingleEvent(event *model.SingleEvent, mode string) error

	GetEvent(eventId string) (interface{}, string, error)
	GetEventsIdsByLogin(login string) ([]string, error)
	RemoveEvent(eventId, mode string) error
	GetAllMembers() (map[string][]string, error)
	RemoveEventIdFromMember(login, eventId string) error
	GetAllEventIds() ([]string, error)

	InsertInvite(login, event_id string) error
	CheckInvite(login, event_id string) error
	RemoveInvite(login, event_id string) error
	GetInviteByLogin(login string) ([]string, error)
}
