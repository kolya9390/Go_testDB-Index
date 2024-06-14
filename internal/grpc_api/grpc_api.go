package grpcapi

import (
	"context"
	"db-index/config"
	"db-index/internal/api"
	"db-index/internal/app/rates"
	"fmt"
	"net"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCAPIService struct {
	app		*rates.App
	logger	*zap.Logger
	tr trace.Tracer

	api.UnimplementedAPIServiceGarantexServer
}



func NewGRPCAPIService(app	*rates.App,
	logger	*zap.Logger,tr trace.Tracer) *GRPCAPIService{
	return &GRPCAPIService{
		app: app,
		logger: logger,
		tr: tr,

		UnimplementedAPIServiceGarantexServer: api.UnimplementedAPIServiceGarantexServer{},

	}
}


func RunGrpcServer(ctx context.Context, logger *zap.Logger, app *rates.App, config *config.Config,tr trace.Tracer) error {
	logger.Info("Starting grpc server")
	grpcAPIService := NewGRPCAPIService(app,logger,tr)
	grpcServer := grpc.NewServer()
	ctx,span :=tr.Start(ctx,"RunGRPCServer")
	defer span.End()
	reflection.Register(grpcServer)
	api.RegisterAPIServiceGarantexServer(grpcServer,grpcAPIService)

	grpcAdres := fmt.Sprintf(":%s",config.GRPCPort)
	listen, err := net.Listen("tcp", grpcAdres)
	if err != nil {
		return fmt.Errorf("GRPC server can't listen requests %s",err.Error())
	}

	logger.Info("GRPC Server",zap.String("server listen requests",listen.Addr().String()))

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

func (g *GRPCAPIService) GetRates(ctx context.Context, empty *emptypb.Empty) (*api.GetRatesResponse, error) {

	_,span := g.tr.Start(ctx,"GetRates")
	defer span.End()
	resp,err := g.app.Get(ctx)
	if err != nil {
		g.logger.Error("GRPS GetRates",zap.String("App Resp",err.Error()))
		return nil,err
	}

	return &api.GetRatesResponse{
		Timestamp: int64(resp.Timestamp),
		AsksPrice: resp.AskPrice,
		BidsPrice: resp.BidPrice,
	},nil
}

func (g *GRPCAPIService) HealthCheck(ctx context.Context, empty *emptypb.Empty) (*api.HealthCheckResponse, error) {
	return &api.HealthCheckResponse{Status: "active"},nil
}