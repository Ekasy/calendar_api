package repository

import (
	"fmt"
	"nocalendar/internal/app/errors"
	"nocalendar/internal/app/events"
	"nocalendar/internal/db"
	"nocalendar/internal/model"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EventsRepository struct {
	mongo  *db.Database
	logger *logrus.Logger
}

func NewEventsRepository(db *db.Database, logger *logrus.Logger) events.EventsRepository {
	return &EventsRepository{
		mongo:  db,
		logger: logger,
	}
}

func (er *EventsRepository) eventToBson(event interface{}) (*bson.M, error) {
	data, err := bson.Marshal(event)
	if err != nil {
		er.logger.Warnf("[requestToBson] cannot marshal request: %s", err.Error())
		return nil, errors.InternalError
	}

	doc := &bson.M{}
	err = bson.Unmarshal(data, doc)
	if err != nil {
		er.logger.Warnf("[requestToBson] cannot unmarshal: %s", err.Error())
		return nil, errors.InternalError
	}
	return doc, nil
}

func (er *EventsRepository) updateEvent(event *model.Event) error {
	filter := bson.M{
		"_id": "json/events",
	}

	doc, err := er.eventToBson(model.BsonEvent{
		Meta:   *event,
		Actual: *event,
	})
	if err != nil {
		return err
	}

	body := bson.M{
		"$set": bson.M{
			fmt.Sprintf("events.%s", event.Id): doc,
		},
	}

	_, err = er.mongo.Conn.UpdateOne(er.mongo.Ctx, filter, body)
	if err != nil {
		er.logger.Warnf("[insertEvent] UpdateOne: %s", err.Error())
		return errors.InternalError
	}
	return nil
}

func (er *EventsRepository) updateActualEvent(event *model.Event) error {
	event.Delta = 0
	event.IsRegular = false
	filter := bson.M{
		"_id": "json/events",
	}

	doc, err := er.eventToBson(event)
	if err != nil {
		return err
	}

	body := bson.M{
		"$set": bson.M{
			fmt.Sprintf("events.%s.actual", event.Id): doc,
		},
	}

	_, err = er.mongo.Conn.UpdateOne(er.mongo.Ctx, filter, body)
	if err != nil {
		er.logger.Warnf("[updateActualEvent] UpdateOne: %s", err.Error())
		return errors.InternalError
	}
	return nil
}

func (er *EventsRepository) updateMetaEvent(event *model.Event) error {
	event.Delta = 0
	event.IsRegular = false
	filter := bson.M{
		"_id": "json/events",
	}

	doc, err := er.eventToBson(event)
	if err != nil {
		return err
	}

	body := bson.M{
		"$set": bson.M{
			fmt.Sprintf("events.%s.meta", event.Id): doc,
		},
	}

	_, err = er.mongo.Conn.UpdateOne(er.mongo.Ctx, filter, body)
	if err != nil {
		er.logger.Warnf("[updateMetaEvent] UpdateOne: %s", err.Error())
		return errors.InternalError
	}
	return nil
}

func (er *EventsRepository) addEventToMember(members []string, eventId string) error {
	filter := bson.M{
		"_id": "json/members",
	}

	for _, member := range members {
		body := bson.M{
			"$addToSet": bson.M{
				fmt.Sprintf("members.%s", member): eventId,
			},
		}

		_, err := er.mongo.Conn.UpdateOne(er.mongo.Ctx, filter, body)
		if err != nil {
			er.logger.Warnf("[addEventToMember] UpdateOne: %s", err.Error())
			return errors.InternalError
		}
	}
	return nil
}

func (er *EventsRepository) InsertEvent(event *model.Event, only_actual bool, only_meta bool) (*model.Event, error) {
	var err error
	if only_actual {
		err = er.updateActualEvent(event)
	} else if only_meta {
		err = er.updateMetaEvent(event)
	} else {
		err = er.updateEvent(event)
	}

	if err != nil {
		return event, err
	}
	err = er.addEventToMember(event.Members, event.Id)
	return event, err
}

func (er *EventsRepository) GetEvent(eventId string) (*model.BsonEvent, error) {
	doc := &model.JsonEvent{}
	opts := options.FindOne()
	opts.SetProjection(bson.M{fmt.Sprintf("events.%s", eventId): 1})
	err := er.mongo.Conn.FindOne(er.mongo.Ctx, bson.M{"_id": "json/events"}, opts).Decode(doc)
	switch err {
	case nil:
		if _, ok := doc.Events[eventId]; !ok {
			return nil, errors.EventNotFound
		}
		event := doc.Events[eventId]
		return &event, err
	case mongo.ErrNoDocuments:
		return nil, errors.EventNotFound
	default:
		return nil, errors.InternalError
	}
}

func (er *EventsRepository) GetEventsIdsByLogin(login string) ([]string, error) {
	doc := &model.BsonMembers{}
	opts := options.FindOne()
	opts.SetProjection(bson.M{fmt.Sprintf("members.%s", login): 1})
	err := er.mongo.Conn.FindOne(er.mongo.Ctx, bson.M{"_id": "json/members"}, opts).Decode(doc)
	switch err {
	case nil:
		if len(doc.Members) == 0 {
			return nil, errors.MemberNotFound
		}
		return doc.Members[login], err
	case mongo.ErrNoDocuments:
		return nil, errors.MemberNotFound
	default:
		return nil, errors.InternalError
	}
}

func (er *EventsRepository) RemoveEvent(eventId string) error {
	filter := bson.M{
		"_id": "json/events",
	}

	body := bson.M{
		"$unset": bson.M{
			fmt.Sprintf("events.%s", eventId): "",
		},
	}

	_, err := er.mongo.Conn.UpdateOne(er.mongo.Ctx, filter, body)
	if err != nil {
		er.logger.Warnf("[RemoveEvent] UpdateOne: %s", err.Error())
		return errors.InternalError
	}
	return nil
}

func (er *EventsRepository) GetAllMembers() (map[string][]string, error) {
	filter := bson.M{
		"_id": "json/members",
	}

	doc := er.mongo.Conn.FindOne(er.mongo.Ctx, filter)
	if doc.Err() != nil {
		er.logger.Warnf("[GetAllMembers] FindOne: %s", doc.Err().Error())
		return nil, errors.InternalError
	}

	hm := &model.BsonMembers{}
	err := doc.Decode(hm)
	if err != nil {
		er.logger.Warnf("[GetAllMembers] Decode: %s", err.Error())
		return nil, errors.InternalError
	}

	return hm.Members, nil
}

func (er *EventsRepository) RemoveEventIdFromMember(login, eventId string) error {
	filter := bson.M{
		"_id": "json/members",
	}

	body := bson.M{
		"$pull": bson.M{
			fmt.Sprintf("members.%s", login): eventId,
		},
	}

	_, err := er.mongo.Conn.UpdateOne(er.mongo.Ctx, filter, body)
	if err != nil {
		er.logger.Warnf("[RemoveEventIdFromMember] UpdateOne: %s", err.Error())
		return errors.InternalError
	}
	return nil
}

func (er *EventsRepository) GetAllEventIds() ([]string, error) {
	step1 := bson.M{
		"$match": bson.M{
			"_id": "json/events",
		},
	}

	step2 := bson.M{
		"$project": bson.M{
			"events": bson.M{
				"$objectToArray": "$events",
			},
		},
	}

	step3 := bson.M{
		"$project": bson.M{
			"keys": "$events.k",
		},
	}

	pipeline := []bson.M{step1, step2, step3}
	cursor, err := er.mongo.Conn.Aggregate(er.mongo.Ctx, pipeline)
	if err != nil {
		er.logger.Warnf("[GetAllEventIds] Aggregate: %s", err.Error())
		return nil, errors.InternalError
	}
	defer cursor.Close(er.mongo.Ctx)

	doc := make([]bson.M, 0)
	err = cursor.All(er.mongo.Ctx, &doc)
	if err != nil {
		er.logger.Warnf("[GetAllEventIds] All: %s", err.Error())
		return nil, errors.InternalError
	}

	if len(doc) != 1 {
		er.logger.Warnf("[GetAllEventIds] returned non 1 document: %d", len(doc))
		return nil, errors.InternalError
	}

	eventIds := make([]string, 0)
	for _, eventId := range doc[0]["keys"].(bson.A) {
		eventIds = append(eventIds, eventId.(string))
	}

	return eventIds, nil
}

func (er *EventsRepository) InsertInvite(invite *model.Invite) error {
	filter := bson.M{
		"_id": "json/invites",
	}

	body := bson.M{
		"$addToSet": bson.M{
			"invites": invite,
		},
	}

	_, err := er.mongo.Conn.UpdateOne(er.mongo.Ctx, filter, body)
	if err != nil {
		er.logger.Warnf("[InsertInvite] UpdateOne: %s", err.Error())
		return errors.InternalError
	}
	return nil
}

func (er *EventsRepository) CheckInvite(invite *model.Invite) error {
	doc := &model.InviteBson{}
	doc.Invites = make([]model.Invite, 0)
	body := bson.M{
		"_id":     "json/invites",
		"invites": invite,
	}
	err := er.mongo.Conn.FindOne(er.mongo.Ctx, body).Decode(doc)
	switch err {
	case nil:
		switch len(doc.Invites) {
		case 0:
			return errors.InviteNotFound
		case 1:
			return nil
		default:
			return errors.FoundManyInvites
		}
	case mongo.ErrNoDocuments:
		return errors.InviteNotFound
	default:
		return errors.InternalError
	}
}

func (er *EventsRepository) RemoveInvite(invite *model.Invite) error {
	filter := bson.M{
		"_id": "json/invites",
	}

	body := bson.M{
		"$pull": bson.M{
			"invites": invite,
		},
	}

	_, err := er.mongo.Conn.UpdateOne(er.mongo.Ctx, filter, body)
	if err != nil {
		er.logger.Warnf("[RemoveInvite] UpdateOne: %s", err.Error())
		return errors.InternalError
	}
	return nil
}
