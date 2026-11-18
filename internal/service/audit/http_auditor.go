package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

func (s *HTTPAuditor) Save(ctx context.Context, event Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create audit request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send audit request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("audit server returned status %d", resp.StatusCode)
	}

	return nil
}
