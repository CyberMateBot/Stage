package main

import (
	"context"
	"log/slog"

	"github.com/twelvepills-936/tgapp-/internal/bot"
	"github.com/twelvepills-936/tgapp-/internal/repository"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
	"github.com/twelvepills-936/tgapp-/internal/service"
	"github.com/twelvepills-936/tgapp-/internal/usecase"
	api "github.com/twelvepills-936/tgapp-/pkg/api"
	"github.com/twelvepills-936/tgapp-/pkg/app"
	"github.com/twelvepills-936/tgapp-/pkg/applinks"
	"github.com/twelvepills-936/tgapp-/pkg/config"
	"github.com/twelvepills-936/tgapp-/pkg/cors"
	"github.com/twelvepills-936/tgapp-/pkg/health"
	"github.com/twelvepills-936/tgapp-/pkg/logger"
	"github.com/twelvepills-936/tgapp-/pkg/swagger"
)

func main() {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	cfg := app.LoadConfigFromEnv()

	application, err := app.New(ctx, cfg)
	if err != nil {
		panic(err)
	}

	addConfig := config.LoadConfig()

	tgBot, err := bot.New()
	if err != nil {
		slog.WarnContext(ctx, "failed to init bot", logger.ErrorAttr(err))
	} else if tgBot != nil {
		go tgBot.StartPolling(ctx)
	}

	pool, err := repository.NewPostgres(ctx, repoModels.ConfigPostgres(addConfig.Postgres))
	if err != nil {
		slog.ErrorContext(ctx, "failed to init postgres", logger.ErrorAttr(err))
		return
	}
	defer pool.Close()
	repo := repository.NewRepository(pool)

	// Create single instances of usecase and service
	uc := usecase.NewUseCase(repo)
	svc := service.NewService(uc)

	// Register gRPC services BEFORE starting the server
	api.RegisterUsersServer(application.GrpcServer, svc)
	api.RegisterCyberMateServer(application.GrpcServer, svc)

	err = application.Init(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to init app", logger.ErrorAttr(err))
		return
	}
	err = api.RegisterUsersHandler(ctx, application.ServeMux, application.GrpcConn)
	if err != nil {
		slog.ErrorContext(ctx, "failed to register users handler", logger.ErrorAttr(err))
		return
	}

	err = api.RegisterCyberMateHandler(ctx, application.ServeMux, application.GrpcConn)
	if err != nil {
		slog.ErrorContext(ctx, "failed to register cybermate handler", logger.ErrorAttr(err))
		return
	}

	httpHandler := cors.Wrap(
		health.Wrap(
			applinks.Wrap(
				swagger.Wrap(application.ServeMux, addConfig.App.SwaggerEnabled),
				addConfig.App,
			),
		),
		addConfig.CORS,
	)
	application.SetHTTPHandler(httpHandler)

	err = application.Run(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to run app", logger.ErrorAttr(err))
		return
	}
}
