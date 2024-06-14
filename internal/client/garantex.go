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

type Client struct {
	garantex http.Client
	Logger   *zap.Logger
	tr trace.Tracer
}

func NewClient(Logger *zap.Logger,tr trace.Tracer) *Client {
	return &Client{
		garantex: http.Client{},
		Logger:   Logger,
		tr: tr,
	}
}

func (c *Client) GetRates(market string,ctx context.Context) (*types.DataMD, error) {
	ctx,span := c.tr.Start(ctx,"garantex GetRates")
	defer span.End()
	url := fmt.Sprintf("https://garantex.org/api/v2/depth?market=%s", market)

	req, err := http.NewRequestWithContext(ctx,"GET",url,nil)

	if err != nil {
		c.Logger.Error("Do Conect", zap.String("err", err.Error()))
		return nil, err
	}

	res, err := c.garantex.Do(req)

	defer func() {
		err := res.Body.Close()
		if err != nil {
			c.Logger.Error("Bad close", zap.String("err", err.Error()))
		}
	}()

	if res.StatusCode != http.StatusOK {
		c.Logger.Error("Unexpected HTTP status code", zap.Int("status_code", res.StatusCode))
		return nil, fmt.Errorf("unexpected HTTP status code: %d", res.StatusCode)
	}

	bodyText, err := io.ReadAll(res.Body)
	if err != nil {
		c.Logger.Error("Read", zap.String("err", err.Error()))
	}

	var respMarketData types.DataMD

	err = json.Unmarshal(bodyText, &respMarketData)
	if err != nil {
		c.Logger.Error("Unmarshal", zap.String("err", err.Error()))
		return nil, err
	}

	c.Logger.Info("sucssesful req api garantex", zap.String("url", url),zap.String("resp Asc Price",respMarketData.Asks[0].Price))

	return &respMarketData, nil

}
