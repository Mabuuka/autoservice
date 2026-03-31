package main

import (
	"log"
	"net/http"

	"kursovaya/backend/internal/config"
	"kursovaya/backend/internal/database"
	"kursovaya/backend/internal/router"
)

func main() {
	cfg := config.Load()

	dbPool := database.NewPostgresPool(cfg)
	defer dbPool.Close()

	appRouter := router.New(cfg, dbPool)
	address := ":" + cfg.AppPort

	log.Printf("Server started on http://localhost:%s", cfg.AppPort)

	if err := http.ListenAndServe(address, appRouter); err != nil {
		log.Fatalf("Server stopped with error: %v", err)
	}
}
