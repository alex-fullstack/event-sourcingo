package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"user/internal/domain/entities"
	"user/internal/domain/usecase"
	"user/internal/endpoints/api"
	"user/internal/endpoints/api/backend"
	"user/internal/endpoints/api/frontend/middleware"
	apiV1 "user/internal/endpoints/api/frontend/v1"
	"user/internal/infrastructure/mongodb/auth"

	coreEntities "github.com/alex-fullstack/event-sourcingo/domain/entities"
	"github.com/alex-fullstack/event-sourcingo/domain/usecases/services"
	"github.com/alex-fullstack/event-sourcingo/endpoints/consumers"
	"github.com/alex-fullstack/event-sourcingo/infrastructure/kafka"
	"github.com/alex-fullstack/event-sourcingo/infrastructure/postgresql"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context) error {
	logger := slog.Default()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	postgresCfg, err := pgxpool.ParseConfig(viper.GetString("postgres_url"))
	if err != nil {
		return err
	}
	maxConn, err := strconv.ParseInt(viper.GetString("postgres_max_conn"), 10, 32)
	if err != nil {
		return err
	}
	postgresCfg.MaxConns = int32(maxConn)
	db, err := postgresql.NewPostgresDB(ctx, postgresCfg)
	if err != nil {
		return err
	}
	defer db.Close()

	mongoCfg := options.Client().ApplyURI(viper.GetString("mongo_url"))
	repository, err := auth.NewMongoDB(ctx, mongoCfg)
	if err != nil {
		return err
	}
	defer func() {
		err = repository.Close(ctx)
		if err != nil {
			logger.Error(err.Error())
		}
	}()

	frontendAPIService := usecase.NewFrontendAPIService(
		services.NewCommandHandler(db, repository),
		repository,
	)
	frontendAPI := api.NewFrontendAPI(ctx, viper.GetString("frontend_addr"), func() http.Handler {
		router := chi.NewRouter()
		router.Mount("/api/v1/", apiV1.New(frontendAPIService, apiV1.NewConverter()))
		return middleware.NewLogger(router)
	}(),
		logger,
	)

	backendAPIService := usecase.NewBackendAPIService(repository)
	backendAPIServer := backend.New(backend.NewConverter(), backendAPIService)
	backendAPI := api.NewBackendAPI(ctx, backendAPIServer, viper.GetString("backend_addr"), logger)

	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}

	publisher := kafka.NewWriter(
		&kafka.Config{
			Address: viper.GetString("kafka_address"),
			Topic:   viper.GetString("kafka_topic"),
		},
	)
	defer func() {
		err = publisher.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}()
	handler := services.NewEventHandler(publisher)
	consumer := consumers.NewTransactionConsumer(
		ctx,
		"es.transaction-handled",
		conn,
		services.NewTransactionHandler(db, handler, logger),
		func(id uuid.UUID) coreEntities.AggregateProvider { return entities.NewUser(id) },
	)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return frontendAPI.GracefulStart(ctx, nil)
	})
	eg.Go(func() error {
		return backendAPI.GracefulStart(ctx, nil)
	})
	eg.Go(func() error {
		return consumer.GracefulStart(ctx, nil)
	})
	return eg.Wait()
}
