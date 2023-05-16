/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kafka

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// KafkaInterface kafka interface
type KafkaInterface interface {
	Publish(ctx context.Context, message []byte) error
	PublishWithTopic(ctx context.Context, topic string, partition int, message []byte) error
	Stop() error
}

type kafkaClient struct {
	kafkaWriter *kafka.Writer
	kafkaReader *kafka.Reader
	topicPrefix string
}

// NewKafkaClient return kafka interface
func NewKafkaClient(writer *kafka.Writer, reader *kafka.Reader, prefix string) KafkaInterface {
	return &kafkaClient{
		kafkaWriter: writer,
		kafkaReader: reader,
		topicPrefix: prefix,
	}
}

// Publish publish kafka message
func (c *kafkaClient) Publish(ctx context.Context, message []byte) error {
	return c.kafkaWriter.WriteMessages(ctx, kafka.Message{Value: message})
}

// Stop stop kafka client
func (c *kafkaClient) Stop() error {
	if c.kafkaWriter != nil {
		err := c.kafkaWriter.Close()
		if err != nil {
			return err
		}
	}
	if c.kafkaReader != nil {
		return c.kafkaReader.Close()
	}
	return nil
}

// PublishWithTopic publish message with topic
func (c *kafkaClient) PublishWithTopic(ctx context.Context, topic string, partition int, message []byte) error {
	if c.topicPrefix != "" {
		topic = fmt.Sprintf("%s-%s", c.topicPrefix, topic)
	}
	return c.kafkaWriter.WriteMessages(ctx, kafka.Message{
		Topic:     topic,
		Partition: partition,
		Value:     message,
	})
}
