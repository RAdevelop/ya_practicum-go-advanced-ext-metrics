package agent

import (
	"encoding/json"
	"net/http"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/go-resty/resty/v2"
)

// HttpAgent - http клиент для отправки метрик на сервер
type HttpAgent struct {
	client *resty.Client
}

func New(client *resty.Client) *HttpAgent {
	return &HttpAgent{
		client: client,
	}
}

func (a HttpAgent) Update(metric models.Metrics) (*http.Response, error) {
	url := "/update"
	body, err := json.Marshal(metric)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.R().
		SetHeader("Content-Type", "application/json").
		SetDoNotParseResponse(true).
		SetBody(body).
		Post(url)

	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}
