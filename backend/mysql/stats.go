package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cschleiden/go-workflows/backend"
	"github.com/cschleiden/go-workflows/core"
	"github.com/cschleiden/go-workflows/workflow"
)

func (mb *mysqlBackend) GetStats(ctx context.Context) (*backend.Stats, error) {
	s := &backend.Stats{}

	tx, err := mb.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get active instances
	row := tx.QueryRowContext(
		ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s i WHERE i.completed_at IS NULL", mb.tables.Instances),
	)
	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("failed to query active instances: %w", err)
	}

	var activeInstances int64
	if err := row.Scan(&activeInstances); err != nil {
		return nil, fmt.Errorf("failed to scan active instances: %w", err)
	}

	s.ActiveWorkflowInstances = activeInstances

	// Get workflow instances ready to be picked up
	now := time.Now()
	workflowRows, err := tx.QueryContext(
		ctx,
		fmt.Sprintf(`SELECT i.queue, COUNT(*)
			FROM %s i
			INNER JOIN %s pe ON i.instance_id = pe.instance_id
			WHERE
				i.state = ? AND i.completed_at IS NULL
				AND (pe.visible_at IS NULL OR pe.visible_at <= ?)
				AND (i.locked_until IS NULL OR i.locked_until < ?)
			GROUP BY i.queue`, mb.tables.Instances, mb.tables.PendingEvents),
		core.WorkflowInstanceStateActive,
		now, // event.visible_at
		now, // locked_until
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query active instances: %w", err)
	}

	s.PendingWorkflowTasks = make(map[core.Queue]int64)

	for workflowRows.Next() {
		var queue string
		var pendingInstances int64
		if err := workflowRows.Scan(&queue, &pendingInstances); err != nil {
			return nil, fmt.Errorf("failed to scan active instances: %w", err)
		}

		s.PendingWorkflowTasks[workflow.Queue(queue)] = pendingInstances
	}

	if err := workflowRows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read active instances: %w", err)
	}

	// Get pending activities
	activityRows, err := tx.QueryContext(
		ctx,
		fmt.Sprintf("SELECT queue, COUNT(*) FROM %s GROUP BY queue", mb.tables.Activities))
	if err != nil {
		return nil, fmt.Errorf("failed to query active activities: %w", err)
	}

	s.PendingActivityTasks = make(map[core.Queue]int64)

	for activityRows.Next() {
		var queue string
		var pendingActivities int64
		if err := activityRows.Scan(&queue, &pendingActivities); err != nil {
			return nil, fmt.Errorf("failed to scan active activities: %w", err)
		}

		s.PendingActivityTasks[workflow.Queue(queue)] = pendingActivities
	}

	if err := activityRows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read active activities: %w", err)
	}

	return s, nil
}
