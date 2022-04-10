package errors

import "fmt"

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func ErrorToBytes(err error) []byte {
	return []byte(fmt.Sprintf(`{"message": "%s"}`, err.Error()))
}

var (
	UserNotFound       *Error = &Error{Message: "user not found"}
	BadPassword        *Error = &Error{Message: "incorrect password"}
	LoginAlreadyExists *Error = &Error{Message: "user with this login already exists"}
	EmailAlreadyExists *Error = &Error{Message: "user with this email already exists"}
	InternalError      *Error = &Error{Message: "something went wrong"}
)
