package main

import (
  "encoding/binary"
	"encoding/json"
	"fmt"
  "strconv"
	"io/ioutil"

  "github.com/google/uuid"
	"github.com/riferrei/srclient"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type ComplexType struct {
	ID   string  `json:"ID"`
	Value string `json:"Value"`
}

func main() {

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
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

  schemaRegistryClient := srclient.CreateSchemaRegistryClient("http://localhost:8081")
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

	for i, word := range []string{"Welcome", "to", "the", "Confluent", "Kafka", "Golang", "client"} {
    messageId := strconv.Itoa(i+1)
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
	}

	// Wait for message deliveries before shutting down
	p.Flush(15 * 1000)

}
