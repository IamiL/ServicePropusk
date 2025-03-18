package main

import (
	"log/slog"
	"os"
	"os/signal"
	_ "rip/docs" // Import swagger docs
	appService "rip/internal/app"
	"rip/internal/config"
	"rip/internal/pkg/logger/handlers/slogpretty"
	"syscall"
)

// @title           RIP API
// @version         1.0
// @description     API for managing buildings and passes.
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8000
// @BasePath  /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Access-Token
// @description Enter your JWT token value
// @security ApiKeyAuth

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	application, deferFunc := appService.New(
		log,
		&cfg.App,
		&cfg.HTTP,
		&cfg.Postgresql,
		&cfg.S3minio,
		&cfg.GraphQL,
	)

	go application.HTTPServer.MustRun()
	go application.GraphQLServer.MustRun()
	go application.RunSwagger(log, "8002")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
	log.Info("stopping server")

	application.HTTPServer.Stop()

	deferFunc()

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelDebug},
			),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelInfo},
			),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
