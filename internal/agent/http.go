package agent

import (
	"net/http"
)

// HttpAgent - http клиент для отправки метрик на сервер
type HttpAgent struct {
	client *http.Client
}

// MetricIn - данные по метрике, которые надо отправить на сервер
type MetricIn struct {
	Type  string
	Name  string
	Value string
}

func New(client *http.Client) *HttpAgent {
	return &HttpAgent{
		client: client,
	}
}

func (a HttpAgent) Update(metric MetricIn) (*http.Response, error) {

	//TODO хардкод адреса, стоит вынести в настройки на уровень конфига
	url := "http://localhost:8080/update/" + metric.Type + "/" + metric.Name + "/" + metric.Value

	resp, err := a.client.Post(url, "text/plain", nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
