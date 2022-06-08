package main

import (
  "encoding/binary"
	"fmt"

	"github.com/riferrei/srclient"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
    "bootstrap.servers": "localhost:9092",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	c.SubscribeTopics([]string{"test", "^aRegex.*[Tt]opic"}, nil)
  schemaRegistryClient := srclient.CreateSchemaRegistryClient("http://localhost:8081")

	for {
		msg, err := c.ReadMessage(-1)
		fmt.Printf("Consumer id: %v\n", msg)
		if err == nil {
      schemaID := binary.BigEndian.Uint32(msg.Value[1:5])
		  schema, err := schemaRegistryClient.GetSchema(int(schemaID))

      if err != nil {
				panic(fmt.Sprintf("Error getting the schema with id '%d' %s", schemaID, err))
			}

		  native, _, _ := schema.Codec().NativeFromBinary(msg.Value[5:])
		  value, _ := schema.Codec().TextualFromNative(nil, native)
		  fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(value))
		} else {
			// The client will automatically try to recover from all errors.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}

	c.Close()
}
