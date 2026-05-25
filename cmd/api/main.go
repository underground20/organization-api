package main

import (
	"app/internal/app"
	"app/internal/config"
	"log/slog"
	"os"
)

func main() {
	logger := setupLogger()
	cfg := config.MustLoad()
	a := app.New(logger, cfg)
	a.Run()
}

func setupLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
