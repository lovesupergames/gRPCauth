package app

import (
	grpcapp "gRPCauth/internal/app/gprc"
	"gRPCauth/internal/services/auth"
	"gRPCauth/storage/sqlite"
	"log/slog"
	"time"
)

type App struct {
	GRPCsrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	//TODO: инициализировать хранилише (storage)
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	// TODO: инициализировать auth service (auth)
	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)
	return &App{
		GRPCsrv: grpcApp,
	}
}
