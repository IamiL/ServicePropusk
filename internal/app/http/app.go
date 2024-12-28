package httpapp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	handler_mux_v1 "rip/internal/http-server/handlers/building"
	passhandler "rip/internal/http-server/handlers/pass"
	passBuildingHandler "rip/internal/http-server/handlers/passBuilding"
	userHandler "rip/internal/http-server/handlers/user"
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

	router.HandleFunc(
		"GET /buildings",
		handler_mux_v1.BuildingsHandler(buildingService, passService),
	)
	router.HandleFunc(
		"GET /buildings/{id}",
		handler_mux_v1.BuildingHandler(buildingService),
	)
	router.HandleFunc(
		"POST /buildings",
		handler_mux_v1.NewBuildingHandler(buildingService),
	)
	router.HandleFunc(
		"PUT /buildings/{id}",
		handler_mux_v1.EditBuildingHandler(buildingService),
	)
	router.HandleFunc(
		"DELETE /buildings/{id}",
		handler_mux_v1.DeleteBuildingHandler(buildingService),
	)
	router.HandleFunc(
		"POST /buildings/{id}/topass",
		handler_mux_v1.AddToPassHandler(log, passService),
	)
	router.HandleFunc(
		"POST /buildings/{id}/preview/save",
		handler_mux_v1.AddBuildingPreview(buildingService),
	)

	router.HandleFunc("GET /passes", passhandler.PassesHandler(passService))
	router.HandleFunc("GET /passes/{id}", passhandler.PassHandler(passService))
	router.HandleFunc(
		"PUT /passes/{id}",
		passhandler.EditPassHandler(passService),
	)
	router.HandleFunc(
		"PUT /passes/{id}/toForm",
		passhandler.ToFormHandler(passService),
	)
	router.HandleFunc(
		"PUT /passes/{id}/reject",
		passhandler.RejectPassHandler(passService),
	)
	router.HandleFunc(
		"PUT /passes/{id}/complete",
		passhandler.CompletePassHandler(passService),
	)
	router.HandleFunc(
		"DELETE /passes/{id}",
		passhandler.DeletePassHandler(passService),
	)

	router.HandleFunc(
		"DELETE /pass/deletebuilding/{id}",
		passBuildingHandler.DeletePassBuilding(),
	)

	router.HandleFunc(
		"POST /users",
		userHandler.RegistrationHandler(userService),
	)
	router.HandleFunc("PUT /users", userHandler.EditUserHandler())
	router.HandleFunc(
		"POST /users/login",
		userHandler.SigninHandler(userService),
	)
	router.HandleFunc(
		"POST /users/logout",
		userHandler.DeauthorizationHandler(userService),
	)
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

	a.log.With(slog.String("op", op)).
		Info("server started", slog.Int("port", a.port))

	if err := a.httpServer.ListenAndServe(); err != nil {
		a.log.Error("failed to start http server")
	}

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
