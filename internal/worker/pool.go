package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"task-scheduler/internal/model"
	"time"
)

type WorkerDB interface {
	GetPendingTasks(ctx context.Context, limit int) ([]model.Task, error)
	MarkCompleted(ctx context.Context, id int, result json.RawMessage) error
	MarkFailed(ctx context.Context, id int, taskErr string) error
	MarkRetry(ctx context.Context, id int, taskErr string, nextAttemptAt time.Time) error
}

type WorkerPool struct {
	db          WorkerDB
	queue       chan model.Task
	processor   map[model.TaskType]TaskProcessor
	workerCount int
	wg          sync.WaitGroup
}

func NewWorkerPool(db WorkerDB, workerCount int) *WorkerPool {
	return &WorkerPool{
		db:          db,
		queue:       make(chan model.Task),
		processor:   make(map[model.TaskType]TaskProcessor),
		workerCount: workerCount,
		wg:          sync.WaitGroup{},
	}
}

func (p *WorkerPool) RegisterHandler(taskType model.TaskType, handler TaskProcessor) {
	p.processor[taskType] = handler
}

func (p *WorkerPool) worker(ctx context.Context, id int) {
	defer p.wg.Done()
	const maxAttempts = 3
	const baseDelay = 4 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-p.queue:
			if !ok {
				return
			}
			slog.Info("working", "worker id", id, "task id", task.ID)
			processor, ok := p.processor[task.Type]
			if !ok {
				slog.Error("invalid task type", "taskType", task.Type)
				p.db.MarkFailed(ctx, task.ID, "invalid task type")
				continue
			}

			result, err := processor(ctx, task.Payload)
			if err != nil {
				slog.Error("worker's handler error", "error", err)
				if task.Attempts+1 < maxAttempts {
					multiplier := 1 << task.Attempts
					delay := baseDelay * time.Duration(multiplier)
					nextAttemptAt := time.Now().Add(delay)

					p.db.MarkRetry(ctx, task.ID, err.Error(), nextAttemptAt)
					continue
				}
				p.db.MarkFailed(ctx, task.ID, err.Error())
				continue
			}

			p.db.MarkCompleted(ctx, task.ID, result)
		}
	}
}

func (p *WorkerPool) dispatcher(ctx context.Context, pollInterval time.Duration) {
	defer p.wg.Done()

	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			const batchSize = 10
			tasks, err := p.db.GetPendingTasks(ctx, batchSize)
			if err != nil {
				slog.Error("dispatcher error", "err", err)
				timer.Reset(pollInterval)
				continue
			}

			for _, task := range tasks {
				select {
				case p.queue <- task:
				case <-ctx.Done():
					return
				}
			}

			if len(tasks) == batchSize {
				timer.Reset(0)
			} else {
				timer.Reset(pollInterval)
			}
		}
	}
}

func (p *WorkerPool) Start(ctx context.Context, pollInterval time.Duration) {
	for w := 1; w <= p.workerCount; w++ {
		p.wg.Add(1)
		go p.worker(ctx, w)
	}

	p.wg.Add(1)
	go p.dispatcher(ctx, pollInterval)
}

func (p *WorkerPool) Stop() {
	p.wg.Wait()
}
