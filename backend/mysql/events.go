package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/cschleiden/go-workflows/backend/history"
	"github.com/cschleiden/go-workflows/core"
)

func (mb *mysqlBackend) insertPendingEvents(ctx context.Context, tx *sql.Tx, instance *core.WorkflowInstance, newEvents []*history.Event) error {
	return mb.insertEvents(ctx, tx, mb.tables.PendingEvents, instance, newEvents)
}

func (mb *mysqlBackend) insertHistoryEvents(ctx context.Context, tx *sql.Tx, instance *core.WorkflowInstance, historyEvents []*history.Event) error {
	return mb.insertEvents(ctx, tx, mb.tables.History, instance, historyEvents)
}

func (mb *mysqlBackend) insertEvents(ctx context.Context, tx *sql.Tx, tableName string, instance *core.WorkflowInstance, events []*history.Event) error {
	const batchSize = 20
	for batchStart := 0; batchStart < len(events); batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > len(events) {
			batchEnd = len(events)
		}
		batchEvents := events[batchStart:batchEnd]

		aquery := fmt.Sprintf("INSERT IGNORE INTO %s (event_id, instance_id, execution_id, data) VALUES (?, ?, ?, ?)", mb.tables.Attributes) + strings.Repeat(", (?, ?, ?, ?)", len(batchEvents)-1)
		aargs := make([]interface{}, 0, len(batchEvents)*4)

		query := fmt.Sprintf("INSERT INTO %s (event_id, sequence_id, instance_id, execution_id, event_type, timestamp, schedule_event_id, visible_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", tableName) +
			strings.Repeat(", (?, ?, ?, ?, ?, ?, ?, ?)", len(batchEvents)-1)

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

func (mb *mysqlBackend) removeFutureEvent(ctx context.Context, tx *sql.Tx, instance *core.WorkflowInstance, scheduleEventID int64) error {
	_, err := tx.ExecContext(
		ctx,
		fmt.Sprintf("DELETE pe, a FROM %s pe INNER JOIN %s a ON pe.event_id = a.event_id WHERE pe.instance_id = ? AND pe.execution_id = ? AND pe.schedule_event_id = ? AND pe.visible_at IS NOT NULL", mb.tables.PendingEvents, mb.tables.Attributes),
		instance.InstanceID,
		instance.ExecutionID,
		scheduleEventID,
	)

	return err
}
