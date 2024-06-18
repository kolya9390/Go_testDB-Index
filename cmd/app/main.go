package main

import (
	"context"
	"db-index/config"
	"db-index/internal/app/rates"
	"db-index/internal/client"
	grpcapi "db-index/internal/grpc_api"
	"db-index/internal/storage/postgres"
	"db-index/pkg/logster"
	"db-index/pkg/tracing"
	"flag"
	"net/http"
	"os"
	"os/signal"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", ".env", "path to config file")
}

func main() {
	flag.Parse()
	
	apiConfig,err := config.LoadConfigFromEnv(configPath)

	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	// Init Logger 

	logger, _ := zap.NewDevelopment()
	zapLogger := logger.With(zap.String("service", "Rates"))
	log := logster.NewFactory(zapLogger)

	// Init Tracing

	tracerProvider := tracing.InitOTEL("GRPC for api Garantex Get Rates", "otlp", log)


	if err != nil {
		logger.Fatal("Env",zap.String("err",err.Error()),zap.String("Path",configPath))
	}

	logger.Info("service starting")

	// Init Storage and Postgres

	storage := postgres.NewStorage(log,tracerProvider.Tracer("Postgres DB"))

	err = storage.InitDB(ctx,apiConfig.Postgres)
	if err != nil{
		logger.Error("Storage",zap.String("error",err.Error()))
	}

	defer func() {
		storage.Stop()
		stop()
	}()


	// Init HTTP Client

	garantex := http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithTracerProvider(tracerProvider),
		),
	}
	sevice := client.NewClient(&garantex,log)
	app := rates.NewApp(log,storage,sevice)

	// init Grpc Server
	err = grpcapi.RunGrpcServer(ctx,log,app,&apiConfig,tracerProvider)
	if err != nil{
		logger.Error("Grpc server",zap.String("error",err.Error()))
	}
}
