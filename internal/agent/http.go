package agent

import (
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HttpAgent - http клиент для отправки метрик на сервер
type HttpAgent struct {
	client HTTPClient
}

// MetricIn - данные по метрике, которые надо отправить на сервер
type MetricIn struct {
	Type  string
	Name  string
	Value string
}

func New(client HTTPClient) *HttpAgent {
	return &HttpAgent{
		client: client,
	}
}

func (a HttpAgent) Update(metric MetricIn) (*http.Response, error) {

	//TODO хардкод адреса, стоит вынести в настройки на уровень конфига
	url := "http://localhost:8080/update/" + metric.Type + "/" + metric.Name + "/" + metric.Value

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/plain")

	return a.client.Do(req)
}
