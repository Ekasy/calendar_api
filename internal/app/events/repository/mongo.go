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

func (er *EventsRepository) InsertEvent(event *model.Event, is_actual bool) (*model.Event, error) {
	var err error
	if is_actual {
		err = er.updateActualEvent(event)
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
