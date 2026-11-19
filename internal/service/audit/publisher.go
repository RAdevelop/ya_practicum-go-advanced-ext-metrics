package audit

import (
	"context"
	"sync"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
)

// eventQueueSize — размер буфера канала на каждого приёмника.
// Если приёмник не успевает вычитывать и буфер заполнился,
// новые события отбрасываются, чтобы не блокировать HTTP-обработчик.
const eventQueueSize = 1024

// worker — горутина-писатель одного приёмника.
type worker struct {
	auditor Auditor
	name    string
	events  chan Event
}

// Publisher — асинхронно рассылает события приёмникам.
type Publisher struct {
	workers []*worker
	log     logger.Logger
	wg      sync.WaitGroup
}

func NewPublisher(log logger.Logger, auditors ...Auditor) *Publisher {
	p := &Publisher{log: log}

	for _, auditor := range auditors {
		w := &worker{
			auditor: auditor,
			name:    auditorName(auditor),
			events:  make(chan Event, eventQueueSize),
		}

		p.wg.Add(1)
		go p.run(w)

		p.workers = append(p.workers, w)
	}

	return p
}

// Notify кладёт событие в каналы приёмников и мгновенно возвращается.
// Если буфер приёмника переполнен, событие для него отбрасывается.
func (p *Publisher) Notify(_ context.Context, event Event) {
	for _, w := range p.workers {
		select {
		case w.events <- event:
		default:
			p.log.Warn("audit event dropped, receiver queue is full", "auditor", w.name)
		}
	}
}

// run — цикл горутины для отдельно взятого приёмника: читает события, пока канал не закрыт.
func (p *Publisher) run(w *worker) {
	defer p.wg.Done()

	for event := range w.events {
		if err := w.auditor.Save(context.Background(), event); err != nil {
			p.log.Error("auditor " + w.name + " failed: " + err.Error())
		}
	}
}

// Stop - останавливает рассылку: закрывает каналы, дожидается завершения горутин.
// Вызывать после остановки HTTP-сервера, когда новые события уже не появляются.
func (p *Publisher) Stop() {
	for _, w := range p.workers {
		close(w.events)
	}
	p.wg.Wait()
}

func auditorName(auditor Auditor) string {
	switch auditor.(type) {
	case *FileAuditor:
		return "file"
	case *HTTPAuditor:
		return "http"
	default:
		return "unknown"
	}
}
