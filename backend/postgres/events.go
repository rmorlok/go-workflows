package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/cschleiden/go-workflows/backend/history"
	"github.com/cschleiden/go-workflows/core"
)

func (pb *postgresBackend) insertPendingEvents(ctx context.Context, tx *sql.Tx, instance *core.WorkflowInstance, newEvents []*history.Event) error {
	return pb.insertEvents(ctx, tx, pb.tables.PendingEvents, instance, newEvents)
}

func (pb *postgresBackend) insertHistoryEvents(ctx context.Context, tx *sql.Tx, instance *core.WorkflowInstance, historyEvents []*history.Event) error {
	return pb.insertEvents(ctx, tx, pb.tables.History, instance, historyEvents)
}

func (pb *postgresBackend) insertEvents(ctx context.Context, tx *sql.Tx, tableName string, instance *core.WorkflowInstance, events []*history.Event) error {
	const batchSize = 20
	for batchStart := 0; batchStart < len(events); batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > len(events) {
			batchEnd = len(events)
		}
		batchEvents := events[batchStart:batchEnd]

		pgPlaceHolder := func(start int, count int, size int) string {
			if count == 0 {
				return ""
			}
			holders := make([]string, count)
			for i := 0; i < count; i++ {
				holders[i] = fmt.Sprintf("($%d", start+i*size)
				for j := 1; j < size; j++ {
					holders[i] += fmt.Sprintf(", $%d", start+i*size+j)
				}
				holders[i] += ")"
			}
			return "," + strings.Join(holders, ", ")
		}

		aquery := fmt.Sprintf("INSERT INTO %s (event_id, instance_id, execution_id, data) VALUES ($1, $2, $3, $4)", pb.tables.Attributes) +
			pgPlaceHolder(5, len(batchEvents)-1, 4) +
			" ON CONFLICT DO NOTHING"
		aargs := make([]interface{}, 0, len(batchEvents)*4)

		query := fmt.Sprintf("INSERT INTO %s (event_id, sequence_id, instance_id, execution_id, event_type, timestamp, schedule_event_id, visible_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", tableName) +
			pgPlaceHolder(9, len(batchEvents)-1, 8)

		args := make([]interface{}, 0, len(batchEvents)*8)

		for _, newEvent := range batchEvents {
			a, err := history.SerializeAttributes(newEvent.Attributes)
			if err != nil {
				return err
			}

			aargs = append(aargs, newEvent.ID, instance.InstanceID, instance.ExecutionID, a)

			args = append(
				args,
				newEvent.ID, newEvent.SequenceID, instance.InstanceID, instance.ExecutionID, newEvent.Type, newEvent.Timestamp, newEvent.ScheduleEventID, newEvent.VisibleAt)
		}

		if _, err := tx.ExecContext(
			ctx,
			aquery,
			aargs...,
		); err != nil {
			return fmt.Errorf("inserting attributes: %w", err)
		}

		_, err := tx.ExecContext(
			ctx,
			query,
			args...,
		)
		if err != nil {
			return fmt.Errorf("inserting events: %w", err)
		}
	}

	return nil
}

func (pb *postgresBackend) removeFutureEvent(ctx context.Context, tx *sql.Tx, instance *core.WorkflowInstance, scheduleEventID int64) error {
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		DELETE FROM %s a
		USING %s pe
		WHERE a.event_id = pe.event_id
		  AND pe.instance_id = $1
		  AND pe.execution_id = $2
		  AND pe.schedule_event_id = $3
		  AND pe.visible_at IS NOT NULL`, pb.tables.Attributes, pb.tables.PendingEvents),
		instance.InstanceID, instance.ExecutionID, scheduleEventID); err != nil {
		return fmt.Errorf("removing attributes for future events: %w", err)
	}

	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		DELETE FROM %s
		WHERE instance_id = $1
		  AND execution_id = $2
		  AND schedule_event_id = $3
		  AND visible_at IS NOT NULL`, pb.tables.PendingEvents),
		instance.InstanceID, instance.ExecutionID, scheduleEventID); err != nil {
		return fmt.Errorf("removing pending future events: %w", err)
	}

	return nil
}
