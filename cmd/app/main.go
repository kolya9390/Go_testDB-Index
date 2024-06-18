package main

import (
	"context"
	"db-index/config"
	"db-index/internal/app/rates"
	"db-index/internal/client"
	grpcapi "db-index/internal/grpc_api"
	"db-index/internal/storage/postgres"
	"db-index/pkg/logster"
	"flag"
	"net/http"
	"os"
	"os/signal"

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


	if err != nil {
		logger.Fatal("Env",zap.String("err",err.Error()),zap.String("Path",configPath))
	}

	logger.Info("service starting")

	// Init Storage and Postgres

	storage := postgres.NewStorage(log)
	err = storage.InitDB(ctx,apiConfig.Postgres)
	if err != nil{
		logger.Error("Storage",zap.String("error",err.Error()))
	}

	defer func() {
		storage.Stop()
		stop()
	}()

	garantex := http.Client{}
	sevice := client.NewClient(&garantex,log.For(ctx))
	app := rates.NewApp(log,storage,sevice)

	// init Grpc Server
	err = grpcapi.RunGrpcServer(ctx,log,app,&apiConfig,"otlp")
	if err != nil{
		logger.Error("Grpc server",zap.String("error",err.Error()))
	}
}
