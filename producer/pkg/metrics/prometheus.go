package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

func InitGauge(name string, help string) *prometheus.GaugeVec {

	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	},
		[]string{"producer"},
	)
}

func InitCounter(name string, help string) prometheus.Counter {

	return prometheus.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	},
	)
}

func RegisterCollector(collector prometheus.Collector) {
	err := prometheus.Register(collector)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Registered successfully")
	}
}
