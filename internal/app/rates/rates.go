package rates

import (
	"context"
	"db-index/internal/client"
	"db-index/internal/domain"
	"db-index/internal/storage/postgres"
	"db-index/pkg/logster"
	"time"
)

type ImplementationApp interface {
	Get(ctx context.Context) (domain.Rates,error)
}

type App struct {
	logger logster.Factory
	storage postgres.Storaged
	servises client.GarantexService
}

func NewApp(logger logster.Factory,
	storage postgres.Storaged,
	servises client.GarantexService) *App {

		return &App{
			logger: logger,
			storage: storage,
			servises: servises,
		}
	}

func (a *App) Get(ctx context.Context) (domain.Rates,error) {

	respRates ,err := a.servises.GetRates("usdtrub",ctx)
	if err != nil {
		return domain.Rates{},err
	}

	rates := domain.Rates{
		Timestamp: time.Duration(respRates.Timestamp),	
		AskPrice: respRates.Asks[0].Price,
		BidPrice: respRates.Bids[0].Price,
	}

	err = a.storage.AddRates(ctx,rates)

	if err != nil {
		return domain.Rates{},err
	}

	return rates,nil
}