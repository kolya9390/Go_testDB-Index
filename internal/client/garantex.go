package client

import (
	"db-index/internal/client/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type GarantexService interface {
	GetRates(market string) (*types.DataMD, error)
}

type Client struct {
	garantex http.Client
	Logger *zap.Logger
}

func NewClient(Logger *zap.Logger ) *Client {
	return &Client{
		garantex: http.Client{},
		Logger: Logger,
	}
}

func (c *Client) GetRates(market string) (*types.DataMD, error) {
	url := fmt.Sprintf("https://garantex.org/api/v2/depth?market={%s}",market)

	resp, err := http.Get(url)

	c.Logger.Error("Do Conect",zap.String("err",err.Error()))
	if err != nil {
		return nil,err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			c.Logger.Error("Bad close",zap.String("err",err.Error()))
		}
		}()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error("Read",zap.String("err",err.Error()))
	}

	var respMarketData types.DataMD

	err = json.Unmarshal(bodyText,&respMarketData)
	if err != nil {
		return nil,err
	}

	c.Logger.Info("sucssesful",zap.String("url",url),zap.String("Response Body",string(bodyText)))

	return &respMarketData,nil

}