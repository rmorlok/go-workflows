package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/cschleiden/go-workflows/backend/history"
	"github.com/cschleiden/go-workflows/backend/test"
	"github.com/cschleiden/go-workflows/core"
)

var _ test.TestBackend = (*sqliteBackend)(nil)

func (sb *sqliteBackend) GetFutureEvents(ctx context.Context) ([]*history.Event, error) {
	tx, err := sb.db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// There is no index on `visible_at`, but this is okay for test only usage.
	futureEvents, err := tx.QueryContext(
		ctx,
		fmt.Sprintf("SELECT pe.id, pe.sequence_id, pe.instance_id, pe.execution_id, pe.event_type, pe.timestamp, pe.schedule_event_id, pe.visible_at, a.data FROM %s pe JOIN %s a ON a.id = pe.id AND a.instance_id = pe.instance_id AND a.execution_id = pe.execution_id WHERE pe.visible_at IS NOT NULL", sb.tables.PendingEvents, sb.tables.Attributes),
	)
	if err != nil {
		return nil, fmt.Errorf("getting history: %w", err)
	}
	if futureEvents.Err() != nil {
		return nil, futureEvents.Err()
	}

	defer futureEvents.Close()

	f := make([]*history.Event, 0)

	for futureEvents.Next() {
		var instanceID, executionID string
		var attributes []byte

		fe := &history.Event{}

		if err := futureEvents.Scan(
			&fe.ID,
			&fe.SequenceID,
			&instanceID,
			&executionID,
			&fe.Type,
			&fe.Timestamp,
			&fe.ScheduleEventID,
			&fe.VisibleAt,
			&attributes,
		); err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}

		a, err := history.DeserializeAttributes(fe.Type, attributes)
		if err != nil {
			return nil, fmt.Errorf("deserializing attributes: %w", err)
		}

		fe.Attributes = a

		f = append(f, fe)
	}

	return f, nil
}

func (sb *sqliteBackend) getPendingEvents(ctx context.Context, tx *sql.Tx, instance *core.WorkflowInstance) ([]*history.Event, error) {
	now := time.Now()
	events, err := tx.QueryContext(
		ctx,
		fmt.Sprintf("SELECT pe.*, a.data FROM %s pe INNER JOIN %s a ON a.id = pe.id AND a.instance_id = pe.instance_id AND a.execution_id = pe.execution_id WHERE pe.instance_id = ? AND pe.execution_id = ? AND (pe.visible_at IS NULL OR pe.visible_at <= ?)", sb.tables.PendingEvents, sb.tables.Attributes),
		instance.InstanceID,
		instance.ExecutionID,
		now,
	)

	if err != nil {
		return nil, fmt.Errorf("getting new events: %w", err)
	}

	defer events.Close()

	pendingEvents := make([]*history.Event, 0)

	for events.Next() {
		pendingEvent, err := scanEvent(events)
		if err != nil {
			return nil, fmt.Errorf("reading event: %w", err)
		}

		pendingEvents = append(pendingEvents, pendingEvent)
	}

	if events.Err() != nil {
		return nil, events.Err()
	}

	return pendingEvents, nil
}

func (sb *sqliteBackend) getHistory(ctx context.Context, tx *sql.Tx, instance *core.WorkflowInstance, lastSequenceID *int64) ([]*history.Event, error) {
	var historyEvents *sql.Rows
	var err error
	if lastSequenceID != nil {
		historyEvents, err = tx.QueryContext(
			ctx, fmt.Sprintf("SELECT h.*, a.data FROM %s h INNER JOIN %s a ON a.id = h.id AND a.instance_id = h.instance_id AND a.execution_id = h.execution_id WHERE h.instance_id = ? AND h.execution_id = ? AND h.sequence_id > ?", sb.tables.History, sb.tables.Attributes), instance.InstanceID, instance.ExecutionID, *lastSequenceID)
	} else {
		historyEvents, err = tx.QueryContext(
			ctx, fmt.Sprintf("SELECT h.*, a.data FROM %s h INNER JOIN %s a ON a.id = h.id AND a.instance_id = h.instance_id AND a.execution_id = h.execution_id WHERE h.instance_id = ? AND h.execution_id = ?", sb.tables.History, sb.tables.Attributes), instance.InstanceID, instance.ExecutionID)
	}
	if err != nil {
		return nil, fmt.Errorf("getting history: %w", err)
	}

	defer historyEvents.Close()

	events := make([]*history.Event, 0)

	for historyEvents.Next() {
		historyEvent, err := scanEvent(historyEvents)
		if err != nil {
			return nil, fmt.Errorf("reading event: %w", err)
		}

		events = append(events, historyEvent)
	}

	if historyEvents.Err() != nil {
		return nil, historyEvents.Err()
	}

	return events, nil
}

