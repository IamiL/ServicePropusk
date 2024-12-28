package app

import (
	"fmt"
	"log/slog"
	httpapp "rip/internal/app/http"
	inMemorySession "rip/internal/repository/inMemory"
	minioRepository "rip/internal/repository/minio"
	"rip/internal/repository/postgres"
	postgresBuilds "rip/internal/repository/postgres/builds"
	postgresPasses "rip/internal/repository/postgres/passes"
	postgresUser "rip/internal/repository/postgres/user"
	buildService "rip/internal/service/build"
	passService "rip/internal/service/pass"
	userService "rip/internal/service/user"
)

type App struct {
	HTTPServer *httpapp.App
}

func New(log *slog.Logger, config string) *App {
	minioEndpoint := "localhost:9000"
	accessKey := "minioadminaccesskey"
	secretKey := "iamilpass"
	buildingsPhotosBucketName := "services"
	staticFelisBucketName := "static"
	buildsPhotosPath := "data/s3Files/buildsMainPhotos/"
	staticFilesPath := "data/s3Files/static/"
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

	tokenStore := inMemorySession.New()

	//defer postgresPool.Close(context.Background())

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

	userRepository, _ := postgresUser.New(postgresPool)

	s3Repo := minioRepository.New(
		s3Session,
		buildingsPhotosBucketName,
		staticFelisBucketName,
		buildsPhotosPath,
		staticFilesPath,
		buildRepository,
	)

	buildingService := buildService.New(
		buildRepository,
		buildRepository,
		buildRepository,
		s3Repo,
		s3Repo,
		buildingsImageHostname,
	)

	passService := passService.New(
		passRepository,
		passRepository,
		passRepository,
		passRepository,
		buildingsImageHostname,
	)

	userService := userService.New(tokenStore, userRepository, userRepository)

	httpApp := httpapp.New(
		log,
		port,
		buildingService,
		passService,
		userService,
	)
	return &App{HTTPServer: httpApp}
}
