package main

import (
	"context"
	"fmt"
	"log"
	"time"

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
		AllowOrigins: "http://localhost:4200, http://localhost:3000",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type",
	}))
	app.Post("/event", handleEventPost)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Insight Garden Backend API is alive and listening")
	})
	app.Get("/session/:id", handleSessionGet)
	app.Listen(":3000")
}

type Event struct {
	SessionID string `json:"session_id"`
	Type      string `json:"type"`
	Page      string `json:"page"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Timestamp time.Time `json:"timestamp"`	// ISO string for now
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

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status": "created",
	})
}

func handleSessionGet(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	rows, err := db.Query(context.Background(),
		`SELECT session_id, type, page, x, y, timestamp
				FROM events
				WHERE session_id = $1
				ORDER BY timestamp ASC`,
		sessionID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("DB error: " + err.Error())
	}
	defer rows.Close()

	var events[] Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.SessionID, &e.Type, &e.Page, &e.X, &e.Y, &e.Timestamp); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("DB scan error: " + err.Error())
		}
		events = append(events, e)
	}
	return c.JSON(events)
}