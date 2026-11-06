package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"task-scheduler/internal/model"
	"task-scheduler/internal/service"
	"testing"
)

type mockRepository struct {
	service.TaskRepository

	tasksToReturn []model.Task
	errToReturn   error
}

func (m *mockRepository) GetTasks(ctx context.Context) ([]model.Task, error) {
	return m.tasksToReturn, m.errToReturn
}

func TestGetHandler(t *testing.T) {

	svc := service.NewTaskService(&mockRepository{
		tasksToReturn: []model.Task{
			{ID: 1, Type: model.TypeEmail},
		},
		errToReturn: nil,
	})
	handler := NewTaskHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()

	handler.GetTasks(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
	}

	if header := rec.Header().Get("Content-Type"); header != "application/json" {
		t.Errorf("got header %s, want %s", header, "application/json")
	}

	var tasks []model.Task

	if err := json.Unmarshal(rec.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("failed to unmarshal res: %v", err)
	}

	if len(tasks) != 1 || tasks[0].ID != 1 {
		t.Errorf("unexpected tasks in response: %+v", tasks)
	}
}
