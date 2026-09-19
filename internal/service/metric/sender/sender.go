package sender

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/agent"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
)

type Sender struct {
	httpAgent *agent.HTTPAgent
	logApp    logger.Logger
}

func New(httpAgent *agent.HTTPAgent, logApp logger.Logger) *Sender {
	return &Sender{
		httpAgent: httpAgent,
		logApp:    logApp,
	}
}

// RunWorker — читает батчи из jobs и шлёт их на сервер.
func (s *Sender) RunWorker(ctx context.Context, id int, jobs <-chan []models.Metrics) {
	for {
		select {
		case <-ctx.Done():
			return
		case metrics, ok := <-jobs:
			if !ok {
				return
			}
			if err := s.updateBatch(ctx, metrics); err != nil {
				s.logApp.Error("RunWorker", "worker", id, "err", err)
			}
		}
	}
}

func (s *Sender) updateBatch(ctx context.Context, metrics []models.Metrics) (err error) {

	resp, err := s.httpAgent.Updates(ctx, metrics)
	defer func() {
		if resp != nil && resp.Body != nil {
			closeErr := resp.Body.Close()
			err = errors.Join(err, closeErr)
		}
	}()
	return s.handleUpdateResponse(resp, err, metrics)
}

func (s *Sender) handleUpdateResponse(resp *http.Response, errResp error, metric any) (err error) {
	if errResp != nil {
		return fmt.Errorf("error updating metric: %v, err: %w", metric, errResp)
	}
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return fmt.Errorf("error body reading for updating metric: %v, err: %w", metric, err)
	}
	return nil
}
