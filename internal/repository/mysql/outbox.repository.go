package mysql

import (
	"context"
	"database/sql"
	"strings"

	"github.com/assidik12/catalyst/internal/domain"
)

type outboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) domain.OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) Save(ctx context.Context, tx *sql.Tx, event domain.OutboxEvent) error {
	q := `INSERT INTO outbox_events (id, aggregate_type, aggregate_id, topic, payload, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	_, err := tx.ExecContext(ctx, q,
		event.ID,
		event.AggregateType,
		event.AggregateID,
		event.Topic,
		event.Payload,
		event.Status,
		event.CreatedAt,
	)
	return err
}

func (r *outboxRepository) FindPending(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	q := `SELECT id, aggregate_type, aggregate_id, topic, payload, status, created_at 
		FROM outbox_events WHERE status = 'PENDING' ORDER BY created_at ASC LIMIT ?`
	
	rows, err := r.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.OutboxEvent
	for rows.Next() {
		var e domain.OutboxEvent
		if err := rows.Scan(&e.ID, &e.AggregateType, &e.AggregateID, &e.Topic, &e.Payload, &e.Status, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *outboxRepository) MarkAsProcessed(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	q := `UPDATE outbox_events SET status = 'PROCESSED' WHERE id IN (?` + strings.Repeat(",?", len(ids)-1) + `)`
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	_, err := r.db.ExecContext(ctx, q, args...)
	return err
}
