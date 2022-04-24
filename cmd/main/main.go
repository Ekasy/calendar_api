package main

import (
	"net/http"
	ncldr_auth_delivery "nocalendar/internal/app/auth/delivery"
	ncldr_auth_repository "nocalendar/internal/app/auth/repository"
	ncldr_auth_usecase "nocalendar/internal/app/auth/usecase"
	ncldr_event_delivery "nocalendar/internal/app/events/delivery"
	ncldr_event_repository "nocalendar/internal/app/events/repository"
	ncldr_event_usecase "nocalendar/internal/app/events/usecase"
	"nocalendar/internal/app/middleware"
	ncldr_db "nocalendar/internal/db"
	ncldr_logger "nocalendar/internal/logger"

	"github.com/gorilla/mux"
)

func main() {
	logger := ncldr_logger.NewLogger()
	db := ncldr_db.NewDatabase(logger)

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.ContentTypeMiddleware)

	ar := ncldr_auth_repository.NewAuthRepository(db, logger)
	au := ncldr_auth_usecase.NewAuthUsecase(ar, logger)
	ad := ncldr_auth_delivery.NewAuthDelivery(au, logger)

	er := ncldr_event_repository.NewEventsRepository(db, logger)
	eu := ncldr_event_usecase.NewEventsUsecase(er, logger)
	ed := ncldr_event_delivery.NewEventsDelivery(eu, au, logger)

	ad.Routing(api)
	ed.Routing(api)

	logger.Infoln("start serving ::8000")
	err := http.ListenAndServe(":8000", r)
	logger.Errorf("http serve error: %s", err.Error())
}
