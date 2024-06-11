package server

import (
	"context"
	"db-index/internal/api"
	"db-index/internal/app/rates"

	"go.uber.org/zap"
)

type GRPCServer struct {
	app		*rates.App
	logger	*zap.Logger

	api.UnimplementedAPIServiceGarantexServer
}


func NewGRPCServer(app	*rates.App,
	logger	*zap.Logger) *GRPCServer{
	return &GRPCServer{
		app: app,
		logger: logger,

		UnimplementedAPIServiceGarantexServer: api.UnimplementedAPIServiceGarantexServer{},

	}
}

func (g *GRPCServer) GetRates(ctx context.Context) (*api.GetRatesResponse, error) {

	resp,err := g.app.Get()
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

func (g *GRPCServer) HealthCheck(ctx context.Context) (*api.HealthCheckResponse, error) {
	return &api.HealthCheckResponse{Status: "active"},nil
}