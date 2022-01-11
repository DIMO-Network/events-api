package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/DIMO-INC/events-api/internal/database"
	"github.com/DIMO-INC/events-api/models"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type EventMessage struct {
	Type    string          `json:"type"`
	Source  string          `json:"source"`
	Subject string          `json:"subject"`
	ID      string          `json:"id"`
	Time    time.Time       `json:"time"`
	Data    json.RawMessage `json:"data"`
}

type partialUser struct {
	UserID string `json:"userId"`
}

type IngestService struct {
	db            func() *database.DBReaderWriter
	logger        *zerolog.Logger
	messageBuffer []*message.Message
}

func NewIngestService(db func() *database.DBReaderWriter, logger *zerolog.Logger) *IngestService {
	return &IngestService{
		db:     db,
		logger: logger,
	}
}

func (i *IngestService) Ingest(messages <-chan *message.Message) {
	ticker := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-ticker.C:
			if err := i.insertEvents(); err != nil {
				i.logger.Err(err).Msg("Failed to upsert buffered events, continuing to buffer")
			}
		case message := <-messages:
			i.messageBuffer = append(i.messageBuffer, message)
		}
	}
}

type eventType struct {
	Type    string
	SubType string
}

var eventTypeMap = map[string]eventType{
	"com.dimo.zone.user.create":               {"User", "Created"},
	"com.dimo.zone.device.create":             {"Device", "Created"},
	"com.dimo.zone.device.delete":             {"Device", "Deleted"},
	"com.dimo.zone.device.integration.create": {"Device", "IntegrationCreated"},
	"com.dimo.zone.device.integration.delete": {"Device", "IntegrationDeleted"},
	"com.dimo.zone.device.odometer.update":    {"Device", "OdometerUpdated"},
	"com.dimo.zone.user.token.issue":          {"User", "TokensIssued"},
	"com.dimo.zone.user.referral.complete":    {"User", "ReferralCompleted"},
}

func (i *IngestService) insertEvents() error {
	tx, err := i.db().Writer.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	for _, message := range i.messageBuffer {
		var event EventMessage
		err = json.Unmarshal(message.Payload, &event)
		if err != nil {
			i.logger.Err(err).Msg("Failed to unmarshal event message, planning to skip")
			continue
		}

		eventType, ok := eventTypeMap[event.Type]
		if !ok {
			i.logger.Error().Msgf("Event %s has unrecognized event type %s, skipping", event.ID, event.Type)
			continue
		}

		var user partialUser
		err = json.Unmarshal(event.Data, &user)
		if err != nil {
			i.logger.Err(err).Msgf("Event %s had an unparseable data field, skipping", event.ID)
			continue
		}

		dbEvent := models.Event{
			ID:        event.ID,
			Type:      eventType.Type,
			SubType:   eventType.SubType,
			UserID:    user.UserID,
			Timestamp: event.Time,
			Data:      null.JSONFrom(event.Data),
		}

		err = dbEvent.Upsert(context.Background(), tx, false, []string{"id"}, boil.Infer(), boil.Infer())
		if err != nil {
			i.logger.Err(err).Msg("Failed to insert event")
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	for _, message := range i.messageBuffer {
		message.Ack()
	}

	// Keep the capacity around. Maybe revisit this.
	i.messageBuffer = i.messageBuffer[:0]
	return nil
}
