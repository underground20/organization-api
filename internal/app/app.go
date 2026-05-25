package app

import (
	"app/internal/config"
	"app/internal/handler"
	loggingMiddleware "app/internal/http/middleware/logger"
	storage "app/internal/storage/postgres"
	"app/internal/validation"
	logErr "app/lib/logger"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	cfg    *config.Config
	server *http.Server
	logger *slog.Logger
}

func New(logger *slog.Logger, cfg *config.Config) *App {
	db, err := dbConnect(cfg)
	if err != nil {
		logger.Error("failed to connect to database", logErr.Err(err))
		os.Exit(1)
	}

	router := setupRouter(logger)
	storage := storage.NewStorage(db, logger)
	newHandler := handler.NewHandler(storage, logger, validation.New())

	router.GET("/departments/:id", newHandler.GetDepartment)
	router.POST("/departments", newHandler.CreateDepartment)
	router.POST("/departments/:id/employees", newHandler.AddEmployee)
	router.PATCH("/departments/:id", newHandler.ChangeParent)
	router.DELETE("/departments/:id", newHandler.DeleteDepartment)

	srv := &http.Server{
		Addr:    ":" + cfg.HostPort,
		Handler: router,
	}

	return &App{
		logger: logger,
		cfg:    cfg,
		server: srv,
	}
}

func (a *App) Run() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.logger.Info("starting http server", slog.String("addr", a.server.Addr))
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("http server failed to start", logErr.Err(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	a.logger.Info("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("server forced to shutdown", logErr.Err(err))
	}

	wg.Wait()
	a.logger.Info("server exiting")
}

func dbConnect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func setupRouter(logger *slog.Logger) *gin.Engine {
	router := gin.Default()
	router.Use(loggingMiddleware.SlogMiddleware(logger))
	return router
}
