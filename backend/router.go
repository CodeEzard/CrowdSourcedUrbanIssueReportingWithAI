package main

import (
	"crowdsourcedurbanissuereportingwithai/backend/internal/handlers"
	"net/http"
)

func RegisterRoutes(feedHandler *handlers.FeedHandler, reportHandler *handlers.ReportHandler) {
	http.HandleFunc("/feed", feedHandler.ServeFeed)
	http.HandleFunc("/report", reportHandler.ServeReport)
}
