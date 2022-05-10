package model

type Invite struct {
	EventId  string `json:"event_id" bson:"event_id"`
	Login    string `json:"login" bson:"login"`
	Accepted bool   `json:"accepted" bson:"accepted"`
	Meta     bool   `json:"meta" bson:"meta"`
}

type InviteBson struct {
	Id      string   `bson:"_id"`
	Invites []Invite `bson:"invites"`
}

type InviteJson struct {
	EventId string `json:"event_id"`
	Meta    bool   `json:"is_regular"`
}
