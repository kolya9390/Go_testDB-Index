package main

import (
	"context"
	"db-index/config"
	"db-index/internal/app/rates"
	"db-index/internal/client"
	grpcapi "db-index/internal/grpc_api"
	"db-index/internal/storage/postgres"
	"db-index/pkg/tracing"
	"flag"
	"os"
	"os/signal"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/contrib/otelfiber"
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

	tr , err := tracing.InitTracer("http://localhost:14268/api/traces", "Rates Service")

	if err != nil {
		logger.Fatal("failed to initialize tracer", zap.Error(err))
	}
	f := fiber.New()
	f.Use(otelfiber.Middleware())


	logger.Info("service starting")

	// Init Storage and Postgres

	storage := postgres.NewStorage(logger,tr)
	err = storage.InitDB(apiConfig.Postgres)
	if err != nil{
		logger.Error("Storage",zap.String("error",err.Error()))
	}


	sevice := client.NewClient(logger,tr)
	app := rates.NewApp(logger,storage,sevice)


	err = grpcapi.RunGrpcServer(ctx,logger,app,&apiConfig,tr)
	if err != nil{
		logger.Error("Grpc server",zap.String("error",err.Error()))
	}

}
