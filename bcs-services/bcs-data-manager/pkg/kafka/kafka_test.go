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
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"github.com/stretchr/testify/assert"
)

func TestPublish(t *testing.T) {
	//conn := getTestKafkaConn()
	//assert.NotNil(t, conn)
	//for i := 0; i < 2; i++ {
	//	err := conn.PublishWithTopic("datamanager", 0, []byte("test"))
	//	assert.Nil(t, err)
	//}
	client := getKafkaInterface()
	err := client.PublishWithTopic(context.Background(), "datamanager", 0, []byte("test"))
	fmt.Println(err)
	assert.Nil(t, err)
}

func getKafkaInterface() KafkaInterface {
	mechanism, err := scram.Mechanism(scram.SHA512, "", "")
	if err != nil {
		panic(err)
	}

	// Transports are responsible for managing connection pools and other resources,
	// it's generally best to create a few of these and share them across your
	// application.
	sharedTransport := &kafka.Transport{
		SASL: mechanism,
	}
	writer := &kafka.Writer{
		Addr: kafka.TCP(""),
		//Topic:                  "datamanager",
		MaxAttempts:            3,
		AllowAutoTopicCreation: true,
		Transport:              sharedTransport,
	}
	return NewKafkaClient(writer, nil, "")
}
