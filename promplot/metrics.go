package promplot

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Metrics fetches data from Prometheus.
func Metrics(server, query string, queryTime time.Time, duration, step time.Duration) (model.Matrix, error) {
	client, err := api.NewClient(api.Config{Address: server})
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus api client: %v", err)
	}

	promAPI := v1.NewAPI(client)

	value, _, err := promAPI.QueryRange(context.Background(), query, v1.Range{
		Start: queryTime.Add(-duration),
		End:   queryTime,
		Step:  duration / step,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query prometheus api: %v", err)
	}

	metrics, ok := value.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unsupported result format: %s", value.Type().String())
	}

	return metrics, nil
}
