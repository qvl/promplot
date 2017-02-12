package promplot

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api/prometheus"
	"github.com/prometheus/common/model"
)

// Metrics fetches data from Prometheus.
func Metrics(server, query string, queryTime time.Time, duration, step time.Duration) (model.Matrix, error) {
	client, err := prometheus.New(prometheus.Config{Address: server})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %v", err)
	}

	api := prometheus.NewQueryAPI(client)
	value, err := api.QueryRange(context.Background(), query, prometheus.Range{
		Start: queryTime.Add(-duration),
		End:   queryTime,
		Step:  duration / step,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query Prometheus: %v", err)
	}

	metrics, ok := value.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unsupported result format: %s", value.Type().String())
	}

	return metrics, nil
}
