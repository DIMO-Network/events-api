package controllers

import (
	"time"

	"github.com/DIMO-INC/events-api/internal/config"
	"github.com/DIMO-INC/events-api/internal/database"
	"github.com/DIMO-INC/events-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type EventsController struct {
	Settings        *config.Settings
	DBS             func() *database.DBReaderWriter
	log             *zerolog.Logger
	allowedLateness time.Duration
}

func NewEventsController(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger) EventsController {
	return EventsController{
		Settings: settings,
		DBS:      dbs,
		log:      logger,
	}
}

func (e *EventsController) GetEvents(c *fiber.Ctx) error {
	events, err := models.Events().All(c.Context(), e.DBS().Reader)
	if err != nil {
		return c.JSON(fiber.Map{"Uhoh": err.Error()})
	}
	return c.JSON(events)
}
