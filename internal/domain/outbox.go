package domain

import (
	"context"
	"database/sql"
	"time"
)

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "PENDING"
	OutboxStatusProcessed OutboxStatus = "PROCESSED"
)

type OutboxEvent struct {
	ID            string
	AggregateType string
	AggregateID   string
	Topic         string
	Payload       []byte
	Status        OutboxStatus
	CreatedAt     time.Time
}

type OutboxRepository interface {
	Save(ctx context.Context, tx *sql.Tx, event OutboxEvent) error
	FindPending(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkAsProcessed(ctx context.Context, ids []string) error
}
