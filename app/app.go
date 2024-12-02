package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	handler_gin_v1 "rip/handler/v1/gin"
	"rip/repository/postgres"
	postgresBuilds "rip/repository/postgres/builds"
	postgresPasses "rip/repository/postgres/passes"
	s3Repository "rip/repository/s3"
	buildService "rip/service/build"
	passService "rip/service/pass"
)

type App struct {
}

func New() *App {
	return &App{}
}

func (*App) MustRun() {
	hostname := "http://localhost:9000"
	buildsPhotosBucketName := ""
	//buildsPhotosPath := "buildsMainPhotos"

	s3Repository.Connect("minio", "minio124", buildsPhotosBucketName)

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

	buildService := buildService.New(buildRepository, hostname)

	passService := passService.New(passRepository, hostname)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	//r.GET(
	//	"/syncs3BuildingsPhotos/", func(c *gin.Context) {
	//		s3Session := s3Repository.Connect()
	//		s3Repo := s3Repository.New(
	//			s3Session,
	//			buildsPhotosBucketName,
	//			buildsPhotosPath,
	//		)
	//
	//		if err := s3Repo.SyncBuildsPhotos(buildRepository); err != nil {
	//			c.Data(
	//				http.StatusOK,
	//				"text/html; charset=utf-8",
	//				[]byte("Error syncing builds photos: "+err.Error()),
	//			)
	//		}
	//	},
	//)

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
