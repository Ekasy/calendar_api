package main

import (
	"nocalendar/internal/app/events"
	ncldr_event_repository "nocalendar/internal/app/events/repository"
	ncldr_db "nocalendar/internal/db"
	ncldr_logger "nocalendar/internal/logger"
	"nocalendar/internal/model"
	"time"
)

var (
	logger           = ncldr_logger.NewLogger()
	badEventIds      = make([]string, 0)
	currentTimestamp = time.Now().Unix()
)

func copyActualByMetaParams(actual *model.Event, meta *model.Event) {
	actual.Title = meta.Title
	actual.Description = meta.Description
	actual.Timestamp = meta.Timestamp
	actual.Author = meta.Author
	actual.Members = meta.Members
	actual.ActiveMembers = meta.ActiveMembers
	actual.Delta = meta.Delta
	actual.IsRegular = meta.IsRegular
}

func updateTimestamp(er events.EventsRepository, event_id string) {
	event, err := er.GetEvent(event_id)
	if err != nil {
		badEventIds = append(badEventIds, event_id)
		return
	}

	// in case of meta event is not regular and actual was held -> update actual by meta params
	if event.Meta.IsRegular && event.Actual.Timestamp < currentTimestamp {
		copyActualByMetaParams(&event.Actual, &event.Meta)
	}

	if event.Meta.IsRegular && event.Meta.Timestamp < currentTimestamp {
		event.Meta.Timestamp = event.Meta.Timestamp + event.Meta.Delta
		_, err = er.InsertEvent(&event.Meta, false, false)
		if err != nil {
			badEventIds = append(badEventIds, event_id)
		}
	}
}

func updateAllTimestamps(er events.EventsRepository) {
	event_ids, err := er.GetAllEventIds()
	if err != nil {
		logger.Fatalln(err.Error())
	}

	for _, event_id := range event_ids {
		// skip initialized event id
		if event_id == "nocalender_event_init" {
			continue
		}

		updateTimestamp(er, event_id)
	}

	if len(badEventIds) > 0 {
		logger.Infoln(badEventIds)
		logger.Fatalln("not all events was updated")
	}
}

func main() {
	db := ncldr_db.NewDatabase(logger)

	er := ncldr_event_repository.NewEventsRepository(db, logger)

	updateAllTimestamps(er)
}
