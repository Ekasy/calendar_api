package model

type BsonMembers struct {
	Id      string              `json:"_id" bson:"_id"`
	Members map[string][]string `json:"members" bson:"members"`
}
