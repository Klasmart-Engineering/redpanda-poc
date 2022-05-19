package pushgateway

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

type pushgateway struct {
	url string
}

func InitPushGateway() *pushgateway {
	return &pushgateway{
		url: "http://localhost:9091/",
	}
}

func (pg pushgateway) Push(c prometheus.Collector) {
	err := push.New(pg.url, "producer").
		Collector(c).
		Push()

	if err != nil {
		fmt.Println("Could not push bytes to Pushgateway:", err)
	}
}
