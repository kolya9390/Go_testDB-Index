package client

import (
	"context"
	"db-index/internal/client/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type GarantexService interface {
	GetRates(market string,ctx context.Context) (*types.DataMD, error)
}

type Client struct {
	garantex http.Client
	Logger   *zap.Logger
}

func NewClient(Logger *zap.Logger) *Client {
	return &Client{
		garantex: http.Client{},
		Logger:   Logger,
	}
}

func (c *Client) GetRates(market string,ctx context.Context) (*types.DataMD, error) {
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
