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

		dbEvent := models.Event{
			ID:      event.ID,
			Type:    event.Type,
			Source:  event.Source,
			Subject: event.Subject,
			Time:    event.Time,
			Data:    null.JSONFrom(event.Data),
		}
		err = dbEvent.Upsert(context.Background(), tx, false, []string{"id"}, boil.Infer(), boil.Infer())
		if err != nil {
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
