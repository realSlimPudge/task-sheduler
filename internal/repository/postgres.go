package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"task-scheduler/internal/apperrors"
	"task-scheduler/internal/model"
	"time"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{
		db: db,
	}
}

func (r *PostgresRepo) GetTasks(ctx context.Context) ([]model.Task, error) {
	const query = `
	SELECT id, type, payload, status, result, COALESCE(error, ''), attempts, created_at, updated_at, next_attempt_at
	FROM tasks;
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task

	for rows.Next() {
		var task model.Task
		var result []byte

		err := rows.Scan(
			&task.ID,
			&task.Type,
			&task.Payload,
			&task.Status,
			&result,
			&task.Error,
			&task.Attempts,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.NextAttemptAt,
		)
		if err != nil {
			return nil, err
		}
		task.Result = json.RawMessage(result)
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *PostgresRepo) GetTaskByID(ctx context.Context, id int) (*model.Task, error) {
	const query = `
	SELECT id, type, payload, status, result, COALESCE(error, ''), attempts, created_at, updated_at, next_attempt_at
 	FROM tasks
	WHERE id = $1;
	`
	var task model.Task
	var result []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.Type,
		&task.Payload,
		&task.Status,
		&result,
		&task.Error,
		&task.Attempts,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.NextAttemptAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	task.Result = json.RawMessage(result)
	return &task, nil
}

func (r *PostgresRepo) GetStats(ctx context.Context) (*model.TaskStats, error) {
	const query = `
	SELECT
		COUNT(*) FILTER (WHERE status = 'pending') AS pending,
		COUNT(*) FILTER (WHERE status = 'running') AS running,
		COUNT(*) FILTER (WHERE status = 'completed') AS completed,
		COUNT(*) FILTER (WHERE status = 'failed') AS failed
	FROM tasks;
	`
	var stats model.TaskStats

	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.Pending,
		&stats.Running,
		&stats.Completed,
		&stats.Failed,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *PostgresRepo) CreateTask(ctx context.Context, taskType model.TaskType, payload json.RawMessage) (*model.Task, error) {
	const query = `
	INSERT INTO tasks (type, payload) VALUES($1, $2)
	RETURNING id, type, payload, status, result, COALESCE(error, ''), attempts, created_at, updated_at, next_attempt_at;
	`
	created := &model.Task{}
	var result []byte

	err := r.db.QueryRowContext(ctx, query, taskType, payload).Scan(
		&created.ID,
		&created.Type,
		&created.Payload,
		&created.Status,
		&result,
		&created.Error,
		&created.Attempts,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.NextAttemptAt,
	)
	if err != nil {
		return nil, err
	}
	created.Result = json.RawMessage(result)
	return created, nil
}

func (r *PostgresRepo) GetPendingTasks(ctx context.Context, limit int) ([]model.Task, error) {
	const query = `
	WITH selected AS (
		SELECT id
		FROM tasks
		WHERE status = 'pending' AND next_attempt_at <= NOW()
		ORDER BY next_attempt_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	)
	UPDATE tasks t
	SET status = 'running', updated_at = NOW()
	FROM selected s
	WHERE t.id = s.id
	RETURNING t.id, t.type, t.payload, t.status, t.attempts, t.created_at, t.updated_at, t.next_attempt_at;
	`

	var tasks []model.Task
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task model.Task

		err := rows.Scan(
			&task.ID,
			&task.Type,
			&task.Payload,
			&task.Status,
			&task.Attempts,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.NextAttemptAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *PostgresRepo) MarkCompleted(ctx context.Context, id int, result json.RawMessage) error {
	const query = `
	UPDATE tasks
	SET
		status = 'completed',
	 	result = $2,
		updated_at = NOW()
	WHERE id = $1;
	`
	sqlResult, err := r.db.ExecContext(ctx, query, id, result)
	if err != nil {
		return err
	}

	aff, err := sqlResult.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *PostgresRepo) MarkFailed(ctx context.Context, id int, taskErr string) error {
	const query = `
	UPDATE tasks
	SET
		status = 'failed',
		error = $2,
		attempts = attempts + 1,
		updated_at = NOW()
	WHERE id = $1;
	`

	sqlResult, err := r.db.ExecContext(ctx, query, id, taskErr)
	if err != nil {
		return err
	}

	aff, err := sqlResult.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *PostgresRepo) MarkRetry(ctx context.Context, id int, taskErr string, nextAttemptAt time.Time) error {
	const query = `
	UPDATE tasks
	SET
		status = 'pending',
		error = $2,
		attempts = attempts + 1,
		next_attempt_at = $3,
		updated_at = NOW()
	WHERE id = $1;
	`
	sqlResult, err := r.db.ExecContext(ctx, query, id, taskErr, nextAttemptAt)
	if err != nil {
		return err
	}

	aff, err := sqlResult.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *PostgresRepo) RestoreZombie() (int64, error) {
	const query = `
	UPDATE tasks
	SET status = 'pending'
	WHERE status = 'running';
	`

	sqlResult, err := r.db.Exec(query)
	if err != nil {
		return 0, err
	}

	aff, err := sqlResult.RowsAffected()
	if err != nil {
		return 0, err
	}
	return aff, nil
}
