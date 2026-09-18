package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/agent"
	configAgent "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/agent"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/collector"
	"github.com/go-resty/resty/v2"
)

type agentFlags struct {
	IntervalReport *uint
	IntervalPoll   *uint
	SingKey        *string
	RateLimit      *uint
}

func main() {
	logApp := logger.New()

	srvAddress := &agent.ServerAddress{Host: "localhost", Port: 8080}
	_ = flag.Value(srvAddress)

	agFlags := &agentFlags{}
	flag.Var(srvAddress, "a", `Адрес сервера: "host:port" без схемы`)
	agFlags.IntervalReport = flag.Uint("r", 10, `Частота в секундах для отправки метрик на сервер`)
	agFlags.IntervalPoll = flag.Uint("p", 2, `Частота в секундах для сбора метрик`)
	agFlags.SingKey = flag.String("k", "", `Ключ для подписи запроса`)
	agFlags.RateLimit = flag.Uint("l", 10, `Количество одновременно исходящих запросов на сервер`)
	flag.Parse()

	configAgentEnv, err := configAgent.NewEnv()
	if err != nil {
		logApp.Error("error", fmt.Errorf("error creating configAgent environment variable: %w", err))
		return
	}

	agentConfig := configAgent.New(configAgentEnv)
	agentConfigUpdate(agentConfig, srvAddress.String(), agFlags)

	httpClient := resty.New()
	httpClient.SetBaseURL("http://" + agentConfig.Address())
	httpAgent := agent.New(httpClient, agentConfig)

	// Контекст, который отменяется по SIGINT/SIGTERM.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pollInterval := time.Duration(agentConfig.PollInterval()) * time.Second
	reportInterval := time.Duration(agentConfig.ReportInterval()) * time.Second
	rateLimit := int(agentConfig.RateLimit())
	if rateLimit < 1 {
		rateLimit = 1
	}

	// Общий буфер метрик, защищённый мьютексом.
	metricCollector := collector.New(httpAgent, logApp)

	var wg sync.WaitGroup

	// 1. Горутина сбора runtime-метрик.
	wg.Add(1)
	go func() {
		defer wg.Done()
		metricCollector.RunPollWorker(ctx, pollInterval)
	}()

	// 2. Горутина сбора gopsutil-метрик.
	wg.Add(1)
	go func() {
		defer wg.Done()
		metricCollector.RunPsutilWorker(ctx, pollInterval)
	}()

	// 3. Пул воркеров-отправителей.
	jobs := make(chan []models.Metrics, rateLimit)
	wg.Add(rateLimit)
	for i := 0; i < rateLimit; i++ {
		go func(id int) {
			defer wg.Done()
			metricCollector.RunSenderWorker(ctx, id, jobs)
		}(i)
	}

	// 4. Горутина, которая раз в reportInterval забирает снапшот буфера
	//    и кладёт его в jobs.
	wg.Add(1)
	go func() {
		defer wg.Done()
		metricCollector.RunReportWorker(ctx, reportInterval, jobs)
	}()

	// Ждём сигнал завершения.
	<-ctx.Done()
	logApp.Info("agent shutting down")

	// Закрываем jobs, чтобы воркеры дочитали и вышли.
	close(jobs)
	wg.Wait()
}

func agentConfigUpdate(agentConfig *configAgent.Config, srvAddress string, agFlags *agentFlags) {

	if agentConfig == nil {
		return
	}

	if agentConfig.Address() == "" {
		agentConfig.AddressSet(srvAddress)
	}

	if agFlags == nil {
		return
	}

	if agentConfig.ReportInterval() == 0 && agFlags.IntervalReport != nil {
		agentConfig.ReportIntervalSet(*agFlags.IntervalReport)
	}

	if agentConfig.PollInterval() == 0 && agFlags.IntervalPoll != nil {
		agentConfig.PollIntervalSet(*agFlags.IntervalPoll)
	}

	if agentConfig.SignKey() == "" && agFlags.SingKey != nil {
		agentConfig.SignKeySet(*agFlags.SingKey)
	}

	if agentConfig.RateLimit() == 0 && agFlags.RateLimit != nil {
		agentConfig.RateLimitSet(*agFlags.RateLimit)
	}
}
