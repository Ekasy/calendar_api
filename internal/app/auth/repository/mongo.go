package repository

import (
	"fmt"
	"nocalendar/internal/app/auth"
	"nocalendar/internal/app/errors"
	"nocalendar/internal/db"
	"nocalendar/internal/model"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuthRepository struct {
	mongo  *db.Database
	logger *logrus.Logger
}

func NewAuthRepository(mongo *db.Database, logger *logrus.Logger) auth.AuthRepository {
	return &AuthRepository{
		mongo:  mongo,
		logger: logger,
	}
}

func (ar *AuthRepository) userToBson(usr *model.User) (*bson.M, error) {
	data, err := bson.Marshal(usr)
	if err != nil {
		ar.logger.Warnf("[requestToBson] cannot marshal request: %s", err.Error())
		return nil, errors.InternalError
	}

	doc := &bson.M{}
	err = bson.Unmarshal(data, doc)
	if err != nil {
		ar.logger.Warnf("[requestToBson] cannot unmarshal: %s", err.Error())
		return nil, errors.InternalError
	}
	return doc, nil
}

func (ar *AuthRepository) Insert(usr *model.User) (*model.User, error) {
	filter := bson.M{
		"_id": "json/users",
	}

	doc, err := ar.userToBson(usr)
	if err != nil {
		return nil, err
	}

	body := bson.M{
		"$set": bson.M{
			fmt.Sprintf("users.%s", usr.Login): doc,
		},
	}

	_, err = ar.mongo.Conn.UpdateOne(ar.mongo.Ctx, filter, body)
	if err != nil {
		ar.logger.Warnf("[Insert] UpdateOne: %s", err.Error())
		return usr, errors.InternalError
	}
	return usr, nil
}

func (ar *AuthRepository) existEmail(email string) (bool, error) {
	step1 := bson.M{
		"$project": bson.M{
			"users": bson.M{
				"$objectToArray": "$users",
			},
		},
	}

	step2 := bson.M{
		"$match": bson.M{
			"users.v.email": email,
		},
	}

	step3 := bson.M{
		"$project": bson.M{
			"users": bson.M{
				"$arrayToObject": "$users",
			},
		},
	}

	pipeline := []bson.M{step1, step2, step3}
	cursor, err := ar.mongo.Conn.Aggregate(ar.mongo.Ctx, pipeline)
	if err != nil {
		ar.logger.Warnf("[checkEmail] Aggregate: %s", err.Error())
		return false, errors.InternalError
	}
	defer cursor.Close(ar.mongo.Ctx)

	doc := make([]*bson.M, 0)
	err = cursor.All(ar.mongo.Ctx, &doc)
	if err != nil {
		ar.logger.Warnf("[checkEmail] cursor.All: %s", err.Error())
		return false, errors.InternalError
	}

	if err := cursor.Err(); err != nil {
		ar.logger.Warnf("[checkEmail] cursor.Err: %s", err.Error())
		return false, errors.InternalError
	}

	if len(doc) == 0 {
		return false, nil
	}
	return true, nil
}

func (ar *AuthRepository) CheckUser(usr *model.User) (bool, error) {
	_, err := ar.GetUser(usr.Login)
	switch err {
	case nil:
		return false, errors.LoginAlreadyExists
	case errors.UserNotFound:
		break
	default:
		return false, errors.InternalError
	}

	exist, err := ar.existEmail(usr.Email)
	if err != nil {
		return false, err
	}

	if exist {
		return false, errors.EmailAlreadyExists
	}

	return true, nil
}

func (ar *AuthRepository) GetUser(login string) (*model.User, error) {
	doc := &model.JsonUser{}
	opts := options.FindOne()
	opts.SetProjection(bson.M{fmt.Sprintf("users.%s", login): 1})
	err := ar.mongo.Conn.FindOne(ar.mongo.Ctx, bson.M{"_id": "json/users"}, opts).Decode(doc)
	switch err {
	case nil:
		if len(doc.Users) == 0 {
			return nil, errors.UserNotFound
		}
		usr := doc.Users[login]
		return &usr, err
	case mongo.ErrNoDocuments:
		return nil, errors.UserNotFound
	default:
		return nil, errors.InternalError
	}
}
