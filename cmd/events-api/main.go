package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/DIMO-INC/events-api/internal/config"
	"github.com/DIMO-INC/events-api/internal/controllers"
	"github.com/DIMO-INC/events-api/internal/database"
	"github.com/DIMO-INC/events-api/models"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func main() {
	gitSha1 := os.Getenv("GIT_SHA1")
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "events-api").
		Str("git-sha1", gitSha1).
		Logger()

	settings, err := config.LoadConfig("settings.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("could not load settings")
	}
	pdb := database.NewDbConnectionFromSettings(ctx, settings)

	arg := ""
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}
	switch arg {
	case "migrate":
		migrateDatabase(logger, settings)
	default:
		startWebAPI(logger, settings, pdb, ctx)
	}
}

type EventMessage struct {
	Type    string          `json:"type"`
	Source  string          `json:"source"`
	Subject string          `json:"subject"`
	ID      string          `json:"id"`
	Time    time.Time       `json:"time"`
	Data    json.RawMessage `json:"data"`
}

func startWebAPI(logger zerolog.Logger, settings *config.Settings, pdb database.DbStore, ctx context.Context) {
	go func() {
		c, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":  "localhost:9092",
			"group.id":           "events-api",
			"auto.offset.reset":  "latest",
			"enable.auto.commit": false,
		})
		if err != nil {
			panic(err)
		}
		c.SubscribeTopics([]string{"events"}, nil)

		last := time.Now()
		var events []EventMessage
		partitions := make(map[kafka.Offset]kafka.TopicPartition)
		for {
			msg, err := c.ReadMessage(500 * time.Millisecond)
			if err == nil {
				var event EventMessage
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					fmt.Printf("JSON parsing error: %v (%v)\n", err, msg)
				} else {
					events = append(events, event)
					if p, ok := partitions[msg.TopicPartition.Offset]; ok {
						if p.Offset < msg.TopicPartition.Offset {
							p.Offset = msg.TopicPartition.Offset
						}
					} else {
						partitions[msg.TopicPartition.Offset] = p
					}
					if now := time.Now(); now.Sub(last) > 10*time.Second {
						func() {
							last = now
							tx, _ := pdb.DBS().Writer.BeginTx(ctx, nil)
							defer tx.Rollback()
							for _, event := range events {
								dbEvent := models.Event{
									Type:    event.Type,
									Source:  event.Source,
									Subject: event.Subject,
									ID:      event.ID,
									Time:    event.Time,
									Data:    null.JSONFrom([]byte(event.Data)),
								}
								dbEvent.Upsert(ctx, pdb.DBS().GetWriterConn(), false, nil, boil.Infer(), boil.Infer())
							}
							tx.Commit()
							temp := make([]kafka.TopicPartition, 0, len(partitions))
							for _, p := range partitions {
								temp = append(temp, p)
							}
							events = nil
							c.CommitOffsets(temp)
						}()
					}
				}
			} else if err.(kafka.Error).Code() != kafka.ErrTimedOut {
				fmt.Printf("Consumer error: %v\n", err)
			}
		}
	}()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return ErrorHandler(c, err, logger)
		},
		DisableStartupMessage: true,
	})
	eventsController := controllers.NewEventsController(settings, pdb.DBS, &logger)

	app.Use(recover.New(recover.Config{
		Next:              nil,
		EnableStackTrace:  true,
		StackTraceHandler: nil,
	}))
	app.Use(cors.New())
	app.Get("/", HealthCheck)

	v1 := app.Group("/v1/events")
	v1.Get("/", eventsController.GetEvents)

	logger.Info().Msg("Server started on port " + settings.Port)

	// Start Server
	if err := app.Listen(":" + settings.Port); err != nil {
		logger.Fatal().Err(err)
	}
}

// HealthCheck godoc
// @Summary Show the status of server.
// @Description get the status of server.
// @Tags root
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func HealthCheck(c *fiber.Ctx) error {
	res := map[string]interface{}{
		"data": "Server is up and running",
	}

	if err := c.JSON(res); err != nil {
		return err
	}

	return nil
}

// ErrorHandler custom handler to log recovered errors using our logger and return json instead of string
func ErrorHandler(c *fiber.Ctx, err error, logger zerolog.Logger) error {
	code := fiber.StatusInternalServerError // Default 500 statuscode

	if e, ok := err.(*fiber.Error); ok {
		// Override status code if fiber.Error type
		code = e.Code
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	logger.Err(err).Msg("caught a panic")

	return c.Status(code).JSON(fiber.Map{
		"error": true,
		"msg":   err.Error(),
	})
}
