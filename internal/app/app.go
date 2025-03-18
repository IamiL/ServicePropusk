package app

import (
	"log/slog"
	"net"
	graphqlapp "rip/internal/app/graphql"
	httpapp "rip/internal/app/http"
	"rip/internal/app/swagger"
	"rip/internal/repository/JWTsecret"
	"rip/internal/repository/postgres"
	postgresBuilds "rip/internal/repository/postgres/buildings"
	postgresPasses "rip/internal/repository/postgres/passes"
	postgresUser "rip/internal/repository/postgres/user"
	"rip/internal/repository/s3minio"
	minioRepository "rip/internal/repository/s3minio"
	authService "rip/internal/service/auth"
	buildService "rip/internal/service/building"
	passService "rip/internal/service/pass"
	passBuildingService "rip/internal/service/passBuilding"
	userService "rip/internal/service/user"
	"time"
)

type App struct {
	HTTPServer    *httpapp.App
	GraphQLServer *graphqlapp.App
	RunSwagger    func(log *slog.Logger, port string)
}

type Config struct {
	TokenTTL             time.Duration `yaml:"token_ttl" env-default:"300h"`
	LaboratoryWorkNumber string        `yaml:"laboratory_work_number"`
}

func New(
	log *slog.Logger,
	appConfig *Config,
	httpConfig *httpapp.Config,
	postgresConfig *postgres.Config,
	s3minioConfig *s3minio.Config,
	graphqlConfig *graphqlapp.Config,
) (
	*App,
	func(),
) {
	s3MinioConn, err := minioRepository.NewConn(s3minioConfig)
	if err != nil {
		log.Error("Error connecting to S3")
	}

	postgresPool, err := postgres.NewConnPool(postgresConfig)
	if err != nil {
		panic(err.Error())
	}

	postgresClose := func() { postgresPool.Close() }

	buildRepository, err := postgresBuilds.New(postgresPool)
	if err != nil {
		panic(err.Error())
	}

	passRepository, err := postgresPasses.New(postgresPool)
	if err != nil {
		panic(err.Error())
	}

	userRepository, _ := postgresUser.New(postgresPool)

	s3Repo := minioRepository.New(
		s3MinioConn,
		s3minioConfig.BuildingsPhotosBucketName,
		s3minioConfig.StaticFilesBucketName,
		s3minioConfig.QRCodesBucketName,
		s3minioConfig.BuildsPhotosLocalPath,
		s3minioConfig.StaticFilesLocalPath,
		buildRepository,
	)

	secretStorage := JWTsecret.NewJWTSecret([]byte("secret"))

	authService := authService.New(
		log,
		appConfig.TokenTTL,
		userRepository,
		secretStorage,
	)

	userService := userService.New(userRepository, userRepository, authService)

	buildingService := buildService.New(
		log,
		buildRepository,
		buildRepository,
		buildRepository,
		s3Repo,
		s3Repo,
		authService,
		net.JoinHostPort(s3minioConfig.Host, s3minioConfig.Port),
	)

	passService := passService.New(
		log,
		passRepository,
		passRepository,
		passRepository,
		passRepository,
		buildRepository,
		authService,
		net.JoinHostPort(s3minioConfig.Host, s3minioConfig.Port),
		s3Repo,
	)

	passBuildingService := passBuildingService.New(
		log,
		passRepository,
		passRepository,
		authService,
	)

	httpApp := httpapp.New(
		log,
		httpConfig,
		buildingService,
		passService,
		passBuildingService,
		userService,
		authService,
		httpConfig.Tls,
	)

	graphqlServer := graphqlapp.New(
		log,
		graphqlConfig.Port,
		graphqlConfig.Addr,
		buildingService,
	)

	return &App{
		HTTPServer:    httpApp,
		GraphQLServer: graphqlServer,
		RunSwagger:    swagger.RunSwagger,
	}, postgresClose
}

func (a *App) Run() error {
	// Start HTTP server
	go a.HTTPServer.MustRun()

	// Start GraphQL server
	go a.GraphQLServer.MustRun()

	return nil
}
