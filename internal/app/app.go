package app

import (
	"log/slog"
	"net"
	graphqlapp "service-propusk-backend/internal/app/graphql"
	httpapp "service-propusk-backend/internal/app/http"
	"service-propusk-backend/internal/app/swagger"
	"service-propusk-backend/internal/repository/JWTsecret"
	"service-propusk-backend/internal/repository/postgres"
	postgresBuilds "service-propusk-backend/internal/repository/postgres/buildings"
	postgresPasses "service-propusk-backend/internal/repository/postgres/passes"
	postgresUser "service-propusk-backend/internal/repository/postgres/user"
	"service-propusk-backend/internal/repository/s3minio"
	minioRepository "service-propusk-backend/internal/repository/s3minio"
	authService "service-propusk-backend/internal/service/auth"
	buildService "service-propusk-backend/internal/service/building"
	passService "service-propusk-backend/internal/service/pass"
	passBuildingService "service-propusk-backend/internal/service/passBuilding"
	userService "service-propusk-backend/internal/service/user"
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
