package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Recurrence Recurrence `json:"recurrnece"` 
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Recurrence struct {
	Type RecurrenceType `json:"type"`
	Interval int 		`json:"interval,omitempty"`
	DayOfMonth int 		`json:"day_of_month,omitempty"`
	Dates []time.Time 	`json:"dates,omitempty"`
	StartDate time.Time `json:"start_date,omitempty"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (r RecurrenceType) Valid() bool {
	switch r {
	case RecurrenceNone, RecurrenceDaily, RecurrenceMonthly, RecurrenceSpecificDates, RecurrenceEvenDays, RecurrenceOddDays:
		return true
	default:
		return false
	}
}
