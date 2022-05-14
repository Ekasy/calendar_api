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
	event.ActiveMembers = addAuthorToMembers(event.ActiveMembers, author)
	event.Id = util.GenerateRandomString(model.LENGTH_OF_EVENT_ID)

	var err error
	if event.IsRegular {
		err = eu.repo.InsertRegularEvent(event.ToRegular(""), model.REGULAR_EVENT)
	} else {
		err = eu.repo.InsertSingleEvent(event.ToSingle(""), model.SINGLE_EVENT)
	}
	if err != nil {
		return "", err
	}

	err = eu.addInvites(event, false /* reinvite */)
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

func mergeEvents(old_event *model.Event, new_event *model.Event, mode string) {
	copyEvent(old_event, new_event)
	if mode == model.REGULAR_EVENT {
		if new_event.Delta == 0 {
			new_event.Delta = old_event.Delta
		}
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

func (eu *EventsUsecase) addInvites(event *model.Event, reinvite bool) error {
	for _, member := range event.Members {
		if event.Author == member {
			continue
		}

		if reinvite {
			err := eu.repo.RemoveInvite(member, event.Id)
			if err != nil {
				return err
			}
		} else {
			err := eu.repo.CheckInvite(member, event.Id)
			switch err {
			case nil: // invite already accepted
				continue
			case errors.InviteNotFound:
			default:
				return err
			}
		}

		err := eu.repo.InsertInvite(member, event.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (eu *EventsUsecase) removeInvites(event *model.Event) error {
	for _, member := range event.Members {
		err := eu.repo.RemoveInvite(member, event.Id)
		if err == errors.InternalError {
			return err
		}
	}
	return nil
}

func (eu *EventsUsecase) EditEvent(event *model.Event, login string) (*model.Event, error) {
	old_event_version, mode, err := eu.repo.GetEvent(event.Id)
	if err != nil {
		return nil, err
	}

	// copy single_event_id for regular event or regular_event_id for single event
	var sup_ev_id string
	if mode == model.REGULAR_EVENT {
		sup_ev_id = old_event_version.(map[string]interface{})["single_event_id"].(string)
	} else {
		sup_ev_id = old_event_version.(map[string]interface{})["regular_event_id"].(string)
	}

	oev := model.ConvertInterfaceToEvent(old_event_version, mode)

	if !isParticipant(oev.Members, login) {
		return nil, errors.HasNoRights
	}

	old_ts := oev.Timestamp
	mergeEvents(oev, event, mode)

	if event.IsRegular {
		err = eu.repo.InsertRegularEvent(event.ToRegular(sup_ev_id), model.REGULAR_EVENT)
	} else {
		if mode == model.REGULAR_EVENT {
			event.Id = util.GenerateRandomString(model.LENGTH_OF_EVENT_ID)
			err = eu.repo.InsertSingleEvent(event.ToSingle(model.ConvertInterfaceToEvent(old_event_version, mode).Id), model.SINGLE_EVENT)
			if err != nil {
				return nil, err
			}
			err = eu.repo.InsertRegularEvent(model.ConvertInterfaceToEvent(old_event_version, mode).ToRegular(event.Id), model.REGULAR_EVENT)
		} else {
			err = eu.repo.InsertSingleEvent(event.ToSingle(sup_ev_id), model.SINGLE_EVENT)
		}
	}
	if err != nil {
		return nil, err
	}

	if old_ts != event.Timestamp {
		if mode == model.REGULAR_EVENT {
			err = eu.removeInvites(event)
			if err != nil {
				return nil, err
			}
		}
		err = eu.addInvites(event, true /* reinvite */)
		if err != nil {
			return nil, err
		}
	}

	return eu.GetEvent(event.Id, login)
}

func (eu *EventsUsecase) GetEvent(eventId string, login string) (*model.Event, error) {
	event, mode, err := eu.repo.GetEvent(eventId)
	if err != nil {
		return nil, err
	}
	return model.ConvertInterfaceToEvent(event, mode), err
}

func (eu *EventsUsecase) GetAllEvents(login string, from, to int64) (*model.JsonEvents, error) {
	eventIds, err := eu.repo.GetEventsIdsByLogin(login)
	switch err {
	case nil:
		break
	case errors.MemberNotFound:
		return &model.JsonEvents{}, nil
	default:
		return nil, err
	}

	events := &model.JsonEvents{}
	events.Events = make([]*model.Event, 0)
	for _, eventId := range eventIds {
		ev, mode, err := eu.repo.GetEvent(eventId)
		if err != nil {
			continue
		}

		event := model.ConvertInterfaceToEvent(ev, mode)
		if mode == model.REGULAR_EVENT {
			if ev.(map[string]interface{})["single_event_id"].(string) == "" {
				if from < event.Timestamp || event.Timestamp < to {
					events.Events = append(events.Events, event.Copy())
				}
			}
			ts := event.Timestamp
			for {
				event.Timestamp += event.Delta * model.DAYS_IN_SECONDS
				if event.Timestamp > to {
					break
				}
				events.Events = append(events.Events, event.Copy())
			}

			event.Timestamp = ts
			for {
				event.Timestamp -= event.Delta * model.DAYS_IN_SECONDS
				if event.Timestamp < from {
					break
				}
				events.Events = append(events.Events, event.Copy())
			}
		} else if mode == model.SINGLE_EVENT {
			if from < event.Timestamp || event.Timestamp < to {
				events.Events = append(events.Events, event.Copy())
			}
		}

	}
	return events, nil
}

func (eu *EventsUsecase) RemoveEvent(eventId, login string) error {
	event, mode, err := eu.repo.GetEvent(eventId)
	if err != nil {
		return err
	}

	if event.(map[string]interface{})["author"].(string) != login {
		return errors.HasNoRights
	}

	return eu.repo.RemoveEvent(eventId, mode)
}

func (eu *EventsUsecase) AcceptInvite(event_id, login string) error {
	err := eu.repo.RemoveInvite(login, event_id)
	if err != nil {
		return err
	}

	// err = eu.repo.InsertInvite(inv)
	ievent, mode, err := eu.repo.GetEvent(event_id)
	if err != nil {
		return err
	}

	sup_ev_id := ""
	switch mode {
	case model.REGULAR_EVENT:
		sup_ev_id = ievent.(map[string]interface{})["single_event_id"].(string)
	case model.SINGLE_EVENT:
		sup_ev_id = ievent.(map[string]interface{})["regular_event_id"].(string)
	}
	event := model.ConvertInterfaceToEvent(ievent, mode)
	for _, member := range event.ActiveMembers {
		if member == login {
			return nil
		}
		event.ActiveMembers = append(event.ActiveMembers, login)
	}

	switch mode {
	case model.REGULAR_EVENT:
		err = eu.repo.InsertRegularEvent(event.ToRegular(sup_ev_id), mode)
	case model.SINGLE_EVENT:
		err = eu.repo.InsertSingleEvent(event.ToSingle(sup_ev_id), mode)
	default:
		err = errors.InternalError
	}
	return err
}

func (eu *EventsUsecase) GetInvites(cgi string, cgi_type string, login string) (*model.InviteJson, error) {
	invites, err := eu.repo.GetInviteByLogin(login)
	if err != nil {
		return nil, err
	}

	invs := &model.InviteJson{}
	invs.Invites = make([]string, 0)
	switch cgi_type {
	case model.EventCgi:
		for _, inv := range invites {
			if inv == cgi {
				invs.Invites = append(invs.Invites, inv)
				return invs, nil
			}
		}
		return nil, errors.InviteNotFound
	case model.NilCgi:
		invs.Invites = invites
		return invs, nil
	}
	return nil, errors.BadInviteCgi
}
