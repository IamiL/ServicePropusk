package main

import (
	"log"

	_ "rip/docs" // Import swagger docs

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	router := gin.Default()

	router.Use(
		cors.New(
			cors.Config{
				AllowOrigins: []string{
					"http://localhost:8080",
					"http://localhost:8081",
				},
				AllowMethods: []string{
					"GET",
					"POST",
					"PUT",
					"DELETE",
					"OPTIONS",
					"HEAD",
				},
				AllowHeaders: []string{
					"Origin",
					"Content-Type",
					"Authorization",
					"Accept",
					"X-Requested-With",
				},
				ExposeHeaders:    []string{"Content-Length"},
				AllowCredentials: true,
				MaxAge:           12 * 3600, // 12 hours
			},
		),
	)

	router.GET(
		"/swagger/*any", ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.URL("http://localhost:8081/swagger/doc.json"),
			ginSwagger.DefaultModelsExpandDepth(1),
			ginSwagger.DocExpansion("none"),
		),
	)

	router.Static("/docs", "./docs")

	if err := router.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
