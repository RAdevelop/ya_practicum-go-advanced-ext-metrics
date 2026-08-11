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

// MetricIn - данные по метрике, которые надо отправить на сервер
type MetricIn struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Value string `json:"value,omitempty"`
	Delta string `json:"delta,omitempty"`
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

func (a HttpAgent) Get(metric MetricIn) (*http.Response, error) {
	url := "/value/" + metric.Type + "/" + metric.ID

	resp, err := a.client.R().
		SetHeader("Content-Type", "text/plain").
		SetDoNotParseResponse(true).
		Get(url)

	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}
