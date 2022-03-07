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

// EventResponseEntry represents a single user event.
type EventResponseEntry struct {
	// ID is the event ID, a KSUID.
	ID string `json:"id"`
	// Type is the type of the event, either "User" or "Device".
	Type string `json:"type" example:"Device"`
	// SubType is a more specific classification of the event.
	SubType string `json:"subType" example:"Created"`
	// UserID is the DIMO user id coming from the authentication service.
	UserID string `json:"userId"`
	// DeviceID is the DIMO device ID when the event relates to a device.
	DeviceID  null.String `json:"deviceId" swaggertype:"string"`
	Timestamp time.Time   `json:"timestamp"`
	// Data is an event-specific object containing more information about the event.
	Data interface{} `json:"data"`
}

// GetEvents godoc
// @Description Lists the user's events in reverse chronological order
// @Tags events
// @Produce json
// @Param device_id query string false "DIMO device ID"
// @Param type query string false "Event type, usually User or Device"
// @Param sub_type query string false "Further refinement of the event type"
// @Success 200 {object} []EventResponseEntry
// @Router /v1/events [get]
func (e *EventsController) GetEvents(c *fiber.Ctx) error {
	userID := getUserID(c)
	mods := []qm.QueryMod{
		models.EventWhere.UserID.EQ(userID),
		qm.OrderBy(models.EventColumns.Timestamp + " DESC"),
		qm.Limit(100),
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
