package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/retryer"
)

// HTTPAuditor — отправляет событие POST-запросом на указанный URL.
type HTTPAuditor struct {
	url    string
	client *http.Client
}

func NewHTTPAuditor(url string) *HTTPAuditor {
	return &HTTPAuditor{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// Save - сохраняет данные о событии в источник хранения данных аудитора
func (a *HTTPAuditor) Save(ctx context.Context, event Event) error {

	_, err := retryer.RetryLinear(ctx, func(ctx context.Context) (struct{}, error) {
		err := a.sendEvent(ctx, event)
		return struct{}{}, err
	}, 2, new(3))
	return err

}

// Close - освобождает ресурсы HTTP-клиента (закрывает idle-соединения).
func (a *HTTPAuditor) Close() error {
	a.client.CloseIdleConnections()
	return nil
}

func (a *HTTPAuditor) sendEvent(ctx context.Context, event Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create audit request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("send audit request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("audit server returned status %d", resp.StatusCode)
	}

	return nil
}
