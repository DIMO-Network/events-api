package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/DIMO-INC/events-api/internal/config"
	"github.com/DIMO-INC/events-api/internal/controllers"
	"github.com/DIMO-INC/events-api/internal/database"
	"github.com/DIMO-INC/events-api/internal/kafka"
	"github.com/DIMO-INC/events-api/internal/services"
	"github.com/Shopify/sarama"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/rs/zerolog"
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
		startEventConsumer(logger, settings, pdb)
		startWebAPI(logger, settings, pdb, ctx)
	}
}

func startEventConsumer(logger zerolog.Logger, settings *config.Settings, pdb database.DbStore) {
	clusterConfig := sarama.NewConfig()
	clusterConfig.Version = sarama.V2_6_0_0
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	cfg := &kafka.Config{
		ClusterConfig:   clusterConfig,
		BrokerAddresses: strings.Split(settings.KafkaBrokers, ","),
		Topic:           settings.EventsTopic,
		GroupID:         "event-sink",
		MaxInFlight:     int64(5),
	}
	consumer, err := kafka.NewConsumer(cfg, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not start consumer")
	}
	ingestService := services.NewIngestService(pdb.DBS, &logger)
	consumer.Start(context.Background(), ingestService.Ingest)

	logger.Info().Msg("kafka consumer started")
}

func startWebAPI(logger zerolog.Logger, settings *config.Settings, pdb database.DbStore, ctx context.Context) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return ErrorHandler(c, err, logger)
		},
		DisableStartupMessage: true,
	})

	keyRefreshInterval := time.Hour
	keyRefreshUnknownKID := true
	jwtAuth := jwtware.New(jwtware.Config{
		KeySetURL:            settings.JWTKeySetURL,
		KeyRefreshInterval:   &keyRefreshInterval,
		KeyRefreshUnknownKID: &keyRefreshUnknownKID,
	})
	eventsController := controllers.NewEventsController(settings, pdb.DBS, &logger)

	app.Use(recover.New(recover.Config{
		Next:              nil,
		EnableStackTrace:  true,
		StackTraceHandler: nil,
	}))
	app.Use(cors.New())
	app.Get("/", HealthCheck)

	v1 := app.Group("/v1/events", jwtAuth)
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
