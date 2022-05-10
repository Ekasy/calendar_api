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
	HasNoRights        *Error = &Error{Message: "user has no rights to access this resource"}

	EventNotFound  *Error = &Error{Message: "event not found"}
	EventNotEdited *Error = &Error{Message: "event not edited"}

	MemberNotFound *Error = &Error{Message: "user has not events"}

	InviteNotFound      *Error = &Error{Message: "invite has not found"}
	InviteAlreadyExists *Error = &Error{Message: "invite already exists"}
	FoundManyInvites    *Error = &Error{Message: "more than one invite was found"}

	InternalError *Error = &Error{Message: "something went wrong"}
)
