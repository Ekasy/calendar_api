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

func updateTimestamp(er events.EventsRepository, event_id string) {
	ievent, mode, err := er.GetEvent(event_id)
	if err != nil {
		badEventIds = append(badEventIds, event_id)
		return
	}

	sup_ev_id := ""
	switch mode {
	case model.REGULAR_EVENT:
		sup_ev_id = ievent.(model.RegularEvent).SingleEventId
	case model.SINGLE_EVENT:
		sup_ev_id = ievent.(model.SingleEvent).RegularEventId
	}
	event := model.ConvertInterfaceToEvent(ievent, mode)

	if event.Timestamp < currentTimestamp {
		switch mode {
		case model.REGULAR_EVENT:
			event.Timestamp += event.Delta * model.DAYS_IN_SECONDS
			err = er.InsertRegularEvent(event.ToRegular(sup_ev_id), mode)
			if err != nil {
				badEventIds = append(badEventIds, event_id)
			}
		case model.SINGLE_EVENT:
			err = er.RemoveEvent(event.Id)
			if err != nil {
				badEventIds = append(badEventIds, event_id)
			}
			e, _, err := er.GetEvent(sup_ev_id)
			if err != nil {
				badEventIds = append(badEventIds, event_id)
			}
			ev := e.(model.RegularEvent)
			ev.SingleEventId = ""
			err = er.InsertRegularEvent(&ev, mode)
			if err != nil {
				badEventIds = append(badEventIds, event_id)
			}
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
		if event_id == "nocalender_regular_event_init" || event_id == "nocalender_single_event_init" {
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
