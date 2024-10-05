package grpcapp

import (
	"fmt"
	authGRPC "gRPCauth/internal/grpc/auth"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, authService authGRPC.Auth, port int) *App {
	server := grpc.NewServer()
	authGRPC.RegisterServerAPI(server, authService)

	return &App{
		log:        log,
		gRPCServer: server,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcApp.run"
	log := a.log.With(slog.String("op", op), slog.Int("port", a.port))
	log.Info("starting gRPC server")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC server running", slog.Int("port", a.port), slog.String("address", lis.Addr().String()))

	if err = a.gRPCServer.Serve(lis); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {
	const op = "grpcApp.stop"
	log := a.log.With(slog.String("op", op))
	log.Info("stopping gRPC server")
	a.gRPCServer.GracefulStop()
}
