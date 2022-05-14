package model

// all possible cgies
const (
	EventCgi string = "event_id"
	NilCgi   string = "nil" // plug
	FromCgi  string = "from"
	ToCgi    string = "to"
)

// consts for access to mongo document
const (
	SINGLE_EVENT  string = "single"
	REGULAR_EVENT string = "regular"
)

// handy constants
const (
	DAYS_IN_SECONDS    int64 = 24 * 60 * 60
	LENGTH_OF_EVENT_ID int   = 32
)
