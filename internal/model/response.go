package model

import "encoding/json"

func ToBytes(body interface{}) []byte {
	buf, err := json.Marshal(body)
	if err != nil {
		return []byte(`{"message": "cannot marshal"}`)
	}
	return buf
}

type ResponseUser struct {
	Message string              `json:"message"`
	User    UserWithoutPassword `json:"user"`
}
