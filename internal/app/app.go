package app

import (
	"go-arch/internal/handler/cron"
	handlehttp "go-arch/internal/handler/http"
	"go-arch/internal/handler/nsq"
	"go-arch/internal/infra/configuration"
	"go-arch/internal/infra/db"
	infraredis "go-arch/internal/infra/redis"
	"go-arch/internal/repository"
	"go-arch/internal/service"
	"go-arch/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type Application struct {
	Config      *configuration.Config
	Logger      zerolog.Logger
	DBRegistry  *db.Registry
	RedisReg    *infraredis.Registry
	Router      *mux.Router
	CronHandler *cron.Handler
	NSQHandler  *nsq.Handler
}

func New(cfg *configuration.Config, logger zerolog.Logger, dbReg *db.Registry, redisReg *infraredis.Registry) *Application {
	app := &Application{
		Config:     cfg,
		Logger:     logger,
		DBRegistry: dbReg,
		RedisReg:   redisReg,
	}

	app.wireHTTP()
	app.CronHandler = cron.NewHandler(logger)
	app.NSQHandler = nsq.NewHandler(logger)

	return app
}

func (a *Application) wireHTTP() {
	// Repositories
	// Uncomment when database is available:
	// mainDB, _ := a.DBRegistry.Get("main")
	// exampleRepo := repository.NewExampleRepository(mainDB)
	var exampleRepo repository.ExampleRepository // nil placeholder

	// Services
	exampleService := service.NewExampleService(exampleRepo, a.Logger)

	// Usecases
	exampleUsecase := usecase.NewExampleUsecase(exampleService, a.Logger)

	// Handlers
	handlers := handlehttp.Handlers{
		Health:  handlehttp.NewHealthHandler(),
		Example: handlehttp.NewExampleHandler(exampleUsecase, a.Logger),
	}

	a.Router = handlehttp.NewRouter(a.Logger, handlers)
}

func (a *Application) Close() error {
	var firstErr error

	if a.DBRegistry != nil {
		if err := a.DBRegistry.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if a.RedisReg != nil {
		if err := a.RedisReg.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
