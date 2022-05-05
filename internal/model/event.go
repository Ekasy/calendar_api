package model

import "sort"

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

func (be *BsonEvent) ToAnswer() map[string]interface{} {
	hm := make(map[string]interface{}, 0)
	hm["message"] = "ok"
	hm["event"] = be
	return hm
}

func (be *BsonEvent) Copy() BsonEvent {
	m := Event{
		Id:            be.Meta.Id,
		Title:         be.Meta.Title,
		Description:   be.Meta.Description,
		Timestamp:     be.Meta.Timestamp,
		Members:       be.Meta.Members,
		ActiveMembers: be.Meta.ActiveMembers,
		Delta:         be.Meta.Delta,
		Author:        be.Meta.Author,
		IsRegular:     be.Meta.IsRegular,
	}
	a := Event{
		Id:            be.Actual.Id,
		Title:         be.Actual.Title,
		Description:   be.Actual.Description,
		Timestamp:     be.Actual.Timestamp,
		Members:       be.Actual.Members,
		ActiveMembers: be.Actual.ActiveMembers,
		Delta:         be.Actual.Delta,
		Author:        be.Actual.Author,
		IsRegular:     be.Actual.IsRegular,
	}
	return BsonEvent{
		Meta:   m,
		Actual: a,
	}
}

func (je *JsonEvent) ToAnswer(sorted bool) map[string]interface{} {
	hm := make(map[string]interface{}, 0)
	hm["message"] = "ok"
	events := make([]BsonEvent, 0)
	for _, value := range je.Events {
		events = append(events, value)
	}

	if sorted {
		sort.Slice(events, func(i, j int) bool {
			return events[i].Actual.Timestamp < events[j].Actual.Timestamp
		})
	}

	hm["events"] = events
	return hm
}

type EventIdsStruct struct {
	Id       string   `bson:"_id"`
	EventIds []string `bson:"keys"`
}
