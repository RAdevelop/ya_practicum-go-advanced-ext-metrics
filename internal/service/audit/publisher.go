package audit

import (
	"context"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
)

// Publisher — хранит список приёмников и оповещает их.
type Publisher struct {
	auditors []Auditor
	log      logger.Logger
}

func NewPublisher(log logger.Logger, auditors ...Auditor) *Publisher {
	return &Publisher{auditors: auditors, log: log}
}

// Notify — отправляет событие во все приёмники.
// Ошибки одного приёмника не прерывают остальные.
func (p *Publisher) Notify(ctx context.Context, event Event) {
	for _, Auditor := range p.auditors {
		if err := Auditor.Save(ctx, event); err != nil {
			p.log.Error("auditor "+auditorName(Auditor)+" failed", "err", err)
		}
	}
}

func auditorName(Auditor Auditor) string {
	switch Auditor.(type) {
	case *FileAuditor:
		return "file"
	case *HTTPAuditor:
		return "http"
	default:
		return "unknown"
	}
}
