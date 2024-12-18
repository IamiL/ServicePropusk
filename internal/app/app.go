package app

import (
	"fmt"
	"log/slog"
	httpapp "rip/internal/app/http"
	minioRepository "rip/internal/repository/minio"
	"rip/internal/repository/postgres"
	postgresBuilds "rip/internal/repository/postgres/builds"
	postgresPasses "rip/internal/repository/postgres/passes"
	buildService "rip/internal/service/build"
	passService "rip/internal/service/pass"
	userService "rip/internal/service/user"
)

type App struct {
	HTTPServer *httpapp.App
}

func New(log *slog.Logger, config string) *App {
	minioEndpoint := "localhost:9000"
	accessKey := "minioadmin"
	secretKey := "minioadmin"
	buildingsPhotosBucketName := "services"
	staticFelisBucketName := "static"
	buildsPhotosPath := "s3Files/buildsMainPhotos/"
	staticFilesPath := "s3Files/static/"
	port := 8080

	buildingsImageHostname := "http://localhost:9000"

	s3Session, err := minioRepository.Connect(
		minioEndpoint,
		accessKey,
		secretKey,
	)
	if err != nil {
		log.Error("Error connecting to S3")
	}

	postgresPool, err := postgres.NewConnPool()
	if err != nil {
		log.Error("", err)
	}

	defer postgresPool.Close()

	buildRepository, err := postgresBuilds.New(postgresPool)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	passRepository, err := postgresPasses.New(postgresPool)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	s3Repo := minioRepository.New(
		s3Session,
		buildingsPhotosBucketName,
		staticFelisBucketName,
		buildsPhotosPath,
		staticFilesPath,
		buildRepository,
	)

	if err := s3Repo.ConfigureMinioStorage(); err != nil {
		log.Error("configure minio storage fatal error:", err.Error())

	}

	buildingService := buildService.New(buildRepository, buildingsImageHostname)

	passService := passService.New(
		passRepository,
		passRepository,
		buildingsImageHostname,
	)

	userService := userService.New()

	httpApp := httpapp.New(
		log,
		port,
		buildingService,
		passService,
		userService,
	)
	return &App{HTTPServer: httpApp}
}
