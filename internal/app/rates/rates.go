package rates

import (
	"db-index/internal/client"
	"db-index/internal/domain"
	"db-index/internal/storage/postgres"
	"time"

	"go.uber.org/zap"
)

type ImplementationApp interface {
	Get() (domain.Rates,error)
}

type App struct {
	logger *zap.Logger
	storage postgres.Soraged
	servises client.GarantexService
}

func NewApp(logger *zap.Logger,
	storage postgres.Soraged,
	servises client.GarantexService) *App {

		return &App{
			logger: logger,
			storage: storage,
			servises: servises,
		}
	}

func (a *App) Get() (domain.Rates,error) {

	respRates ,err := a.servises.GetRates("usdrub")
	if err != nil {
		return domain.Rates{},err
	}

	rates := domain.Rates{
		Timestamp: time.Duration(respRates.Timestamp),	
		AskPrice: respRates.Asks[0].Price,
		BidPrice: respRates.Bids[0].Price,
	}

	err = a.storage.Add(rates)

	if err != nil {
		return domain.Rates{},err
	}

	return rates,nil
}