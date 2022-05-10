package usecase

import (
	"nocalendar/internal/app/errors"
	"nocalendar/internal/app/events"
	"nocalendar/internal/model"
	"nocalendar/internal/util"
	"strconv"

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
	event.ActiveMembers = addAuthorToMembers(event.ActiveMembers, author)
	event.Id = util.GenerateRandomString(32)

	_, err := eu.repo.InsertEvent(event, false, false)
	if err != nil {
		return "", err
	}

	err = eu.addInvites(event, false, true)
	if err != nil {
		return "", err
	}

	return event.Id, nil
}

func copyEvent(old_event, new_event *model.Event) {
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

	new_event.Author = old_event.Author
}

func mergeEvents(old_event *model.BsonEvent, new_event *model.Event, affectMeta bool) {
	if affectMeta {
		copyEvent(&old_event.Meta, new_event)
	} else {
		copyEvent(&old_event.Actual, new_event)
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

func (eu *EventsUsecase) addInvites(event *model.Event, reinvite bool, meta bool) error {
	for _, member := range event.Members {
		if event.Author == member {
			continue
		}

		inv := &model.Invite{
			EventId:  event.Id,
			Login:    member,
			Accepted: true,
			Meta:     meta,
		}

		if reinvite {
			err := eu.repo.RemoveInvite(inv)
			if err != nil {
				return err
			}
		} else {
			err := eu.repo.CheckInvite(inv)
			switch err {
			case nil: // invite already accepted
				continue
			case errors.InviteNotFound:
			default:
				return err
			}
		}

		inv.Accepted = false
		err := eu.repo.InsertInvite(inv)
		if err != nil {
			return err
		}
	}
	return nil
}

func (eu *EventsUsecase) removeInvites(event *model.Event) error {
	for _, member := range event.Members {
		inv := &model.Invite{
			EventId:  event.Id,
			Login:    member,
			Accepted: false,
		}

		err := eu.repo.RemoveInvite(inv)
		if err == errors.InternalError {
			return err
		}

		inv.Accepted = true
		err = eu.repo.RemoveInvite(inv)
		if err == errors.InternalError {
			return err
		}
	}
	return nil
}

func (eu *EventsUsecase) EditEvent(event *model.Event, login string, affectMeta bool) (*model.BsonEvent, error) {
	old_event_version, err := eu.GetEvent(event.Id, login)
	if err != nil {
		return nil, err
	}

	if !isParticipant(old_event_version.Actual.Members, login) {
		return nil, errors.HasNoRights
	}

	old_ts := old_event_version.Actual.Timestamp
	mergeEvents(old_event_version, event, affectMeta)

	_, err = eu.repo.InsertEvent(event, !affectMeta, false)
	if err != nil {
		return nil, err
	}

	if old_ts != event.Timestamp {
		if affectMeta {
			eu.removeInvites(event)
		}
		err = eu.addInvites(event, true, affectMeta)
		if err != nil {
			return nil, err
		}
	}

	return eu.GetEvent(event.Id, login)
}

func (eu *EventsUsecase) GetEvent(eventId string, login string) (*model.BsonEvent, error) {
	return eu.repo.GetEvent(eventId)
}

func (eu *EventsUsecase) GetAllEvents(login string, from, to int64) (*model.JsonEvent, error) {
	eventIds, err := eu.repo.GetEventsIdsByLogin(login)
	switch err {
	case nil:
		break
	case errors.MemberNotFound:
		return &model.JsonEvent{}, nil
	default:
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

		events.Events[eventId] = ev.Copy()
		if ev.Meta.Delta == 0 || !ev.Meta.IsRegular {
			continue
		}

		ts := ev.Meta.Timestamp
		idx := int64(0)
		for {
			ev.Meta.Timestamp += ev.Actual.Delta * 24 * 60 * 60
			idx += 1
			if ev.Meta.Timestamp > to {
				break
			}
			events.Events[eventId+strconv.FormatInt(idx, 10)] = ev.Copy()
		}

		ev.Meta.Timestamp = ts
		idx = 0
		for {
			ev.Meta.Timestamp -= ev.Actual.Delta * 24 * 60 * 60
			idx -= 1
			if ev.Meta.Timestamp < from {
				break
			}
			events.Events[eventId+strconv.FormatInt(idx, 10)] = ev.Copy()
		}
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

func (eu *EventsUsecase) AcceptInvite(event_id, login string, meta bool) error {
	inv := &model.Invite{
		EventId:  event_id,
		Login:    login,
		Accepted: false,
		Meta:     meta,
	}

	err := eu.repo.RemoveInvite(inv)
	if err != nil {
		return err
	}

	// err = eu.repo.InsertInvite(inv)
	event, err := eu.repo.GetEvent(event_id)
	if err != nil {
		return err
	}

	if meta {
		for _, member := range event.Meta.ActiveMembers {
			if member == login {
				return nil
			}
		}
		event.Meta.ActiveMembers = append(event.Meta.ActiveMembers, login)
	} else {
		for _, member := range event.Actual.ActiveMembers {
			if member == login {
				return nil
			}
		}
		event.Actual.ActiveMembers = append(event.Actual.ActiveMembers, login)
	}

	if meta && event.Actual.IsRegular { // if edit meta and actual event is regular
		_, err = eu.repo.InsertEvent(&event.Meta, false, false)
	} else if meta { // other case - edit only regular part of event
		_, err = eu.repo.InsertEvent(&event.Meta, false, true)
	} else { // edit only actual event
		_, err = eu.repo.InsertEvent(&event.Actual, true, false)
	}
	return err
}
