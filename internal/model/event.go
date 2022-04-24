package model

type Event struct {
	Id            string   `json:"id"`
	Title         string   `json:"title" bson:"title"`
	Description   string   `json:"description" bson:"description"`
	Timestamp     int64    `json:"timestamp" bson:"timestamp"`
	Members       []string `json:"members" bson:"members"`
	ActiveMembers []string `json:"active_members" bson:"active_members"`
	Author        string   `json:"author" bson:"author"`
	IsRegular     bool     `json:"is_regular" bson:"is_regular"`
	Delta         int64    `json:"delta" bson:"delta"`
}

type BsonEvent struct {
	Meta   Event `json:"meta" bson:"meta"`
	Actual Event `json:"actual" bson:"actual"`
}

type BsonMetaEvent struct {
	Id   string `bson:"id"`
	Meta Event  `bson:"meta"`
}

type BsonActualEvent struct {
	Id     string `bson:"id"`
	Actual Event  `bson:"actual"`
}

type JsonEvent struct {
	Id     string               `json:"_id" bson:"_id"`
	Events map[string]BsonEvent `json:"events" bson:"events"`
}