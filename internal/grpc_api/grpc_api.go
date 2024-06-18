package grpcapi

import (
	"context"
	"db-index/config"
	"db-index/internal/api"
	"db-index/internal/app/rates"
	"db-index/pkg/logster"
	"encoding/json"
	"fmt"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCGetRatesService interface {
	GetRates(ctx context.Context, empty *emptypb.Empty) (*api.GetRatesResponse, error)
	HealthCheck(ctx context.Context, empty *emptypb.Empty) (*api.HealthCheckResponse, error)
}

type GRPCAPIServer struct {
	implementationApp rates.ImplementationApp
	logger            logster.Factory

	api.UnimplementedAPIServiceGarantexServer
}

func NewGRPCAPIService(implementationApp rates.ImplementationApp,
	logger logster.Factory) *GRPCAPIServer {
	return &GRPCAPIServer{
		implementationApp: implementationApp,
		logger:            logger,

		UnimplementedAPIServiceGarantexServer: api.UnimplementedAPIServiceGarantexServer{},
	}
}

func RunGrpcServer(ctx context.Context,
	logger logster.Factory,
	implementationApp rates.ImplementationApp,
	config *config.Config,
	tracerProvider trace.TracerProvider) error {
	logger.Bg().Info("Starting", zap.String("address", config.GRPCPort), zap.String("type", "gRPC"))

	grpcAPIService := NewGRPCAPIService(implementationApp, logger)

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(tracerProvider))),
	)

	reflection.Register(grpcServer)

	api.RegisterAPIServiceGarantexServer(grpcServer, grpcAPIService)

	grpcAdres := fmt.Sprintf(":%s", config.GRPCPort)

	listen, err := net.Listen("tcp", grpcAdres)

	if err != nil {
		logger.Bg().Fatal("Unable to create http listener", zap.Error(err))
		return fmt.Errorf("GRPC server can't listen requests %s", err.Error())
	}

	errListen := make(chan error, 1)
	go func() {
		errListen <- grpcServer.Serve(listen)
	}()

	select {
	case <-ctx.Done():
		grpcServer.GracefulStop()
		return nil
	case err = <-errListen:
		return fmt.Errorf("can't run grpc server: %w", err)
	}
}

func (g *GRPCAPIServer) GetRates(ctx context.Context, empty *emptypb.Empty) (*api.GetRatesResponse, error) {
	g.logger.For(ctx).Info("GetRates GRPC method")
	resp, err := g.implementationApp.Get(ctx)
	if err != nil {
		g.logger.For(ctx).Error("Failed to get rates", zap.String("App Resp", err.Error()))
		return nil, err
	}

	g.logger.For(ctx).Info(
		"Get rates successful",
		zap.String("Resp Rates", toJSON(resp)),
	)

	return &api.GetRatesResponse{
		Timestamp: int64(resp.Timestamp),
		AsksPrice: resp.AskPrice,
		BidsPrice: resp.BidPrice,
	}, nil
}

func (g *GRPCAPIServer) HealthCheck(ctx context.Context, empty *emptypb.Empty) (*api.HealthCheckResponse, error) {
	g.logger.For(ctx).Info("HealthCheck successful",zap.String("Status","active"))
	return &api.HealthCheckResponse{Status: "active"}, nil
}

func toJSON(v any) string {
	str, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(str)
}
