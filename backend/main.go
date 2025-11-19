package main

import (
	"context"
	config "crowdsourcedurbanissuereportingwithai/backend/configs"
	"crowdsourcedurbanissuereportingwithai/backend/internal/auth"
	"crowdsourcedurbanissuereportingwithai/backend/internal/cache"
	"crowdsourcedurbanissuereportingwithai/backend/internal/handlers"
	"crowdsourcedurbanissuereportingwithai/backend/internal/repository"
	"crowdsourcedurbanissuereportingwithai/backend/internal/services"
	"crowdsourcedurbanissuereportingwithai/backend/models"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	config.LoadEnv()

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=post4321 dbname=Civicissue port=5432 sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	_, err = sqlDB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
	if err != nil {
		log.Fatal("Failed to enable uuid-ossp extension:", err)
	}

	// AutoMigrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Issue{},
		&models.Post{},
		&models.Comment{},
		&models.Upvote{},
	)
	if err != nil {
		log.Fatal(err)
	}

	postRepo := repository.NewPostRepository(db)
	feedService := services.NewFeedService(postRepo)
	reportService := services.NewReportService(postRepo)
	feedHandler := handlers.NewFeedHandler(feedService)
	reportHandler := handlers.NewReportHandler(reportService)
	mlHandler := handlers.NewMLHandler()

	jwtSvc := auth.NewJWTService()

	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo)

	
	redisAddr := config.GetRedisAddr()
	var redisClient *redis.Client
	if redisAddr != "" {
		redisClient = cache.NewRedisClient(redisAddr, config.GetRedisPassword())
	}

	authHandler := handlers.NewAuthHandler(authService, jwtSvc, redisClient)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	http.HandleFunc("/api/endpoint", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if d := r.URL.Query().Get("delay_ms"); d != "" {
			if ms, err := strconv.Atoi(d); err == nil && ms > 0 {
				time.Sleep(time.Duration(ms) * time.Millisecond)
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"endpoint":"/api/endpoint"}`))
	})

	frontendDir := os.Getenv("FRONTEND_DIR")
	if frontendDir == "" {
		candidates := []string{"./frontend", "../frontend", "./public", "../public"}
		found := ""
		for _, c := range candidates {
			idx := filepath.Join(c, "index.html")
			if _, err := os.Stat(idx); err == nil {
				found = c
				break
			}
		}
		if found != "" {
			frontendDir = found
		} else {

			frontendDir = "../frontend"
		}
	}

	fileSystem := http.Dir(frontendDir)
	fileServer := http.FileServer(fileSystem)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			idx := filepath.Join(frontendDir, "index.html")
			if _, err := os.Stat(idx); err == nil {
				http.ServeFile(w, r, idx)
				return
			}
			http.NotFound(w, r)
			return
		}
		
		if strings.HasSuffix(r.URL.Path, "/") {
			
			idxPath := path.Clean(r.URL.Path + "index.html")
			
			fsPath := filepath.Join(frontendDir, strings.TrimPrefix(idxPath, "/"))
			if _, err := os.Stat(fsPath); err != nil {
				http.NotFound(w, r)
				return
			}
			http.ServeFile(w, r, fsPath)
			return
		}
		
		fileServer.ServeHTTP(w, r)
	})
	http.HandleFunc("/feed", feedHandler.ServeFeed)
	http.HandleFunc("/login", authHandler.Login)
	http.HandleFunc("/register", authHandler.Register)
	
	http.HandleFunc("/google-login", authHandler.GoogleLogin)
	
	http.HandleFunc("/classify-image", mlHandler.ServeClassifyImage)
	http.HandleFunc("/predict-urgency", mlHandler.ServePredictUrgency)
	authMw := auth.AuthMiddleware(jwtSvc, redisClient)

	
	if strings.ToLower(os.Getenv("DISABLE_AUTH")) == "true" {
		devEmail := os.Getenv("DEV_TEST_USER_EMAIL")
		if devEmail == "" {
			devEmail = "dev@example.com"
		}
		devPass := os.Getenv("DEV_TEST_USER_PASSWORD")
		if devPass == "" {
			devPass = "devpass"
		}
		// try to get existing user or register one
		var devUser *models.User
		if u, err := userRepo.GetByEmail(devEmail); err == nil {
			devUser = u
		} else {
			u, err := authService.Register("Dev User", devEmail, devPass)
			if err != nil {
				log.Fatalf("failed to create dev test user: %v", err)
			}
			devUser = u
		}
		// set global dev user id used by handlers
		handlers.DevTestUserID = devUser.ID

		// register routes without auth middleware for convenience
		http.Handle("/report", http.HandlerFunc(reportHandler.ServeReport))
		http.Handle("/logout", http.HandlerFunc(authHandler.Logout))
		http.Handle("/comment", http.HandlerFunc(reportHandler.ServeComment))
		http.Handle("/upvote", http.HandlerFunc(reportHandler.ServeUpvote))
		log.Println("DISABLE_AUTH=true: auth disabled for local testing; using dev user:", devEmail)
	} else {
		http.Handle("/report", authMw(http.HandlerFunc(reportHandler.ServeReport)))
		http.Handle("/logout", authMw(http.HandlerFunc(authHandler.Logout)))
		// Comments and upvotes are protected endpoints — user must be authenticated
		http.Handle("/comment", authMw(http.HandlerFunc(reportHandler.ServeComment)))
		http.Handle("/upvote", authMw(http.HandlerFunc(reportHandler.ServeUpvote)))
	}

	// Log redis status
	if redisClient == nil {
		log.Println("Redis not configured; token revocation disabled")
	} else {
		log.Println("Redis configured; token revocation enabled")
	}

	// Allow overriding the listen port with the PORT env var (useful when :8080 is in use)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	// Build a CORS wrapper so frontend hosted on a different origin can call the API.
	// For local development if ALLOWED_ORIGIN is not set we allow '*' but do NOT
	// set credentials (browsers reject Access-Control-Allow-Credentials with '*').
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN") // e.g. https://your-app.vercel.app
	allowCredentials := false
	if allowedOrigin == "" {
		allowedOrigin = "*"
	} else {
		allowCredentials = true
	}

	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			if allowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: corsHandler,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server running on %s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited properly")
}
