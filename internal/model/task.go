package model

import (
	"encoding/json"
	"fmt"
	"task-scheduler/internal/apperrors"
	"time"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type TaskType string

const (
	TypeEmail  TaskType = "email"
	TypeImage  TaskType = "image"
	TypeReport TaskType = "report"
)

func (t TaskType) IsValid() error {
	switch t {
	case TypeEmail, TypeImage, TypeReport:
		return nil
	default:
		return fmt.Errorf("%w: %w", apperrors.ErrValidation, apperrors.ErrInvalidTaskType)
	}
}

type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type ImagePayload struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type ReportPayload struct {
	ReportID int    `json:"report_id"`
	Format   string `json:"format"`
}

func ValidatePayload(taskType TaskType, payload json.RawMessage) error {
	switch taskType {
	case TypeEmail:
		var p EmailPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		if p.To == "" || p.Subject == "" || p.Body == "" {
			return fmt.Errorf("%w: %w", apperrors.ErrValidation, apperrors.ErrInvalidPayload)
		}
	case TypeImage:
		var p ImagePayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		if p.URL == "" || p.Width <= 0 || p.Height <= 0 {
			return fmt.Errorf("%w: %w", apperrors.ErrValidation, apperrors.ErrInvalidPayload)
		}
	case TypeReport:
		var p ReportPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		if p.ReportID <= 0 || (p.Format != "pdf" && p.Format != "csv") {
			return fmt.Errorf("%w: %w", apperrors.ErrValidation, apperrors.ErrInvalidPayload)
		}
	default:
		return apperrors.ErrInvalidTaskType
	}
	return nil
}

type Task struct {
	ID            int
	Type          TaskType
	Payload       json.RawMessage
	Status        TaskStatus
	Result        json.RawMessage
	Error         string
	Attempts      int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	NextAttemptAt time.Time
}

type TaskStats struct {
	Pending   int
	Running   int
	Completed int
	Failed    int
}
