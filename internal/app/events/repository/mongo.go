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

func (er *EventsRepository) InsertRegularEvent(event *model.RegularEvent, mode string) error {
	filter := bson.M{
		"_id": "json/events",
	}

	body := bson.M{
		"$set": bson.M{
			fmt.Sprintf("%s.%s", mode, event.Id): event,
		},
	}

	_, err := er.mongo.Conn.UpdateOne(er.mongo.Ctx, filter, body)
	if err != nil {
		er.logger.Warnf("[InsertRegularEvent] UpdateOne: %s", err.Error())
		return errors.InternalError
	}

	err = er.addEventToMember(event.Members, event.Id)
	return err
}

func (er *EventsRepository) InsertSingleEvent(event *model.SingleEvent, mode string) error {
	filter := bson.M{
		"_id": "json/events",
	}

	body := bson.M{
		"$set": bson.M{
			fmt.Sprintf("%s.%s", mode, event.Id): event,
		},
	}

	_, err := er.mongo.Conn.UpdateOne(er.mongo.Ctx, filter, body)
	if err != nil {
		er.logger.Warnf("[InsertSingleEvent] UpdateOne: %s", err.Error())
		return errors.InternalError
	}

	err = er.addEventToMember(event.Members, event.Id)
	return err
}

func (er *EventsRepository) getRegularEvent(eventId string) (interface{}, error) {
	doc := &model.BsonRegularEvent{}
	opts := options.FindOne()
	opts.SetProjection(bson.M{fmt.Sprintf("regular.%s", eventId): 1})
	err := er.mongo.Conn.FindOne(er.mongo.Ctx, bson.M{"_id": "json/events"}, opts).Decode(doc)
	switch err {
	case nil:
		if _, ok := doc.Events[eventId]; !ok {
			return nil, errors.EventNotFound
		}
		return doc.Events[eventId], nil
	case mongo.ErrNoDocuments:
		return nil, errors.EventNotFound
	default:
		return nil, errors.InternalError
	}
}

func (er *EventsRepository) getSingleEvent(eventId string) (interface{}, error) {
	doc := &model.BsonSingleEvent{}
	opts := options.FindOne()
	opts.SetProjection(bson.M{fmt.Sprintf("single.%s", eventId): 1})
	err := er.mongo.Conn.FindOne(er.mongo.Ctx, bson.M{"_id": "json/events"}, opts).Decode(doc)
	switch err {
	case nil:
		if _, ok := doc.Events[eventId]; !ok {
			return nil, errors.EventNotFound
		}
		return doc.Events[eventId], nil
	case mongo.ErrNoDocuments:
		return nil, errors.EventNotFound
	default:
		return nil, errors.InternalError
	}
}

func (er *EventsRepository) GetEvent(eventId string) (interface{}, string, error) {
	event, err := er.getRegularEvent(eventId)
	switch err {
	case nil:
		return event, model.REGULAR_EVENT, nil
	case errors.EventNotFound:
		break
	case errors.InternalError:
		return nil, "", err
	}

	event, err = er.getSingleEvent(eventId)
	switch err {
	case nil:
		return event, model.SINGLE_EVENT, nil
	case errors.EventNotFound:
		break
	case errors.InternalError:
		return nil, "", err
	}

	return nil, "", errors.EventNotFound
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

func (er *EventsRepository) RemoveEvent(eventId, mode string) error {
	filter := bson.M{
		"_id": "json/events",
	}

	body := bson.M{
		"$unset": bson.M{
			fmt.Sprintf("%s.%s", mode, eventId): "",
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

func (er *EventsRepository) InsertInvite(login, event_id string) error {
	filter := bson.M{
		"_id": "json/invites",
	}

	body := bson.M{
		"$addToSet": bson.M{
			fmt.Sprintf("invites.%s", login): event_id,
		},
	}

	_, err := er.mongo.Conn.UpdateOne(er.mongo.Ctx, filter, body)
	if err != nil {
		er.logger.Warnf("[InsertInvite] UpdateOne: %s", err.Error())
		return errors.InternalError
	}
	return nil
}

func (er *EventsRepository) CheckInvite(login, event_id string) error {
	doc := &model.InviteBson{}
	doc.Invites = make(map[string][]string, 0)
	body := bson.M{
		"_id": "json/invites",
	}

	opts := options.FindOne()
	opts.SetProjection(bson.M{fmt.Sprintf("invites.%s", login): 1})

	err := er.mongo.Conn.FindOne(er.mongo.Ctx, body).Decode(doc)
	switch err {
	case nil:
		if len(doc.Invites) == 0 {
			return errors.InviteNotFound
		} else {
			for key, value := range doc.Invites {
				if key == login {
					for _, v := range value {
						if v == event_id {
							return nil
						}
					}
				}
			}
			return errors.InviteNotFound
		}
	case mongo.ErrNoDocuments:
		return errors.InviteNotFound
	default:
		return errors.InternalError
	}
}

func (er *EventsRepository) RemoveInvite(login, event_id string) error {
	filter := bson.M{
		"_id": "json/invites",
	}

	body := bson.M{
		"$pull": bson.M{
			fmt.Sprintf("invites.%s", login): event_id,
		},
	}

	_, err := er.mongo.Conn.UpdateOne(er.mongo.Ctx, filter, body)
	if err != nil {
		er.logger.Warnf("[RemoveInvite] UpdateOne: %s", err.Error())
		return errors.InternalError
	}
	return nil
}

func (er *EventsRepository) GetInviteByLogin(login string) ([]string, error) {
	doc := &model.InviteBson{}
	doc.Invites = make(map[string][]string, 0)
	body := bson.M{
		"_id": "json/invites",
	}

	opts := options.FindOne()
	opts.SetProjection(bson.M{fmt.Sprintf("invites.%s", login): 1})

	err := er.mongo.Conn.FindOne(er.mongo.Ctx, body).Decode(doc)
	switch err {
	case nil:
		if len(doc.Invites) == 0 {
			return nil, errors.InviteNotFound
		} else {
			for key, value := range doc.Invites {
				if key == login {
					return value, nil
				}
			}
			return nil, errors.InviteNotFound
		}
	case mongo.ErrNoDocuments:
		return nil, errors.InviteNotFound
	default:
		return nil, errors.InternalError
	}
}
