package main

import (
	"gRPCauth/internal/app"
	"gRPCauth/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	//TODO: инициалиировать объект конфига
	cfg := config.MustLoad()
	//TODO:  Инициализировать логгер
	log := setupLogger(cfg.Env)
	log.Info("starting server")
	//TODO: инициализация приложения (app)
	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	go application.GRPCsrv.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	sign := <-stop
	log.Info("stopping app", slog.String("reason", sign.String()))
	application.GRPCsrv.Stop()
	log.Info("shutting down server")
	//TODO: запустить gRPC-сервер приложения
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return logger
}
