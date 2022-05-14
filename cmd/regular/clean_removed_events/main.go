package main

import (
	"fmt"
	"nocalendar/internal/app/errors"
	"nocalendar/internal/app/events"
	ncldr_event_repository "nocalendar/internal/app/events/repository"
	ncldr_db "nocalendar/internal/db"
	ncldr_logger "nocalendar/internal/logger"
)

func removeUnusedEvents(er events.EventsRepository, member string, arr_events []string) bool {
	all_removed := true
	for _, event_id := range arr_events {
		_, _, err := er.GetEvent(event_id)
		if err == errors.EventNotFound {
			err = er.RemoveEventIdFromMember(member, event_id)
			if err != nil {
				fmt.Printf("Not removed event id: %s\nError: %s\n\n", event_id, err.Error())
				all_removed = false
			}
		}
	}
	return all_removed
}

func removeAllUnusedEvents(er events.EventsRepository) {
	members, err := er.GetAllMembers()
	if err != nil {
		fmt.Println(err.Error())
		panic(1)
	}

	all_removed := true
	for member, arr_events := range members {
		all_removed = all_removed && removeUnusedEvents(er, member, arr_events)
	}

	if !all_removed {
		panic(2)
	}
}

func main() {
	logger := ncldr_logger.NewLogger()
	db := ncldr_db.NewDatabase(logger)

	er := ncldr_event_repository.NewEventsRepository(db, logger)

	removeAllUnusedEvents(er)
}
