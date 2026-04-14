package task

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Recurrence:   normalized.Recurrence,
	}
	now := s.now()
	if model.Recurrence.Type == taskdomain.RecurrenceDaily && model.Recurrence.StartDate.IsZero() {
  		model.Recurrence.StartDate = toDate(now)
 	}
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Recurrence:   normalized.Recurrence,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	recurrence, err := validateRecurrence(input.Recurrence)
 	if err != nil {
  		return CreateInput{}, err
 	}
 	input.Recurrence = recurrence

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	recurrence, err := validateRecurrence(input.Recurrence)
 	if err != nil {
  		return UpdateInput{}, err
 	}
 	input.Recurrence = recurrence

 	return input, nil
}

func validateRecurrence(input taskdomain.Recurrence) (taskdomain.Recurrence, error) {
 	if input.Type == "" {
  	input.Type = taskdomain.RecurrenceNone
 	}

 	if !input.Type.Valid() {
  		return taskdomain.Recurrence{}, fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
 	}

 	input.Dates = normalizeDates(input.Dates)

 	switch input.Type {
 	case taskdomain.RecurrenceNone:
 		return taskdomain.Recurrence{Type: taskdomain.RecurrenceNone}, nil
 	case taskdomain.RecurrenceDaily:
  		if input.Interval <= 0 {
   			return taskdomain.Recurrence{}, fmt.Errorf("%w: interval must be positive", ErrInvalidInput)
  		}
  		if !input.StartDate.IsZero() {
   			input.StartDate = toDate(input.StartDate)
  		}
  		input.DayOfMonth = 0
  		input.Dates = nil
 	case taskdomain.RecurrenceMonthly:
  		if input.DayOfMonth < 1 || input.DayOfMonth > 30 {
  			return taskdomain.Recurrence{}, fmt.Errorf("%w: day_of_month must be from 1 to 30", ErrInvalidInput)
  		}
  		input.Interval = 0
  		input.Dates = nil
  		input.StartDate = time.Time{}
 	case taskdomain.RecurrenceSpecificDates:
  		if len(input.Dates) == 0 {
   			return taskdomain.Recurrence{}, fmt.Errorf("%w: at least one date is required", ErrInvalidInput)
  		}
  		input.Interval = 0
  		input.DayOfMonth = 0
  		input.StartDate = time.Time{}
 	case taskdomain.RecurrenceEvenDays, taskdomain.RecurrenceOddDays:
  		input.Interval = 0
  		input.DayOfMonth = 0
  		input.Dates = nil
  		input.StartDate = time.Time{}
	}

 	return input, nil
}

func normalizeDates(dates []time.Time) []time.Time {
 	if len(dates) == 0 {
  		return nil
 	}

 	out := make([]time.Time, 0, len(dates))
 	seen := make(map[string]struct{}, len(dates))
 	for _, d := range dates {
  		nd := toDate(d)
  		key := nd.Format(time.DateOnly)
  		if _, ok := seen[key]; ok {
   			continue
  		}
  		seen[key] = struct{}{}
  		out = append(out, nd)
 	}

 	slices.SortFunc(out, func(a, b time.Time) int {
  		return a.Compare(b)
 	})
 	return out
}

func toDate(t time.Time) time.Time {
 	t = t.UTC()
 	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
