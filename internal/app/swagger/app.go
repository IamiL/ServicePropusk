package swagger

import (
	"fmt"
	"log"

	_ "rip/docs" // Import swagger docs

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RunSwagger(port string) {
	router := gin.Default()

	// Configure CORS
	router.Use(
		cors.New(
			cors.Config{
				AllowOrigins: []string{
					"http://localhost:8080",
					"http://localhost:8081",
					"http://localhost:8000",
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

	// Add Swagger UI with custom configuration
	router.GET(
		"/swagger/*any", ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.URL("http://localhost:8002/swagger/doc.json"),
			ginSwagger.DefaultModelsExpandDepth(1),
			ginSwagger.DocExpansion("none"),
		),
	)

	// Serve static files from docs directory
	router.Static("/docs", "./docs")

	// Start server
	fmt.Println("Swagger UI is available at http://localhost:8002/swagger/index.html")
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
