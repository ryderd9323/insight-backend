package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jackc/pgx/v5"
)

var db *pgx.Conn

func main() {
	// Connect to Postgres
	var err error
	db, err = pgx.Connect(context.Background(), "postgres://ryder:secret@localhost:5433/insight")
	
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close(context.Background())
	
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:4200",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type",
	}))
	app.Post("/event", handleEventPost)
	app.Listen(":3000")
}

type Event struct {
	SessionID string `json:"session_id"`
	Type      string `json:"type"`
	Page      string `json:"page"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Timestamp string `json:"timestamp"`	// ISO string for now
}

func handleEventPost(c *fiber.Ctx) error {
	var event Event
	if err := c.BodyParser(&event); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid JSON")
	}

	_, err := db.Exec(context.Background(),
		"INSERT INTO events (session_id, type, page, x, y, timestamp) VALUES ($1, $2, $3, $4, $5, $6)",
		event.SessionID, event.Type, event.Page, event.X, event.Y, event.Timestamp,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Database error: %v", err))
	}

	return c.SendStatus(fiber.StatusCreated)
}