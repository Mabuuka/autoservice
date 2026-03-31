package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"kursovaya/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func buildDatabaseURL(cfg config.Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)
}

func NewPostgresPool(cfg config.Config) *pgxpool.Pool {
	databaseURL := buildDatabaseURL(cfg)

	var pool *pgxpool.Pool
	var err error

	for attempt := 1; attempt <= 10; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		pool, err = pgxpool.New(ctx, databaseURL)
		if err == nil {
			err = pool.Ping(ctx)
		}

		cancel()

		if err == nil {
			log.Println("Database connection established")
			return pool
		}

		log.Printf("Database is not ready yet (attempt %d/10): %v", attempt, err)
		time.Sleep(3 * time.Second)
	}

	log.Fatalf("Could not connect to database: %v", err)
	return nil
}
