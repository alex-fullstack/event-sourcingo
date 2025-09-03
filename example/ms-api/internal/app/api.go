package app

import (
	"api/internal/domain/usecase"
	"api/internal/endpoints/api"
	"api/internal/endpoints/api/frontend/middleware"
	apiV1 "api/internal/endpoints/api/frontend/v1"
	"api/internal/endpoints/consumers"
	elastic "api/internal/infrastructure/elasticsearch"
	"api/internal/infrastructure/grpc"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-chi/chi/v5"
	kafkaGo "github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context) error {
	logger := slog.Default()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	elasticCfg := elasticsearch.Config{
		Addresses:              []string{viper.GetString("elasticsearch_address")},
		Username:               viper.GetString("elasticsearch_username"),
		Password:               viper.GetString("elasticsearch_password"),
		CertificateFingerprint: viper.GetString("elasticsearch_certificate_fingerprint"),
	}
	el, err := elastic.NewClient(elasticCfg)
	if err != nil {
		return err
	}
	userService := usecase.NewUserService(el)

	userKafkaReaderCfg := kafkaGo.ReaderConfig{
		Brokers:   []string{viper.GetString("kafka_address")},
		Topic:     viper.GetString("kafka_user_consumer_topic"),
		Partition: 0,
		GroupID:   viper.GetString("kafka_user_consumer_group"),
	}

	userConsumer := consumers.NewUserConsumer(ctx, userService, userKafkaReaderCfg, logger)

	policyService := usecase.NewPolicyService(el)

	policyKafkaReaderCfg := kafkaGo.ReaderConfig{
		Brokers:   []string{viper.GetString("kafka_address")},
		Topic:     viper.GetString("kafka_policy_consumer_topic"),
		Partition: 0,
		GroupID:   viper.GetString("kafka_policy_consumer_group"),
	}

	policyConsumer := consumers.NewPolicyConsumer(ctx, policyService, policyKafkaReaderCfg, logger)

	policyRepository, err := grpc.NewClient(viper.GetString("policy_backend_addr"))
	if err != nil {
		return err
	}
	defer func() {
		err = policyRepository.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}()
	frontendAPIService := usecase.NewFrontendAPIService(el, policyRepository)
	frontendAPI := api.NewFrontendAPI(
		ctx, viper.GetString("frontend_addr"),
		func() http.Handler {
			router := chi.NewRouter()
			router.Mount("/api/v1/", apiV1.New(frontendAPIService, apiV1.NewConverter()))
			return middleware.NewLogger(router)
		}(),
		logger,
	)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return userConsumer.GracefulStart(ctx, nil)
	})
	eg.Go(func() error {
		return policyConsumer.GracefulStart(ctx, nil)
	})
	eg.Go(func() error {
		return frontendAPI.GracefulStart(ctx, nil)
	})
	return eg.Wait()
}
