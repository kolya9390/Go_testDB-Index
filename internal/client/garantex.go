package client

import (
	"context"
	"db-index/internal/client/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type GarantexService interface {
	GetRates(market string,ctx context.Context) (*types.DataMD, error)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	garantex HTTPClient
	logger   *zap.Logger
	tr trace.Tracer
}

func NewClient(garantex HTTPClient,logger *zap.Logger) *Client {
	return &Client{
		garantex: garantex,
		logger:   logger,
		tr: trace.NewNoopTracerProvider().Tracer("Garantex GetRates"),
	}
}

func (c *Client) GetRates(market string,ctx context.Context) (*types.DataMD, error) {
	url := fmt.Sprintf("https://garantex.org/api/v2/depth?market=%s", market)

	_,span := c.tr.Start(ctx,"garantex GetRates",trace.WithAttributes())
	defer span.End()

	req, err := http.NewRequestWithContext(ctx,"GET",url,nil)

	if err != nil {
		c.logger.Error("Do Conect", zap.String("err", err.Error()))
		return nil, err
	}

	res, err := c.garantex.Do(req)

	defer func() {
		err := res.Body.Close()
		if err != nil {
			c.logger.Error("Bad close", zap.String("err", err.Error()))
		}
	}()

	if res.StatusCode != http.StatusOK {
		c.logger.Error("Unexpected HTTP status code", zap.Int("status_code", res.StatusCode))
		return nil, fmt.Errorf("unexpected HTTP status code: %d", res.StatusCode)
	}

	bodyText, err := io.ReadAll(res.Body)
	if err != nil {
		c.logger.Error("Read", zap.String("err", err.Error()))
	}

	var respMarketData types.DataMD

	err = json.Unmarshal(bodyText, &respMarketData)
	if err != nil {
		c.logger.Error("Unmarshal", zap.String("err", err.Error()))
		return nil, err
	}

	c.logger.Info("sucssesful req api garantex", zap.String("url", url),zap.String("resp Asc Price",respMarketData.Asks[0].Price))

	return &respMarketData, nil

}