type Scanner interface {
	Scan(dest ...interface{}) error
}

func scanEvent(row Scanner) (*history.Event, error) {
	var instanceID, executionID string
	var attributes []byte

	historyEvent := &history.Event{}

	if err := row.Scan(
		&historyEvent.ID,
		&historyEvent.SequenceID,
		&instanceID,
		&executionID,
		&historyEvent.Type,
		&historyEvent.Timestamp,
		&historyEvent.ScheduleEventID,
		&historyEvent.VisibleAt,
		&attributes,
	); err != nil {
		return historyEvent, fmt.Errorf("scanning event: %w", err)
	}

	a, err := history.DeserializeAttributes(historyEvent.Type, attributes)
	if err != nil {
		return historyEvent, fmt.Errorf("deserializing attributes: %w", err)
	}

	historyEvent.Attributes = a

	return historyEvent, nil
}

func (sb *sqliteBackend) insertPendingEvents(ctx context.Context, tx *sql.Tx, instance *core.WorkflowInstance, newEvents []*history.Event) error {
	return sb.insertEvents(ctx, tx, sb.tables.PendingEvents, instance, newEvents)
}

func (sb *sqliteBackend) insertHistoryEvents(ctx context.Context, tx *sql.Tx, instance *core.WorkflowInstance, historyEvents []*history.Event) error {
	return sb.insertEvents(ctx, tx, sb.tables.History, instance, historyEvents)
}

func (sb *sqliteBackend) insertEvents(ctx context.Context, tx *sql.Tx, tableName string, instance *core.WorkflowInstance, events []*history.Event) error {
	const batchSize = 20
	for batchStart := 0; batchStart < len(events); batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > len(events) {
			batchEnd = len(events)
		}
		batchEvents := events[batchStart:batchEnd]

		// INSERT OR IGNORE since the attributes might already exist due to an event being moved from pending to history.
		aquery := fmt.Sprintf("INSERT OR IGNORE INTO %s (id, instance_id, execution_id, data) VALUES (?, ?, ?, ?)", sb.tables.Attributes) + strings.Repeat(", (?, ?, ?, ?)", len(batchEvents)-1)
		aargs := make([]interface{}, 0, len(batchEvents)*4)

		query := fmt.Sprintf("INSERT INTO %s (id, sequence_id, instance_id, execution_id, event_type, timestamp, schedule_event_id, visible_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", tableName) +
			strings.Repeat(", (?, ?, ?, ?, ?, ?, ?, ?)", len(batchEvents)-1)

		args := make([]interface{}, 0, len(batchEvents)*8)

		for _, newEvent := range batchEvents {
			a, err := history.SerializeAttributes(newEvent.Attributes)
			if err != nil {
				return err
			}

			aargs = append(aargs, newEvent.ID, instance.InstanceID, instance.ExecutionID, a)

			args = append(
				args, newEvent.ID, newEvent.SequenceID, instance.InstanceID, instance.ExecutionID, newEvent.Type, newEvent.Timestamp, newEvent.ScheduleEventID, newEvent.VisibleAt)
		}

		if _, err := tx.ExecContext(
			ctx,
			aquery,
			aargs...,
		); err != nil {
			return fmt.Errorf("inserting attributes: %w", err)
		}

		if _, err := tx.ExecContext(
			ctx,
			query,
			args...,
		); err != nil {
			return fmt.Errorf("inserting events: %w", err)
		}
	}

	return nil
}

func (sb *sqliteBackend) removeFutureEvent(ctx context.Context, tx *sql.Tx, instance *core.WorkflowInstance, scheduleEventID int64) error {
	row, err := tx.QueryContext(
		ctx,
		fmt.Sprintf("DELETE FROM %s WHERE instance_id = ? AND execution_id = ? AND schedule_event_id = ? AND visible_at IS NOT NULL RETURNING id", sb.tables.PendingEvents),
		instance.InstanceID,
		instance.ExecutionID,
		scheduleEventID,
	)
	if err != nil {
		return fmt.Errorf("removing future event: %w", err)
	}

	ids := make([]interface{}, 0)

	defer row.Close()
	for row.Next() {
		var id string
		if err := row.Scan(&id); err != nil {
			return fmt.Errorf("scanning id: %w", err)
		}

		ids = append(ids, id)
	}

	if row.Err() != nil {
		return row.Err()
	}

	// Delete attributes
	if len(ids) > 0 {
		query := fmt.Sprintf("DELETE FROM %s WHERE id IN (?)", sb.tables.Attributes) + strings.Repeat(", (?)", len(ids)-1)

		if _, err := tx.ExecContext(
			ctx,
			query,
			ids...,
		); err != nil {
			return fmt.Errorf("removing attributes: %w", err)
		}
	}

	return err
}
