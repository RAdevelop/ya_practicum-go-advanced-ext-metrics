package agent

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

// HttpAgent - http клиент для отправки метрик на сервер
type HttpAgent struct {
	client *resty.Client
}

// MetricIn - данные по метрике, которые надо отправить на сервер
type MetricIn struct {
	Type  string
	Name  string
	Value string
}

func New(client *resty.Client) *HttpAgent {
	return &HttpAgent{
		client: client,
	}
}

func (a HttpAgent) Update(metric MetricIn) (*http.Response, error) {

	//TODO хардкод адреса, стоит вынести в настройки на уровень конфига
	url := "/update/" + metric.Type + "/" + metric.Name + "/" + metric.Value

	resp, err := a.client.R().
		SetHeader("Content-Type", "text/plain").
		SetDoNotParseResponse(true).
		Post(url)

	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}
