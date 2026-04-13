package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Recurrence recurrenceDTO `json:"recurrence"`
}

type recurrenceDTO struct {
	Type 	taskdomain.RecurrenceType `json:"type"`
	Interval int `json:"interval,omitempty"`
	DayOfMonth int `json:"day_of_month,omitempty"`
	Dates []time.Time `json:"dates,omitempty"`
	StartDate time.Time `json:"start_date,omitempty"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Recurrence recurrenceDTO      `json:"recurrence"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Recurrence: recurrenceDTO{
   			Type:       task.Recurrence.Type,
   			Interval:   task.Recurrence.Interval,
   			DayOfMonth: task.Recurrence.DayOfMonth,
   			Dates:      task.Recurrence.Dates,
   			StartDate:  task.Recurrence.StartDate,
  		},
  		CreatedAt: task.CreatedAt,
  		UpdatedAt: task.UpdatedAt,
	}
}
