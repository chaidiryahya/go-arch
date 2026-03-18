package http

import (
	"go-arch/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type Handlers struct {
	Health  *HealthHandler
	Example *ExampleHandler
}

type ExampleHandler struct {
	usecase usecase.ExampleUsecase
	logger  zerolog.Logger
}

func NewExampleHandler(uc usecase.ExampleUsecase, logger zerolog.Logger) *ExampleHandler {
	return &ExampleHandler{
		usecase: uc,
		logger:  logger,
	}
}

func NewRouter(logger zerolog.Logger, handlers Handlers) *mux.Router {
	r := mux.NewRouter()

	// Global middleware
	r.Use(RequestIDMiddleware)
	r.Use(LoggingMiddleware(logger))
	r.Use(RecoveryMiddleware(logger))

	// Health check
	r.HandleFunc("/health", handlers.Health.Health).Methods("GET")

	// Example routes
	// r.HandleFunc("/examples/{id}", handlers.Example.GetByID).Methods("GET")

	return r
}
