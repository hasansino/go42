package metrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	vmetrics "github.com/VictoriaMetrics/metrics"
)

var (
	mutex        sync.RWMutex
	globalLabels map[string]interface{}
)

func init() {
	globalLabels = make(map[string]interface{})
}

func RegisterGlobalLabels(labels map[string]interface{}) {
	if len(labels) == 0 {
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	for k, v := range labels {
		globalLabels[k] = v
	}
}

func Counter(name string, labels map[string]interface{}) *vmetrics.Counter {
	return vmetrics.GetOrCreateCounter(constructMetric(name, labels))
}

func Gauge(name string, labels map[string]interface{}) *vmetrics.Gauge {
	return vmetrics.GetOrCreateGauge(constructMetric(name, labels), nil)
}

// constructMetric takes metrics name and labels, and returns a string representation of the metric.
// It also adds global labels to the metric if they are set.
// It is a little bit more complicated than just appending labels to the name, because
// we need to sort the labels to make sure that the same labels in different order
// will not create different metrics.
func constructMetric(name string, labels map[string]interface{}) string {
	// mutex actually may not be needed at all if we don't modify globalLabels
	// aside from single call in main.go, but keep it for safety anyway
	mutex.RLock()
	defer mutex.RUnlock()

	if len(labels) == 0 && len(globalLabels) == 0 {
		return name
	}

	totalLabels := len(labels)
	globalLabelsCount := len(globalLabels)

	if globalLabelsCount > 0 {
		for k := range globalLabels {
			if _, exists := labels[k]; !exists {
				totalLabels++
			}
		}
	}

	if totalLabels == 0 {
		return name
	}

	keys := make([]string, 0, totalLabels)

	for k := range labels {
		keys = append(keys, k)
	}
	if globalLabelsCount > 0 {
		for k := range globalLabels {
			if _, exists := labels[k]; !exists {
				keys = append(keys, k)
			}
		}
	}

	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteString(name)
	builder.WriteByte('{')

	for i, k := range keys {
		if i > 0 {
			builder.WriteByte(',')
		}

		builder.WriteString(k)
		builder.WriteString(`="`)

		var v interface{}
		if val, exists := labels[k]; exists {
			v = val
		} else {
			v = globalLabels[k]
		}

		// fmt.Sprintf may be slow, but it is trade off for convenience
		builder.WriteString(fmt.Sprintf("%v", v))
		builder.WriteByte('"')
	}

	builder.WriteByte('}')

	return builder.String()
}
