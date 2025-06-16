package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"policy/internal/domain/entities"
	"policy/internal/domain/usecase"
	"policy/internal/endpoints/cli"
	"policy/internal/infrastructure/grpc"
	"policy/internal/infrastructure/mongodb"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	coreEntities "gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/entities"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/usecases/services"
	coreConsumers "gitverse.ru/aleksandr-bebyakov/event-sourcingo/endpoints/consumers"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/infrastructure/kafka"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/infrastructure/postgresql"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/sync/errgroup"
)

const pauseTimeout = 100 * time.Millisecond

func Execute(ctx context.Context, cmdName string, args ...string) error {
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

	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}

	defer conn.Release()

	publisher := kafka.NewWriter(
		&kafka.Config{Address: viper.GetString("kafka_address"), Topic: viper.GetString("kafka_topic")},
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
		services.NewTransactionHandler(db, services.NewEventHandler(publisher)),
		func(id uuid.UUID) coreEntities.AggregateProvider { return entities.NewPolicy(id) },
	)
	userRepository, err := grpc.NewClient(viper.GetString("user_backend_addr"))
	if err != nil {
		return err
	}
	defer func() {
		err = userRepository.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}()
	adminService := usecase.NewAdminService(handler, repository, userRepository)
	adminCli := cli.New(adminService, cli.NewConverter())
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return consumer.GracefulStart(ctx, nil)
	})
	eg.Go(func() error {
		time.Sleep(pauseTimeout)
		return adminCli.RunCmd(ctx, cmdName, args...)
	})

	return eg.Wait()
}
