package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	handler_gin_v1 "rip/internal/handler/v1/gin"
	minioRepository "rip/internal/repository/minio"
	"rip/internal/repository/postgres"
	postgresBuilds "rip/internal/repository/postgres/builds"
	postgresPasses "rip/internal/repository/postgres/passes"
	buildService "rip/internal/service/build"
	passService "rip/internal/service/pass"
)

type App struct {
}

func New() *App {
	return &App{}
}

func (*App) MustRun() {
	minioEndpoint := "localhost:9000"
	accessKey := "minioadmin"
	secretKey := "minioadmin"
	buildingsPhotosBucketName := "services"
	staticFelisBucketName := "static"
	buildsPhotosPath := "s3Files/buildsMainPhotos/"
	staticFilesPath := "s3Files/static/"

	hostname := "http://localhost:9000"

	s3Session, err := minioRepository.Connect(
		minioEndpoint,
		accessKey,
		secretKey,
	)
	if err != nil {
		log.Println("Error connecting to S3")
	}

	postgresPool, err := postgres.NewConnPool()
	if err != nil {
		log.Fatal(err)
	}

	defer postgresPool.Close()

	buildRepository, err := postgresBuilds.New(postgresPool)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	passRepository, err := postgresPasses.New(postgresPool)
	if err != nil {
		fmt.Println(err.Error())
		return
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
		log.Fatal("configure minio storage fatal error:", err.Error())

	}

	buildService := buildService.New(buildRepository, hostname)

	passService := passService.New(passRepository, hostname)

	r := gin.Default()

	gin.SetMode(gin.ReleaseMode)

	r.LoadHTMLGlob("internal/templates/*")

	r.GET(
		"/syncs3BuildingsPhotos/", func(c *gin.Context) {
			if err := s3Repo.SyncBuildsPhotos(); err != nil {
				c.Data(
					http.StatusOK,
					"text/html; charset=utf-8",
					[]byte("Error syncing builds photos: "+err.Error()),
				)

				return
			}

			c.Data(
				http.StatusOK,
				"text/html; charset=utf-8",
				[]byte("Updated builds photos success"),
			)
		},
	)

	r.GET(
		"/printbuildingsbucketpolicy", func(c *gin.Context) {
			s3Repo.PrintBuilbingsBucketPolice()
			c.Data(
				http.StatusOK,
				"text/html; charset=utf-8",
				[]byte("policy printed"),
			)
		},
	)

	r.GET(
		"/printstaticbucketpolicy", func(c *gin.Context) {
			s3Repo.PrintStaticBucketPolice()
			c.Data(
				http.StatusOK,
				"text/html; charset=utf-8",
				[]byte("policy printed"),
			)
		},
	)

	r.GET(
		"/", handler_gin_v1.MainPage(buildService, passService),
	)

	r.GET(
		"/pass/:id", handler_gin_v1.PassPage(passService),
	)

	r.GET(
		"/buildings/:id", handler_gin_v1.BuildingPage(buildService),
	)

	r.POST("/add_to_pass/:id", handler_gin_v1.AddToPass(passService))

	r.POST("/pass/delete_pass/:id", handler_gin_v1.DeletePass(passService))

	r.Static("/image", "./static/images")
	r.Static("/style", "./static/styles")

	log.Println("Server start up")

	if err := r.Run(); err != nil {
		log.Fatalf(err.Error())
	}

	log.Println("Server down")
}
