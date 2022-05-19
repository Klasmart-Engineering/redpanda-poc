package main

import (
	"producer/example/pkg/metrics"
	"producer/example/pkg/services/pushgateway"
)

const producerCount int = 4

var (
	bytesCollector      = metrics.InitGauge("bytes_in", "Bytes produced by producers pool")
	pushgatewayInstance = pushgateway.InitPushGateway()

	messages = []string{
		"Welcome to",
		"the Confluent",
		"Kafka Golang",
		"client",
	}
)

func init() {
	metrics.RegisterCollector(bytesCollector)
}

func main() {
	idx := 0
	for producerId := 0; producerId < producerCount; producerId++ {
		produce(producerId, messages[idx:idx+1])
		idx = idx + 1
	}
}
