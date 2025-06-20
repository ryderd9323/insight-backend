package main

import (
	"context"
	"log"
	"time"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Event struct {
	SessionID string  `json:"session_id"`
	Type      string  `json:"type"`
	Page      string  `json:"page"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Timestamp time.Time  `json:"timestamp"`
}

// Represents a point on the heatmap
type Dot struct {
	X	float64 `json:"x"`
	Y	float64	`json:"y"`
}

// Global variable to hold the database connection pool
var db *pgxpool.Pool

func main() {
	// Set up the Fiber app with CORS middleware
	app := fiber.New()
	app.Use(cors.New())
	/* app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:4200, http://localhost:3000",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type",
	})) */

	// Load database URL from environment or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://ryder:secret@localhost:5433/insight"
	}

	// Initialize pgx connection pool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	db = pool	// Assign to global variable
	
	// Routes and handlers
	app.Post("/event", handleEventPost)
	app.Get("/session/:id", handleSessionGet)
	app.Get("/heatmap/:page", handleHeatmapGet)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Insight Garden Backend API is alive and listening")
	})
	
	log.Fatal(app.Listen(":3000"))
}

// handleEventPost handles POST requests to /event
// Inserts a new interaction event into the database
func handleEventPost(c *fiber.Ctx) error {
	e := new(Event)

	// Parse the JSON body into the Event struct
	if err := c.BodyParser(&e); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Bad request")
	}

	// Insert the event into the events table
	_, err := db.Exec(context.Background(),
		`INSERT INTO events (session_id, type, page, x, y, timestamp)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		e.SessionID, e.Type, e.Page, e.X, e.Y, e.Timestamp)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("DB insert error: " + err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status": "created",
	})
}

// handleHeatmapGet handles GET requests to /heatmap/:page
// Returns (x,y) points aggregated from event data on the given page
func handleHeatmapGet(c *fiber.Ctx) error {
	page := c.Params("page")
	
	// Query for all (x,y) points and their count
	rows, err := db.Query(context.Background(),
		`SELECT x, y, COUNT(*) as count
		 FROM events
		 WHERE page = $1
		 GROUP BY x, y`,
		page,
	)

	if err != nil {
		log.Printf("Heatmap query error: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("DB query error")
	}
	defer rows.Close()

	// Scan through the rows and populate the dots array
	var dots[] Dot
	for rows.Next() {
		var d Dot
		if err := rows.Scan(&d.X, &d.Y); err != nil {
			log.Printf("Heatmap scan error: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Heatmap scan error")
		}
		dots = append(dots, d)
	}

	return c.JSON(dots)
}

// handleSessionGet handles GET requests to /session/:id
// Retrieves a chronological list of all events in the session
func handleSessionGet(c *fiber.Ctx) error {
	sessionID := c.Params("id")

	// Query events for given session ID
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
