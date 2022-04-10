package model

type Event struct {
	Id            string   `json:"id"`
	Title         string   `json:"title" bson:"title"`
	Description   string   `json:"descripton" bson:"description"`
	Timestamp     int64    `json:"timestamp" bson:"timestamp"`
	Members       []string `json:"members" bson:"members"`
	ActiveMembers []string `json:"active_members" bson:"active_members"`
	Author        string   `json:"author" bson:"author"`
	IsRegular     bool     `json:"is_regular" bson:"is_regular"`
	Delta         int64    `json:"delta" bson:"delta"`
}

type BsonEvent struct {
	Id     string `bson:"id"`
	Meta   Event  `bson:"event"`
	Actual Event  `bson:"actual"`
}

type BsonMetaEvent struct {
	Id   string `bson:"id"`
	Meta Event  `bson:"event"`
}

type BsonActualEvent struct {
	Id     string `bson:"id"`
	Actual Event  `bson:"actual"`
}
