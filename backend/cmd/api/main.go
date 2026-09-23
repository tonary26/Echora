package main

import (
	"context"
	"echora/internal/database"
	"echora/internal/handlers"
	"echora/internal/middleware"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Проблема с подгружением .env")
	}

	ctx := context.Background()
	databaseUrl := os.Getenv("DB_URL")
	serverPort := os.Getenv("APP_PORT")

	log.Printf("Сервер %s запускается..", serverPort)

	db, err := database.Connect(ctx, databaseUrl)
	if err != nil {
		log.Printf("Ошибка запуска сервера %s", serverPort)
		log.Fatal()
	}
	defer db.Close()

	log.Println("Сервер успешно подключен к бд")

	authStore := database.NewAuthStore(db)
	authHandler := handlers.NewAuthHandler(authStore)

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/register", middleware.MethodHandler(authHandler.Register, http.MethodPost))
	mux.HandleFunc("/auth/login", middleware.MethodHandler(authHandler.Login, http.MethodPost))
	mux.HandleFunc("/auth/me", middleware.MethodHandler(middleware.AuthMiddleware(authHandler.CurrentUser), http.MethodGet))

	trackStore := database.NewTrackStore(db)

	fileStore, err := database.NewFileStore(ctx,
		os.Getenv("S3_ENDPOINT"), os.Getenv("S3_PUBLIC_ENDPOINT"),
		os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY"), os.Getenv("S3_BUCKET"))
	if err != nil {
		log.Fatal(err)
	}

	trackHandler := handlers.NewTrackHandler(trackStore, fileStore)
	mux.HandleFunc("/tracks/", middleware.MethodHandler(middleware.AuthMiddleware(trackHandler.Upload), http.MethodPost))
	mux.HandleFunc("/tracks/create", middleware.MethodHandler(middleware.AuthMiddleware(trackHandler.GetUserTracks), http.MethodGet))
	mux.HandleFunc("/tracks/{id}/stream-url", middleware.MethodHandler(middleware.AuthMiddleware(trackHandler.StreamURL), http.MethodGet))

	corsMux := middleware.CORSMiddleware(mux)
	loggedMux := middleware.LoggingMiddleware(corsMux)
	serverAddr := ":" + serverPort

	log.Printf("Сервер %s запущен", serverPort)
	err = http.ListenAndServe(serverAddr, loggedMux)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
}
