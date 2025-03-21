package main

import (
	"log/slog"
	"os"
	"service-propusk-backend/internal/config"
	"service-propusk-backend/internal/pkg/logger/handlers/slogpretty"
	"service-propusk-backend/internal/pkg/logger/sl"
	"service-propusk-backend/internal/repository/postgres"
	postgresBuilds "service-propusk-backend/internal/repository/postgres/buildings"
	minioRepository "service-propusk-backend/internal/repository/s3minio"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	//minioEndpoint := "localhost:9000"
	//accessKey := "minioadminaccesskey"
	//secretKey := "iamilpass"
	//accessKey := "minioadmin"
	//secretKey := "minioadmin"
	buildingsPhotosBucketName := "services"
	staticFelisBucketName := "static"
	buildsPhotosPath := "data/s3Files/buildsMainPhotos/"
	staticFilesPath := "data/s3Files/static/"

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	postgresPool, err := postgres.NewConnPool(&cfg.Postgresql)
	if err != nil {
		panic(err.Error())
	}

	buildRepository, err := postgresBuilds.New(postgresPool)
	if err != nil {
		log.Error("Error creating postgres build repository", sl.Err(err))
	}

	s3Session, err := minioRepository.NewConn(
		&cfg.S3minio,
	)
	if err != nil {
		log.Error("Error creating S3 connection", sl.Err(err))
	}

	s3Repo := minioRepository.New(
		s3Session,
		buildingsPhotosBucketName,
		staticFelisBucketName,
		cfg.S3minio.QRCodesBucketName,
		buildsPhotosPath,
		staticFilesPath,
		buildRepository,
	)

	if err := s3Repo.ConfigureMinioStorage(); err != nil {
		log.Error("configure minio storage failed", sl.Err(err))
	}

	if err := s3Repo.SyncBuildsPhotos(); err != nil {
		log.Error("sync builds photo failed", sl.Err(err))
	}
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
