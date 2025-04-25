package metrics

import (
	"testing"
)

func TestConstructMetric(t *testing.T) {
	tests := []struct {
		name         string
		metricName   string
		labels       map[string]interface{}
		globalLabels map[string]interface{}
		want         string
	}{
		{
			name:       "empty labels",
			metricName: "test_metric",
			labels:     nil,
			want:       "test_metric",
		},
		{
			name:       "single label",
			metricName: "http_requests_total",
			labels: map[string]interface{}{
				"method": "GET",
			},
			want: `http_requests_total{method="GET"}`,
		},
		{
			name:       "multiple labels",
			metricName: "memory_usage",
			labels: map[string]interface{}{
				"host":     "server1",
				"instance": "prod",
				"region":   "us-east",
			},
			want: `memory_usage{host="server1",instance="prod",region="us-east"}`,
		},
		{
			name:       "labels with different value types",
			metricName: "mixed_metric",
			labels: map[string]interface{}{
				"bool":   true,
				"int":    42,
				"string": "value",
			},
			want: `mixed_metric{bool="true",int="42",string="value"}`,
		},
		{
			name:       "with global labels",
			metricName: "test_metric",
			labels: map[string]interface{}{
				"local": "value",
			},
			globalLabels: map[string]interface{}{
				"env":     "prod",
				"service": "api",
			},
			want: `test_metric{env="prod",local="value",service="api"}`,
		},
		{
			name:       "global labels override",
			metricName: "override_metric",
			labels: map[string]interface{}{
				"env": "dev",
			},
			globalLabels: map[string]interface{}{
				"env": "prod",
			},
			want: `override_metric{env="dev"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset and set global labels for this test
			globalLabels = make(map[string]interface{})
			if tt.globalLabels != nil {
				RegisterGlobalLabels(tt.globalLabels)
			}

			got := constructMetric(tt.metricName, tt.labels)
			if got != tt.want {
				t.Errorf("constructMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}
