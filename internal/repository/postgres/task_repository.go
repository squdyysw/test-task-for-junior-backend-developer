package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, recurrence_type, recurrence_config, created_at updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, status, recurrence_type, recurrence_config, created_at, updated_at
	`

	recurrenceConfig, err := marshalRecurrenceConfig(task.Recurrence)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.recurrence_type,
		task.recurrence_config,
		task.CreatedAt,
		task.UpdatedAt
	)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, recurrence_type, recurrence_config, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			recurrence_type = $4,
			recurrence_config = $5,
			updated_at = $6
		WHERE id = $7
		RETURNING id, title, description, status, recurrence_type, recurrence_config, created_at, updated_at
	`

	recurrenceConfig, err := marshalRecurrenceConfig(task.Recurrence)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.recurrence_type,
		task.recurrence_config,
		task.CreatedAt,
		task.UpdatedAt,
		task.ID
	)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, recurrence_type, recurrence_config, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task           taskdomain.Task
		status 		   string
		recurrenceType string
		recurrenceRaw  []byte
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&recurrenceType,
		&recurrenceRaw,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	task.Recurrence = taskdomain.Recurrence{
  		Type: taskdomain.RecurrenceType(recurrenceType),
 	}
 	if err := unmarshalRecurrenceConfig(recurrenceRaw, &task.Recurrence); err != nil {
  		return nil, err
 	}

 	return &task, nil
}

type recurrenceConfig struct {
 	Interval   int         `json:"interval,omitempty"`
 	DayOfMonth int         `json:"day_of_month, omitempty"`
  	Dates      []timeAlias `json:"dates, omitempty"`
  	StartDate  *timeAlias  `json:"start_date, omitempty"`
}

type timeAlias string

func marshalRecurrenceConfig(recurrence taskdomain.Recurrence) ([]byte, error) {
 	cfg := recurrenceConfig{
  		Interval: recurrence.Interval,
  		DayOfMonth: recurrence.DayOfMonth,
  		Dates: make([]timeAlias, 0, len(recurrence.Dates)),
 	}
 	for _, date := range recurrence.Dates {
  		cfg.Dates = append(cfg.Dates, timeAlias(date.Format("2006-01-02")))
 	}
 	if !recurrence.StartDate.IsZero() {
  		v := timeAlias(recurrence.StartDate.Format("2006-01-02"))
  		cfg.StartDate = &v
 	}

 	raw, err := json.Marshal(cfg)
 	if err != nil {
  		return nil, fmt.Errorf("marshal recurrence config: %w", err)
 	}

 	return raw, nil
}

func unmarshalRecurrenceConfig(raw []byte, recurrence *taskdomain.Recurrence) error {
 	if len(raw) == 0 {
 		return nil
 	}

 	var cfg recurrenceConfig
 	if err := json.Unmarshal(raw, &cfg); err != nil {
  		return fmt.Errorf("unmarshal recurrence config: %w", err)
 	}

 	recurrence.Interval = cfg.Interval
 	recurrence.DayOfMonth = cfg.DayOfMonth
 	recurrence.Dates = make([]time.Time, 0, len(cfg.Dates))
 	for _, date := range cfg.Dates {
  		parsed, err := time.Parse("2006-01-02", string(date))
  		if err != nil {
   		return fmt.Errorf("parse recurrence date: %w", err)
  		}
  	recurrence.Dates = append(recurrence.Dates, parsed.UTC())
 	}
 	if cfg.StartDate != nil {
  		parsed, err := time.Parse("2006-01-02", string(*cfg.StartDate))
  		if err != nil {
   			return fmt.Errorf("parse recurrence start_date: %w", err)
  		}
  		recurrence.StartDate = parsed.UTC()
 	}

 	return nil
}
