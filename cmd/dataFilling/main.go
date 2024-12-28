package main

import (
	"fmt"
	minioRepository "rip/internal/repository/minio"
	"rip/internal/repository/postgres"
	postgresBuilds "rip/internal/repository/postgres/builds"
)

func main() {
	minioEndpoint := "localhost:9000"
	accessKey := "minioadminaccesskey"
	secretKey := "iamilpass"
	buildingsPhotosBucketName := "services"
	staticFelisBucketName := "static"
	buildsPhotosPath := "data/s3Files/buildsMainPhotos/"
	staticFilesPath := "data/s3Files/static/"

	postgresPool, err := postgres.NewConnPool()
	if err != nil {
		fmt.Println(err.Error())
	}

	buildRepository, err := postgresBuilds.New(postgresPool)
	if err != nil {
		fmt.Println(err.Error())

	}

	s3Session, err := minioRepository.Connect(
		minioEndpoint,
		accessKey,
		secretKey,
	)
	if err != nil {
		fmt.Println("Error connecting to S3")
	}

	s3Repo := minioRepository.New(
		s3Session,
		buildingsPhotosBucketName,
		staticFelisBucketName,
		buildsPhotosPath,
		staticFilesPath,
		buildRepository,
	)

	if err := s3Repo.SyncBuildsPhotos(); err != nil {
		fmt.Println(err.Error())
	}
}
