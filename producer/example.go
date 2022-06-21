package main

import (
	"encoding/binary"
	"encoding/json"
	"os"

	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/riferrei/srclient"
)

type ComplexType struct {
	ID    string `json:"ID"`
	Value string `json:"Value"`
}

func mustGetEnv(key string) string {
	if val := os.Getenv(key); "" != val {
		return val
	}
	panic(fmt.Sprintf("environment variable %s unset", key))
}

const producerCount = 2

var (
	topic = "test"
	msg   = []string{"Welcome", "to", "the", "Confluent", "Kafka", "Golang", "client"}
	svc   *cloudwatch.CloudWatch
)

func createCustomMetric(val []byte, producerId int) {
	_, err := svc.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: aws.String("AWS/CustomMetric"),
		MetricData: []*cloudwatch.MetricDatum{
			&cloudwatch.MetricDatum{
				MetricName: aws.String("customMetric"),
				Unit:       aws.String("Bytes"),
				Value:      aws.Float64(float64(len(val))),
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  aws.String("producer"),
						Value: aws.String(strconv.Itoa(producerId)),
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Println("Error adding metrics:", err.Error())
		return
	}
}

func listMetrics() {
	result, err := svc.ListMetrics(&cloudwatch.ListMetricsInput{
		Namespace: aws.String("custom.metric"),
	})

	if err != nil {
		fmt.Println("Error getting metrics:", err.Error())
		return
	}

	for _, metric := range result.Metrics {
		fmt.Println(*metric.MetricName)

		for _, dim := range metric.Dimensions {
			fmt.Println(*dim.Name+":", *dim.Value)
			fmt.Println()
		}
	}
}

func init() {

	localstackURL := mustGetEnv("LOCALSTACK_URL")

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	sess.Config.Endpoint = &localstackURL

	svc = cloudwatch.New(sess)
}

func main() {
	i := 0

	for producerId := 0; producerId < producerCount; producerId++ {

		produce(msg[i:i+1], producerId)
		i = i + 1
	}

}

func produce(msg []string, producerId int) {

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

		newComplexType := ComplexType{ID: messageId, Value: word}
		value, err := json.Marshal(newComplexType)
		native, _, err := schema.Codec().NativeFromTextual(value)
		fmt.Printf("Producer value: %v\n", native)
		valueBytes, err := schema.Codec().BinaryFromNative(nil, native)

		if err != nil {
			_, err = svc.PutMetricData(&cloudwatch.PutMetricDataInput{
				Namespace: aws.String("AWS/CustomMetric"),
				MetricData: []*cloudwatch.MetricDatum{
					&cloudwatch.MetricDatum{
						MetricName: aws.String("FailedMessages"),
						Unit:       aws.String("Count"),
						Value:      aws.Float64(1),
					},
				},
			})

			if err != nil {
				fmt.Println("Error adding metrics:", err.Error())
				return
			}
		}

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
		fmt.Println("Total bytes produced: ", len(recordValue))

		_, err = svc.PutMetricData(&cloudwatch.PutMetricDataInput{
			Namespace: aws.String("AWS/CustomMetric"),
			MetricData: []*cloudwatch.MetricDatum{
				&cloudwatch.MetricDatum{
					MetricName: aws.String("customMetric"),
					Unit:       aws.String("Bytes"),
					Value:      aws.Float64(float64(len(recordValue))),
					Dimensions: []*cloudwatch.Dimension{
						&cloudwatch.Dimension{
							Name:  aws.String("producer"),
							Value: aws.String(strconv.Itoa(producerId)),
						},
					},
				},
			},
		})

		if err != nil {
			fmt.Println("Error adding metrics:", err.Error())
			return
		}
	}

	// Wait for message deliveries before shutting down
	p.Flush(15 * 1000)

}
