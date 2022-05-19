package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/riferrei/srclient"
	"io/ioutil"
	"strconv"
	"time"
)

type ComplexType struct {
	ID    string `json:"ID"`
	Value string `json:"Value"`
}

func produce(producerId int, msg []string) {

	bytesCollector.WithLabelValues(strconv.Itoa(producerId)).Set(0.0)
	pushgatewayInstance.Push(bytesCollector)

	// Hang for 2s in order to make prometheus to catch up with the latest updated metrics
	// Prometheus scrap time is 1s for the pushgateway
	time.Sleep(2 * time.Second)

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:19092"})
	if err != nil {
		panic(err)
	}
	defer p.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	// Produce messages to topic (asynchronously)
	topic := "test"

	schemaRegistryClient := srclient.CreateSchemaRegistryClient("http://localhost:18081")
	schema, err := schemaRegistryClient.GetLatestSchema(topic)
	fmt.Printf("Producer schema: %v\n", schema)
	if schema == nil {
		schemaBytes, _ := ioutil.ReadFile("complexType.avsc")
		schema, err = schemaRegistryClient.CreateSchema(topic, string(schemaBytes), srclient.Avro)
		if err != nil {
			panic(fmt.Sprintf("Error creating the schema %s", err))
		}
	}
	schemaIDBytes := make([]byte, 4)
	fmt.Printf("Producer schema id: %v\n", schema.ID)
	binary.BigEndian.PutUint32(schemaIDBytes, uint32(schema.ID()))

	for i, word := range msg {
		messageId := strconv.Itoa(i + 1)
		fmt.Printf("Producer id: %v\n", messageId)

		newComplexType := ComplexType{ID: messageId, Value: word}
		value, _ := json.Marshal(newComplexType)
		native, _, _ := schema.Codec().NativeFromTextual(value)
		fmt.Printf("Producer value: %v\n", native)
		valueBytes, _ := schema.Codec().BinaryFromNative(nil, native)

		var recordValue []byte
		recordValue = append(recordValue, byte(0))
		recordValue = append(recordValue, schemaIDBytes...)
		recordValue = append(recordValue, valueBytes...)

		key, _ := uuid.NewUUID()

		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            []byte(key.String()),
			Value:          recordValue,
		}, nil)

		fmt.Println("Total bytes in: ", len(recordValue))

		bytesCollector.WithLabelValues(strconv.Itoa(producerId)).Set(float64(len(recordValue)))
		pushgatewayInstance.Push(bytesCollector)
		time.Sleep(2 * time.Second)
		bytesCollector.WithLabelValues(strconv.Itoa(producerId)).Set(0.0)
		pushgatewayInstance.Push(bytesCollector)
	}

	// Wait for message deliveries before shutting down
	p.Flush(15 * 1000)
}
