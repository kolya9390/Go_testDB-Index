package client

import (
	"context"
	"db-index/internal/client/types"
	"db-index/pkg/logster"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestClient_GetRates(t *testing.T) {

	mockRespData := types.DataMD{
		Timestamp: 1718574098,
		Asks: []types.Order{
			{
				Price:  "62.18",
				Volume: "10",
				Amount: "921.8",
				Factor: "1",
				Type:   "ask",
			},
			{
				Price:  "92.19",
				Volume: "5",
				Amount: "460.95",
				Factor: "1",
				Type:   "ask",
			},
		},
		Bids: []types.Order{
			{
				Price:  "62.06",
				Volume: "8",
				Amount: "736.48",
				Factor: "1",
				Type:   "bid",
			},
			{
				Price:  "62.05",
				Volume: "15",
				Amount: "1380.75",
				Factor: "1",
				Type:   "bid",
			},
		},
	}
		mockClient := &MockHTTPClient{
			Handler: func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "https://garantex.org/api/v2/depth?market=usdtrub", req.URL.String())
	
				respBody, _ := json.Marshal(mockRespData)
				resp := httptest.NewRecorder()
				resp.WriteHeader(http.StatusOK)
				resp.Write(respBody)
	
				return resp.Result(), nil
			},
		}


	loggerConfig := zap.NewDevelopmentConfig()	
	logger, _ := loggerConfig.Build()
	zapLogger := logger.With(zap.String("service", "TESTS gatantex Client api"))
	log := logster.NewFactory(zapLogger)
	defer logger.Sync()

	client :=  &Client{
		garantex: mockClient,
		logger: log,
	}

	ctx := context.Background()
	data, err := client.GetRates("usdtrub", ctx)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, int64(1718574098), data.Timestamp)
	assert.Equal(t, "62.18", data.Asks[0].Price)
	assert.Equal(t, "62.06", data.Bids[0].Price)
}

type MockHTTPClient struct {
	Handler func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.Handler(req)
}