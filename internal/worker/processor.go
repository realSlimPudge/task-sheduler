package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"task-scheduler/internal/model"
	"time"
)

type TaskProcessor func(ctx context.Context, payload json.RawMessage) (json.RawMessage, error)

type EmailResult struct {
	Status string    `json:"status"`
	SentAt time.Time `json:"sentAt"`
}

type ImageResult struct {
	OldURL string `json:"oldUrl"`
	NewURL string `json:"newUrl"`
}

type ReportResult struct {
	Report string `json:"report"`
}

func ProcessEmail(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var emailPayload model.EmailPayload
	if err := json.Unmarshal(payload, &emailPayload); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(3 * time.Second):
	}

	slog.Info("send email", "to", emailPayload.To, "body", emailPayload.Body, "subject", emailPayload.Subject)

	result := EmailResult{
		Status: "sent",
		SentAt: time.Now(),
	}
	resByte, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(resByte), nil
}

func ProcessImage(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var imagePayload model.ImagePayload
	if err := json.Unmarshal(payload, &imagePayload); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
	}

	slog.Info("processing image", "width", imagePayload.Width, "height", imagePayload.Height, "url", imagePayload.URL)

	result := ImageResult{
		OldURL: imagePayload.URL,
		NewURL: "newURL/" + imagePayload.URL,
	}
	resByte, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(resByte), nil
}

func ProcessReport(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var reportPayload model.ReportPayload
	if err := json.Unmarshal(payload, &reportPayload); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(10 * time.Second):
	}

	slog.Info("generating report", "format", reportPayload.Format, "id", reportPayload.ReportID)

	ID := strconv.Itoa(reportPayload.ReportID)

	result := ReportResult{
		Report: "report_" + ID + "." + reportPayload.Format,
	}
	resByte, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(resByte), nil
}
