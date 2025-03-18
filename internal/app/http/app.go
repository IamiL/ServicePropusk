package httpapp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	handler_mux_v1 "rip/internal/http-server/handlers/building"
	passhandler "rip/internal/http-server/handlers/pass"
	passBuildingHandler "rip/internal/http-server/handlers/passBuilding"
	userHandler "rip/internal/http-server/handlers/user"
	"rip/internal/pkg/logger/sl"
	authService "rip/internal/service/auth"
	buildingService "rip/internal/service/building"
	passService "rip/internal/service/pass"
	passBuildingService "rip/internal/service/passBuilding"
	userService "rip/internal/service/user"
	"time"

	_ "rip/docs" // Import swagger docs

	httpSwagger "github.com/swaggo/http-swagger"
)

type Config struct {
	Port    int           `yaml:"port" default:"8080"`
	Timeout time.Duration `yaml:"timeout"`
	Tls     bool          `yaml:"tls" default:"true"`
	Addr    string        `yaml:"address" default:"0.0.0.0"`
}

type App struct {
	log        *slog.Logger
	httpServer *http.Server
	port       int
	Tls        bool
}

func New(
	log *slog.Logger,
	config *Config,
	buildingService *buildingService.BuildingService,
	passService *passService.PassService,
	passBuildingService *passBuildingService.PassBuildingService,
	userService *userService.UserService,
	authService *authService.AuthService,
	https bool,
) *App {
	router := http.NewServeMux()

	// Add Swagger UI endpoint
	router.HandleFunc(
		"GET /swagger/*", httpSwagger.Handler(
			httpSwagger.URL("http://localhost:8000/swagger/doc.json"),
			httpSwagger.DeepLinking(true),
			httpSwagger.DocExpansion("none"),
		),
	)

	router.HandleFunc(
		"GET /buildings",
		handler_mux_v1.BuildingsHandler(log, buildingService, passService),
	)
	router.HandleFunc(
		"GET /buildings/{id}",
		handler_mux_v1.BuildingHandler(buildingService),
	)
	router.HandleFunc(
		"POST /buildings",
		handler_mux_v1.NewBuildingHandler(log, buildingService),
	)
	router.HandleFunc(
		"PUT /buildings/{id}",
		handler_mux_v1.EditBuildingHandler(log, buildingService),
	)
	router.HandleFunc(
		"DELETE /buildings/{id}",
		handler_mux_v1.DeleteBuildingHandler(log, buildingService),
	)
	router.HandleFunc(
		"POST /buildings/{id}/draft",
		handler_mux_v1.AddToPassHandler(log, passService),
	)
	router.HandleFunc(
		"POST /buildings/{id}/image",
		handler_mux_v1.AddBuildingPreview(log, buildingService),
	)

	router.HandleFunc(
		"GET /passes",
		passhandler.PassesHandler(log, passService),
	)
	router.HandleFunc(
		"GET /passes/{id}",
		passhandler.PassHandler(log, passService),
	)
	router.HandleFunc(
		"PUT /passes/{id}",
		passhandler.EditPassHandler(log, passService),
	)
	router.HandleFunc(
		"PUT /passes/{id}/submit",
		passhandler.ToFormHandler(log, passService),
	)
	router.HandleFunc(
		"PUT /passes/{id}/reject",
		passhandler.RejectPassHandler(log, passService),
	)
	router.HandleFunc(
		"PUT /passes/{id}/complete",
		passhandler.CompletePassHandler(log, passService),
	)
	router.HandleFunc(
		"DELETE /passes/{id}",
		passhandler.DeletePassHandler(log, passService),
	)

	router.HandleFunc(
		"DELETE /passes/{passId}/building/{buildingId}",
		passBuildingHandler.DeletePassBuilding(log, passBuildingService),
	)
	router.HandleFunc(
		"PUT /passes/{passId}/building/{buildingId}/main",
		passBuildingHandler.PutPassBuilding(log, passBuildingService),
	)

	router.HandleFunc(
		"POST /users",
		userHandler.RegistrationHandler(log, userService),
	)
	router.HandleFunc(
		"PUT /users",
		userHandler.EditUserHandler(log, userService),
	)
	router.HandleFunc(
		"POST /users/login",
		userHandler.SigninHandler(log, authService, https),
	)
	router.HandleFunc(
		"POST /users/logout",
		userHandler.LogoutHandler(log, userService),
	)

	// Применяем middleware в правильном порядке
	handler := tokenToHeaderMiddleware(router)
	handler = corsMiddleware(handler)

	srv := &http.Server{
		Addr:    config.Addr,
		Handler: handler,
	}

	fmt.Println(config.Port)

	return &App{log: log, httpServer: srv, port: config.Port, Tls: config.Tls}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "httpapp.Run"

	a.log.With(slog.String("op", op)).
		Info("server started", slog.Int("port", a.port))

	if a.Tls {
		if err := a.httpServer.ListenAndServeTLS(
			"server.crt",
			"server.key",
		); err != nil {
			a.log.Error("failed to start https server", sl.Err(err))
			fmt.Println("err: ", err.Error())
		}
	} else {
		if err := a.httpServer.ListenAndServe(); err != nil {
			a.log.Error("failed to start http server", sl.Err(err))
			fmt.Println("err: ", err.Error())
		}
	}

	return nil
}

