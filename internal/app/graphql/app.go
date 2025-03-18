package graphqlapp

import (
	"log/slog"
	"rip/graph"

	"rip/graph/generated"
	buildingService "rip/internal/service/building"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

type Config struct {
	Port string `yaml:"port" env-default:"8082"`
	Addr string `yaml:"address" env-default:"0.0.0.0"`
}

type App struct {
	log             *slog.Logger
	port            string
	addr            string
	buildingService *buildingService.BuildingService
	router          *gin.Engine
}

func New(
	log *slog.Logger,
	port string,
	addr string,
	buildingService *buildingService.BuildingService,
) *App {
	router := gin.Default()

	// Configure CORS
	router.Use(corsMiddleware())

	app := &App{
		log:             log,
		port:            port,
		buildingService: buildingService,
		addr:            addr,
		router:          router,
	}

	app.setupRoutes()

	return app
}

func (a *App) setupRoutes() {
	// GraphQL endpoint
	a.router.POST("/query", a.graphqlHandler())
	// Playground
	a.router.GET("/", a.playgroundHandler())
}

func (a *App) graphqlHandler() gin.HandlerFunc {

	graphResolver := graph.NewResolver(a.buildingService)

	h := handler.NewDefaultServer(
		generated.NewExecutableSchema(
			generated.Config{
				Resolvers: graphResolver,
			},
		),
	)

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

type Resolver struct {
	buildingService *buildingService.BuildingService
}

func (a *App) playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	a.log.Info(
		"starting graphql server",
		slog.String("port", a.port),
	)

	return a.router.Run(a.addr)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
		)
		c.Writer.Header().Set(
			"Access-Control-Allow-Methods",
			"POST, OPTIONS, GET, PUT, DELETE",
		)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
