package internal

import (
	"net/http"
	"s3-saver/internal/config"
	"s3-saver/internal/http-server/handlers"
	"s3-saver/internal/http-server/middleware/logger"
	logUtil "s3-saver/internal/lib/logger/slog"
	"s3-saver/internal/service"

	_ "s3-saver/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func Main() {
	appConfig := config.MustLoadAppConfig()
	log := config.MustConfigureSlogLogger(appConfig.Env)

	log.Info("starting app")

	router := chi.NewRouter()

	// CORS middleware для Swagger UI
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With"},
		ExposedHeaders:   []string{"Link", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	//router.Use(middleware.URLFormat)

	s3Service := service.NewS3Service(appConfig)
	historyService := service.NewInMemoryHistory(appConfig.MaxCachedVideosUrl)
	videoHandler := handlers.NewVideoHandler(s3Service, log, appConfig, historyService)

	// Swagger документация
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // <--- Оставьте только это
	))

	// API routes
	router.Post("/upload/video", videoHandler.Upload)
	router.Get("/recent", videoHandler.GetRecentList)

	server := &http.Server{
		Addr:         appConfig.HttpConfig.Host + ":" + appConfig.HttpConfig.Port,
		Handler:      router,
		ReadTimeout:  appConfig.HttpConfig.Timeout,
		WriteTimeout: appConfig.HttpConfig.Timeout,
		IdleTimeout:  appConfig.HttpConfig.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("server error", logUtil.Err(err))
	}

	log.Warn("stopping app")
}
