package controllers

import (
	"time"

	"github.com/DIMO-INC/events-api/internal/config"
	"github.com/DIMO-INC/events-api/internal/database"
	"github.com/DIMO-INC/events-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type EventsController struct {
	Settings *config.Settings
	DBS      func() *database.DBReaderWriter
	log      *zerolog.Logger
}

func NewEventsController(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger) EventsController {
	return EventsController{
		Settings: settings,
		DBS:      dbs,
		log:      logger,
	}
}

// Just want to get the fields into camelCase. Is there a better way?
type EventResponseEntry struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	SubType   string      `json:"subType"`
	UserID    string      `json:"userId"`
	DeviceID  null.String `json:"deviceId"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

func (e *EventsController) GetEvents(c *fiber.Ctx) error {
	userID := getUserID(c)
	mods := []qm.QueryMod{
		models.EventWhere.UserID.EQ(userID),
		qm.OrderBy(models.EventColumns.Timestamp + " DESC"),
	}

	deviceID := c.Query("device_id")
	if deviceID != "" {
		mods = append(mods, models.EventWhere.DeviceID.EQ(null.StringFrom(deviceID)))
	}

	eventType := c.Query("type")
	if eventType != "" {
		mods = append(mods, models.EventWhere.Type.EQ(eventType))
		eventSubType := c.Query("sub_type")
		if eventSubType != "" {
			mods = append(mods, models.EventWhere.SubType.EQ(eventSubType))
		}
	}

	events, err := models.Events(mods...).All(c.Context(), e.DBS().Reader)
	if err != nil {
		return c.JSON(fiber.Map{"Uhoh": err.Error()})
	}

	if events == nil {
		return c.JSON(models.EventSlice{})
	}

	respEvents := make([]EventResponseEntry, len(events))
	for i, event := range events {
		respEvents[i] = EventResponseEntry{
			ID:        event.ID,
			Type:      event.Type,
			SubType:   event.SubType,
			UserID:    event.UserID,
			Timestamp: event.Timestamp,
			Data:      event.Data,
		}
	}

	return c.JSON(respEvents)
}
