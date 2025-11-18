package main

import (
	"crowdsourcedurbanissuereportingwithai/backend/internal/auth"
	"crowdsourcedurbanissuereportingwithai/backend/internal/handlers"
	"net/http"

	"github.com/redis/go-redis/v9"
)

// RegisterRoutes registers public and protected routes. The report route is
// protected by the provided Auth middleware.
func RegisterRoutes(feedHandler *handlers.FeedHandler, reportHandler *handlers.ReportHandler, authHandler *handlers.AuthHandler, jwtAuth *auth.JWTService, rdb *redis.Client) {
	http.HandleFunc("/feed", feedHandler.ServeFeed)
	http.HandleFunc("/login", authHandler.Login)
	http.HandleFunc("/register", authHandler.Register)

	// protect /report with AuthMiddleware
	authMw := auth.AuthMiddleware(jwtAuth, rdb)
	http.Handle("/report", authMw(http.HandlerFunc(reportHandler.ServeReport)))
	// Protected logout route
	http.Handle("/logout", authMw(http.HandlerFunc(authHandler.Logout)))
	// Comments and upvotes
	http.Handle("/comment", authMw(http.HandlerFunc(reportHandler.ServeComment)))
	http.Handle("/upvote", authMw(http.HandlerFunc(reportHandler.ServeUpvote)))
	
	// Admin: protect with AuthMiddleware + AdminMiddleware
	adminMw := auth.AdminMiddleware(http.HandlerFunc(reportHandler.ServeUpdateStatus))
	http.Handle("/api/admin/post-status", authMw(adminMw))
	adminMw2 := auth.AdminMiddleware(http.HandlerFunc(feedHandler.ServeAdminFeed))
	http.Handle("/api/admin/issues", authMw(adminMw2))
}