func (a *App) Stop() {
	const op = "httpapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping HTTP server", slog.Int("port", a.port))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.log.Error("server closed with err: %+v", err)
		os.Exit(1)
	}

	a.log.Info("Gracefully stopped")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("origin: ", r.Header.Get("origin"))

			// Определяем Origin запроса
			origin := r.Header.Get("Origin")
			allowedOrigins := map[string]bool{
				"http://tauri.localhost":    true,
				"https://localhost:3000":    true, // React dev server
				"http://195.19.39.177:3000": true,
				"https://iamil.github.io":   true,
				"http://localhost:8002":     true, // Swagger UI
			}

			// Устанавливаем CORS заголовки только для разрешенных origins
			if allowedOrigins[origin] {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set(
					"Access-Control-Allow-Methods",
					"GET, POST, OPTIONS, PUT, DELETE, PATCH, HEAD",
				)
				w.Header().Set(
					"Access-Control-Allow-Headers",
					"Origin, Content-Type, Authorization, Accept, X-Requested-With, X-Access-Token",
				)
				w.Header().Set(
					"Access-Control-Expose-Headers",
					"Content-Length",
				)
				w.Header().Set("Access-Control-Max-Age", "43200") // 12 hours
			}

			// Кэширование и другие заголовки для React приложения
			w.Header().Set(
				"Cache-Control",
				"no-store, no-cache, must-revalidate, private",
			)
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")

			// Если это OPTIONS запрос, возвращаем пустой ответ
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Передаем управление следующему обработчику
			next.ServeHTTP(w, r)
		},
	)
}

func tokenToHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем наличие X-Access-Token заголовка
		token := r.Header.Get("X-Access-Token")
		if token != "" {
			// Создаем cookie с токеном
			cookie := &http.Cookie{
				Name:     "access_token",
				Value:    token,
				MaxAge:   30000,
				Path:     "/",
				HttpOnly: true,
				Secure:   false, // Для локальной разработки
				SameSite: http.SameSiteLaxMode,
			}

			// Добавляем cookie в текущий запрос
			r.AddCookie(cookie)

			// Также устанавливаем cookie в ответ для будущих запросов
			http.SetCookie(w, cookie)
		}
		next.ServeHTTP(w, r)
	})
}

//func authenticateMiddleware(
//	next http.Handler,
//	authService authService.AuthService,
//) http.Handler {
//	return http.HandlerFunc(
//		func(w http.ResponseWriter, r *http.Request) {
//			tokenString, err := r.Cookie("token")
//			if err != nil {
//				fmt.Println("Token missing in cookie")
//				http.Redirect(w, r, "/auth", http.StatusUnauthorized)
//				return
//			}
//
//			// Verify the token
//			token, err := authService.VerifyToken(tokenString.Value)
//			if err != nil {
//				fmt.Printf("Token verification failed: %v\\n", err)
//				http.Redirect(w, r, "/auth", http.StatusUnauthorized)
//				return
//			}
//
//			// Print information about the verified token
//			fmt.Printf(
//				"Token verified successfully. Claims: %+v\\n",
//				token.Claims,
//			)
//
//			// Continue with the next middleware or route handler
//			next.ServeHTTP(w, r)
//		},
//	)
//}
