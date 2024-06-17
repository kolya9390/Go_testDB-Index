package main

import (
	"context"
	"db-index/config"
	"db-index/internal/app/rates"
	"db-index/internal/client"
	grpcapi "db-index/internal/grpc_api"
	"db-index/internal/storage/postgres"
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

	// Init Logger 

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	if err != nil {
		logger.Fatal("Env",zap.String("err",err.Error()),zap.String("Path",configPath))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Init Tracing

	
	if err != nil {
		logger.Fatal("failed to initialize tracer", zap.Error(err))
	}

	logger.Info("service starting")

	// Init Storage and Postgres

	storage := postgres.NewStorage(logger)
	err = storage.InitDB(apiConfig.Postgres)
	if err != nil{
		logger.Error("Storage",zap.String("error",err.Error()))
	}

	garantex := http.Client{}
	sevice := client.NewClient(&garantex,logger)
	app := rates.NewApp(logger,storage,sevice)


	err = grpcapi.RunGrpcServer(ctx,logger,app,&apiConfig)
	if err != nil{
		logger.Error("Grpc server",zap.String("error",err.Error()))
	}

}
