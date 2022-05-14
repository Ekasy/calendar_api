package model

type InviteBson struct {
	Id      string              `json:"_id"`
	Invites map[string][]string `json:"invites"`
}

type InviteJson struct {
	Invites []string `json:"invites"`
}

func (ij *InviteJson) ToAnswer() interface{} {
	hm := make(map[string]interface{})
	hm["message"] = "ok"
	hm["invites"] = ij.Invites
	return hm
}
