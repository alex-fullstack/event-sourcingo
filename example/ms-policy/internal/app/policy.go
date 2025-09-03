package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"policy/internal/domain/entities"
	"policy/internal/domain/usecase"
	"policy/internal/endpoints/api"
	"policy/internal/endpoints/api/backend"
	"policy/internal/endpoints/consumers"
	"policy/internal/infrastructure/mongodb"
	"strconv"
	"syscall"

	coreEntities "github.com/alex-fullstack/event-sourcingo/domain/entities"
	"github.com/alex-fullstack/event-sourcingo/domain/usecases/services"
	coreConsumers "github.com/alex-fullstack/event-sourcingo/endpoints/consumers"
	"github.com/alex-fullstack/event-sourcingo/infrastructure/kafka"
	"github.com/alex-fullstack/event-sourcingo/infrastructure/postgresql"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	kafkaGo "github.com/segmentio/kafka-go"
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
	repository, err := mongodb.NewMongoDB(ctx, mongoCfg)
	if err != nil {
		return err
	}
	defer func() {
		err = repository.Close(ctx)
		if err != nil {
			slog.Error(err.Error())
		}
	}()

	handler := services.NewCommandHandler(db, repository)
	userService := usecase.NewUserService(handler, repository)

	kafkaCfg := kafkaGo.ReaderConfig{
		Brokers:   []string{viper.GetString("kafka_address")},
		Topic:     viper.GetString("kafka_user_consumer_topic"),
		Partition: 0,
		GroupID:   viper.GetString("kafka_user_consumer_group"),
	}

	msUserConsumer := consumers.NewMSUserConsumer(ctx, userService, kafkaCfg, logger)

	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}

	defer conn.Release()

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

	consumer := coreConsumers.NewTransactionConsumer(
		ctx,
		"es.transaction-handled",
		conn,
		services.NewTransactionHandler(db, services.NewEventHandler(publisher), logger),
		func(id uuid.UUID) coreEntities.AggregateProvider { return entities.NewPolicy(id) },
	)

	backendAPIService := usecase.NewBackendAPIService(handler, repository)

	backendAPIServer := backend.New(backend.NewConverter(), backendAPIService)
	backendAPI := api.NewBackendAPI(ctx, backendAPIServer, viper.GetString("backend_addr"), logger)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return msUserConsumer.GracefulStart(ctx, nil)
	})
	eg.Go(func() error {
		return consumer.GracefulStart(ctx, nil)
	})
	eg.Go(func() error {
		return backendAPI.GracefulStart(ctx, nil)
	})
	return eg.Wait()
}
