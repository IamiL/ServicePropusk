package httpapp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	buildService "rip/internal/service/build"
	passService "rip/internal/service/pass"
	userService "rip/internal/service/user"
	"time"
)

type App struct {
	log        *slog.Logger
	httpServer *http.Server
	port       int
}

func New(
	log *slog.Logger,
	port int,
	buildingService *buildService.BuildingService,
	passService *passService.PassService,
	userService *userService.UserService,
) *App {
	router := http.NewServeMux()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	return &App{log: log, httpServer: srv, port: port}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "httpapp.Run"

	if err := a.httpServer.ListenAndServe(); err != nil {
		a.log.Error("failed to start http server")
	}

	a.log.With(slog.String("op", op)).
		Info("server started", slog.Int("port", a.port))

	return nil
}

func (a *App) Stop() {
	const op = "httpapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping HTTP server", slog.Int("port", a.port))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.log.Error("server closed with err: %+v", err)
		os.Exit(1)
	}

	a.log.Info("Gracefully stopped")
}
