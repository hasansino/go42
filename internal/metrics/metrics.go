package metrics

import (
	"fmt"

	vmetrics "github.com/VictoriaMetrics/metrics"
)

var globalLabels map[string]string

func init() {
	globalLabels = make(map[string]string)
}

func RegisterGlobalLabels(lables map[string]string) {
	if len(lables) == 0 {
		return
	}
	for k, v := range lables {
		globalLabels[k] = v
	}
}

func Counter(name string, labels map[string]interface{}) *vmetrics.Counter {
	return vmetrics.GetOrCreateCounter(constructMetric(name, labels))
}

func Gauge(name string, labels map[string]interface{}) *vmetrics.Gauge {
	return vmetrics.GetOrCreateGauge(constructMetric(name, labels), nil)
}

func constructMetric(name string, labels map[string]interface{}) string {
	metric := name
	if len(labels) > 0 {
		metric += "{"
		for k, v := range labels {
			metric += fmt.Sprintf("%s=\"%s\",", k, v)
		}
		metric = metric[:len(metric)-1] + "}"
	}
	return metric
}
