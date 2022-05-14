package model

import (
	"sort"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

func (e *Event) Copy() *Event {
	return &Event{
		Id:            e.Id,
		Title:         e.Title,
		Description:   e.Description,
		Timestamp:     e.Timestamp,
		Members:       e.Members,
		ActiveMembers: e.ActiveMembers,
		Author:        e.Author,
		IsRegular:     e.IsRegular,
		Delta:         e.Delta,
	}
}

func (e *Event) ToAnswer() interface{} {
	hm := make(map[string]interface{}, 0)
	hm["message"] = "ok"
	hm["event"] = e
	return hm
}

func (e *Event) ToRegular(single_event_id string) *RegularEvent {
	return &RegularEvent{
		Id:            e.Id,
		Title:         e.Title,
		Description:   e.Description,
		Timestamp:     e.Timestamp,
		Members:       e.Members,
		ActiveMembers: e.ActiveMembers,
		Author:        e.Author,
		Delta:         e.Delta,
		SingleEventId: single_event_id,
	}
}

func (e *Event) ToSingle(regular_event_id string) *SingleEvent {
	return &SingleEvent{
		Id:             e.Id,
		Title:          e.Title,
		Description:    e.Description,
		Timestamp:      e.Timestamp,
		Members:        e.Members,
		ActiveMembers:  e.ActiveMembers,
		Author:         e.Author,
		RegularEventId: regular_event_id,
	}
}

type RegularEvent struct {
	Id            string
	Title         string   `bson:"title"`
	Description   string   `bson:"description"`
	Timestamp     int64    `bson:"timestamp"`
	Members       []string `bson:"members"`
	ActiveMembers []string `bson:"active_members"`
	Author        string   `bson:"author"`
	Delta         int64    `bson:"delta"`
	SingleEventId string   `bson:"single_event_id"`
}

func (re *RegularEvent) ToEvent() *Event {
	return &Event{
		Id:            re.Id,
		Title:         re.Title,
		Description:   re.Description,
		Timestamp:     re.Timestamp,
		Members:       re.Members,
		ActiveMembers: re.ActiveMembers,
		Author:        re.Author,
		Delta:         re.Delta,
		IsRegular:     true,
	}
}

type SingleEvent struct {
	Id             string
	Title          string   `bson:"title"`
	Description    string   `bson:"description"`
	Timestamp      int64    `bson:"timestamp"`
	Members        []string `bson:"members"`
	ActiveMembers  []string `bson:"active_members"`
	Author         string   `bson:"author"`
	RegularEventId string   `bson:"regular_event_id"`
}

func (re *SingleEvent) ToEvent() *Event {
	return &Event{
		Id:            re.Id,
		Title:         re.Title,
		Description:   re.Description,
		Timestamp:     re.Timestamp,
		Members:       re.Members,
		ActiveMembers: re.ActiveMembers,
		Author:        re.Author,
		Delta:         0,
		IsRegular:     false,
	}
}

type BsonRegularEvent struct {
	Id     string                 `bson:"_id"`
	Events map[string]interface{} `bson:"regular"`
}

type BsonSingleEvent struct {
	Id     string                 `bson:"_id"`
	Events map[string]interface{} `bson:"single"`
}

func ConvertInterfaceToEvent(event interface{}, mode string) *Event {
	ievent := event.(map[string]interface{})
	e := &Event{
		Id:          ievent["id"].(string),
		Title:       ievent["title"].(string),
		Description: ievent["description"].(string),
		Timestamp:   ievent["timestamp"].(int64),
		Author:      ievent["author"].(string),
	}
	e.Members = make([]string, 0)
	for _, val := range ievent["members"].(primitive.A) {
		e.Members = append(e.Members, val.(string))
	}
	e.ActiveMembers = make([]string, 0)
	for _, val := range ievent["active_members"].(primitive.A) {
		e.ActiveMembers = append(e.ActiveMembers, val.(string))
	}
	switch mode {
	case REGULAR_EVENT:
		e.Delta = ievent["delta"].(int64)
		e.IsRegular = true
		return e
	case SINGLE_EVENT:
		e.Delta = 0
		e.IsRegular = false
		return e
	}
	return nil
}

type JsonEvents struct {
	Events []*Event `json:"events"`
}

func (je *JsonEvents) ToAnswer(sorted bool) interface{} {
	hm := make(map[string]interface{}, 0)
	hm["message"] = "ok"
	if sorted {
		sort.Slice(je.Events, func(i, j int) bool {
			return je.Events[i].Timestamp < je.Events[j].Timestamp
		})
	}

	hm["events"] = je.Events
	return hm
}

type EventIdsStruct struct {
	Id       string   `bson:"_id"`
	EventIds []string `bson:"keys"`
}
