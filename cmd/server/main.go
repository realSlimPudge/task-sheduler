package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"task-scheduler/config"
	"task-scheduler/internal/database"
	"task-scheduler/internal/handler"
	"task-scheduler/internal/middleware"
	"task-scheduler/internal/model"
	"task-scheduler/internal/repository"
	"task-scheduler/internal/service"
	"task-scheduler/internal/worker"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки конфига ", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	db, err := database.NewPostgresDB(cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("БД успешно подключена")
	defer db.Close()

	postgresRepo := repository.NewPostgresRepo(db)

	restored, err := postgresRepo.RestoreZombie()
	if err != nil {
		slog.Error("error restore zombie tasks", "error", err)
	} else if restored > 0 {
		slog.Info("restored zombie tasks", "count", restored)
	}

	workerPool := worker.NewWorkerPool(postgresRepo, cfg.WorkerCount)
	workerPool.RegisterHandler(model.TypeEmail, worker.ProcessEmail)
	workerPool.RegisterHandler(model.TypeImage, worker.ProcessImage)
	workerPool.RegisterHandler(model.TypeReport, worker.ProcessReport)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	workerPool.Start(workerCtx, 2*time.Second)

	service := service.NewTaskService(postgresRepo)
	handler := handler.NewTaskHandler(service)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /tasks", handler.GetTasks)
	mux.HandleFunc("GET /tasks/{id}", handler.GetTaskByID)
	mux.HandleFunc("GET /tasks/stats", handler.GetStats)
	mux.HandleFunc("POST /task", handler.CreateTask)

	wrappedMux := middleware.RequestID(middleware.Logging(middleware.Recover(mux)))

	server := &http.Server{
		Addr:        ":" + cfg.Port,
		Handler:     wrappedMux,
		ReadTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("Сервер успешно запущен на порте: %s", server.Addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server crash: %v", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	workerCancel()
	workerPool.Stop()
	if err != nil {
		log.Print("Не получилось выключить нормально")
		if err := server.Close(); err != nil {
			log.Print("Не получилось выключить никак")
		}
		return
	}

	log.Print("Сервер успешно остановлен")
}
