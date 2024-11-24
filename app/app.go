package app

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	handler_gin_v1 "rip/handler/v1/gin"
	buildRepositoryPostgres "rip/repository/builds/postgres"
	passRepositoryPostgres "rip/repository/passes/postgres"
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

	postgresPool, err := pgxpool.New(
		context.Background(),
		fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			"localhost",
			"5432",
			"iamil-admin",
			"adminpass",
			"service-propusk",
		),
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer postgresPool.Close()

	buildRepository, err := buildRepositoryPostgres.New(postgresPool)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	passRepository, err := passRepositoryPostgres.New(postgresPool)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	buildService := buildService.New(buildRepository, hostname)

	passService := passService.New(passRepository, buildService, hostname)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.GET(
		"/", handler_gin_v1.MainPage(buildService, passService),
	)

	r.GET(
		"/pass/:id", handler_gin_v1.PassPage(passService),
	)

	r.GET(
		"/buildings/:id", handler_gin_v1.BuildPage(buildService),
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
