package db

import (
	"context"
	"nocalendar/internal/app/errors"
	"os"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	Conn   *mongo.Collection
	Ctx    context.Context
	logger *logrus.Logger
}

func (d *Database) find(body bson.M, opts *options.FindOneOptions) (bool, error) {
	res := d.Conn.FindOne(d.Ctx, body, opts)
	switch res.Err() {
	case nil:
		break
	case mongo.ErrNoDocuments:
		return false, nil
	default:
		return false, res.Err()
	}
	data := make(map[string]interface{})
	err := res.Decode(data)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *Database) insert(body bson.M) error {
	_, err := d.Conn.InsertOne(d.Ctx, body)
	return err
}

func (d *Database) initCollections() error {
	opts := options.FindOne()
	opts.SetProjection(bson.M{"users.nocalender_user_init.id": 1})
	exist, err := d.find(bson.M{"_id": "json/users"}, opts)
	if err != nil {
		d.logger.Warnf("[initCollections] find user: %s", err.Error())
		return errors.InternalError
	}

	if !exist {
		err = d.insert(bson.M{
			"_id": "json/users",
			"users": bson.M{
				"nocalender_user_init": bson.M{
					"id": 1,
				},
			},
		})
		if err != nil {
			d.logger.Warnf("[initCollections] insert user: %s", err.Error())
			return errors.InternalError
		}
	}

	opts = options.FindOne()
	opts.SetProjection(bson.M{"tokens.nocalender_token_init": 1})
	exist, err = d.find(bson.M{"_id": "json/tokens"}, opts)
	if err != nil {
		d.logger.Warnf("[initCollections] find token: %s", err.Error())
		return errors.InternalError
	}

	if !exist {
		err = d.insert(bson.M{
			"_id": "json/tokens",
			"tokens": bson.M{
				"nocalender_token_init": 1,
			},
		})
		if err != nil {
			d.logger.Warnf("[initCollections] insert token: %s", err.Error())
			return errors.InternalError
		}
	}

	opts = options.FindOne()
	opts.SetProjection(bson.M{"events.nocalender_event_init.id": 1})
	exist, err = d.find(bson.M{"_id": "json/events"}, opts)
	if err != nil {
		d.logger.Warnf("[initCollections] find event: %s", err.Error())
		return errors.InternalError
	}

	if !exist {
		err = d.insert(bson.M{
			"_id": "json/events",
			"regular": bson.M{
				"nocalender_regular_event_init": bson.M{
					"id": "1",
				},
			},
			"single": bson.M{
				"nocalender_single_event_init": bson.M{
					"id": "1",
				},
			},
		})
		if err != nil {
			d.logger.Warnf("[initCollections] insert event: %s", err.Error())
			return errors.InternalError
		}
	}

	opts = options.FindOne()
	opts.SetProjection(bson.M{"members.nocalender_member_init": 1})
	exist, err = d.find(bson.M{"_id": "json/members"}, opts)
	if err != nil {
		d.logger.Warnf("[initCollections] find member: %s", err.Error())
		return errors.InternalError
	}

	if !exist {
		err = d.insert(bson.M{
			"_id": "json/members",
			"members": bson.M{
				"nocalender_member_init": bson.A{"nocalender_regular_event_init"},
			},
		})
		if err != nil {
			d.logger.Warnf("[initCollections] insert member: %s", err.Error())
			return errors.InternalError
		}
	}

	opts = options.FindOne()
	opts.SetProjection(bson.M{"invites": 1})
	exist, err = d.find(bson.M{"_id": "json/invites"}, opts)
	if err != nil {
		d.logger.Warnf("[initCollections] find member: %s", err.Error())
		return errors.InternalError
	}

	if !exist {
		err = d.insert(bson.M{
			"_id": "json/invites",
			"invites": bson.M{
				"nocalender_user_init": bson.A{"nocalender_single_event_init"},
			},
		})
		if err != nil {
			d.logger.Warnf("[initCollections] insert member: %s", err.Error())
			return errors.InternalError
		}
	}
	return nil
}

func NewDatabase(logger *logrus.Logger) *Database {
	mongo_url := os.Getenv("MONGO_URL")
	mongo_db := os.Getenv("MONGO_DB")
	mongo_collection := os.Getenv("MONGO_COLLECTION")
	if mongo_url == "" || mongo_db == "" || mongo_collection == "" {
		logger.Fatalln("[NewDatabase] cannot get env MONGO_URL")
	}
	ctx := context.TODO()
	clientOptions := options.Client().ApplyURI(mongo_url)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Fatalf("[NewDatabase] cannot connect to mongo: %s", err.Error())
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Fatalf("[NewDatabase] cannot ping to mongo: %s", err.Error())
	}

	collection := client.Database(mongo_db).Collection(mongo_collection)
	db := &Database{
		Conn:   collection,
		Ctx:    ctx,
		logger: logger,
	}

	err = db.initCollections()
	if err != nil {
		logger.Fatalln("[NewDatabase] cannot init collections")
	}
	return db
}
