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

package worker

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/micro/go-micro/v2/broker"
)

// MockPod data
type MockPod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func connectQueue() (msgqueue.MessageQueue, error) {
	// var (
	//	resourceToQueue = map[string]string{
	//		common.DataJobQueue: common.DataJobQueue,
	//	}
	// )
	commonOption := msgqueue.CommonOpts(&msgqueue.CommonOptions{
		QueueFlag:       true,
		QueueKind:       msgqueue.QueueKind("rabbitmq"),
		ResourceToQueue: map[string]string{common.DataJobQueue: common.DataJobQueue},
		// ResourceToQueue: resourceToQueue,
		Address: "amqp://root:123456@127.0.0.1:5672",
	})
	exchangeOption := msgqueue.Exchange(
		&msgqueue.ExchangeOptions{
			Name:           "bcs-data-manager",
			Durable:        true,
			PrefetchCount:  30,
			PrefetchGlobal: true,
		})
	natStreamingOption := msgqueue.NatsOpts(
		&msgqueue.NatsOptions{
			ClusterID:      "",
			ConnectTimeout: time.Duration(300) * time.Second,
			ConnectRetry:   true,
		})
	publishOption := msgqueue.PublishOpts(
		&msgqueue.PublishOptions{
			TopicName:    common.DataJobQueue,
			DeliveryMode: uint8(2),
		})
	arguments := make(map[string]interface{})
	queueArgumentsRaw := "x-message-ttl:120000"
	queueArguments := strings.Split(queueArgumentsRaw, ";")
	if len(queueArguments) > 0 {
		for _, data := range queueArguments {
			dList := strings.Split(data, ":")
			if len(dList) == 2 {
				arguments[dList[0]] = dList[1]
			}
		}
	}
	fmt.Println(arguments)
	subscribeOption := msgqueue.SubscribeOpts(
		&msgqueue.SubscribeOptions{
			TopicName:      common.DataJobQueue,
			QueueName:      common.DataJobQueue,
			DisableAutoAck: true,
			Durable:        true,
			AckOnSuccess:   true,
			RequeueOnError: true,

			//
			// DeliverAllMessage: true,
			// ManualAckMode:     true,
			// EnableAckWait:     true,
			//
			// AckWaitDuration: time.Duration(30) * time.Second,
			// MaxInFlight:     0,

			// QueueArguments: arguments,
			QueueArguments: map[string]interface{}{
				"x-message-ttl": 120000,
			},
		})
	msgQueue, err := msgqueue.NewMsgQueue(commonOption, exchangeOption, natStreamingOption, publishOption, subscribeOption)
	if err != nil {
		return nil, err
	}

	return msgQueue, err

	// var (
	//	resourceToQueue = map[string]string{
	//		"Pod":   "Pod",
	//		"Event": "Event",
	//	}
	//	queueOptions []msgqueue.QueueOption
	// )
	//
	// commonOpts := msgqueue.CommonOpts(
	//	&msgqueue.CommonOptions{
	//		QueueFlag:       true,
	//		QueueKind:       "rabbitmq",
	//		ResourceToQueue: resourceToQueue,
	//		Address:         "amqp://root:123456@127.0.0.1:5672",
	//	})
	//
	// exchangeOpts := msgqueue.Exchange(
	//	&msgqueue.ExchangeOptions{
	//		Name:           "bcs-storage-test",
	//		Durable:        true,
	//		PrefetchCount:  0,
	//		PrefetchGlobal: false,
	//	})
	//
	// publishOpts := msgqueue.PublishOpts(
	//	&msgqueue.PublishOptions{
	//		TopicName:    "Pod",
	//		DeliveryMode: 2,
	//	})
	//
	// subscribeOpts := msgqueue.SubscribeOpts(&msgqueue.SubscribeOptions{
	//	TopicName:      "Pod",
	//	QueueName:      "Pod",
	//	DisableAutoAck: true,
	//	Durable:        true,
	//	AckOnSuccess:   true,
	//	RequeueOnError: true,
	//	QueueArguments: map[string]interface{}{
	//		"x-message-ttl": 60000,
	//	},
	// })
	// queueOptions = append(queueOptions, commonOpts, exchangeOpts, publishOpts, subscribeOpts)
	//
	// queue, err := msgqueue.NewMsgQueue(queueOptions...)
	// if err != nil {
	//	return nil, err
	// }
	//
	// return queue, err
}

func TestMsgQueue_Publish(t *testing.T) {
	q, err := connectQueue()
	if err != nil {
		t.Error(err)
		return
	}
	defer q.Stop()

	qType, _ := q.String()
	t.Log(qType)

	go pub(q)

	sub, err := q.Subscribe(&testHandler{name: "hello world"}, []msgqueue.Filter{
		&msgqueue.DefaultNamespaceFilter{
			FilterKind: "namespace",
			Namespace:  "default",
		},
		&msgqueue.DefaultClusterFilter{
			FilterKind: "clusterId",
			ClusterID:  "1",
		},
	}, common.DataJobQueue)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Unsubscribe()

	select {}
}

func TestMsgQueue_Subscribe(t *testing.T) {
	q, err := connectQueue()
	if err != nil {
		t.Error(err)
		return
	}
	defer q.Stop()

	qType, _ := q.String()
	t.Log(qType)

	go pub(q)
	sub, err := q.Subscribe(&testHandler{name: "hello world"}, []msgqueue.Filter{
		&msgqueue.DefaultClusterFilter{
			FilterKind: "clusterId",
			ClusterID:  "1",
		},
	}, common.DataJobQueue)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Unsubscribe()

	select {}
}

type testHandler struct {
	name string
}

// Name show handler name
func (h *testHandler) Name() string {
	return h.name
}

// Handle handle data
func (h *testHandler) Handle(ctx context.Context, data []byte) error {
	fmt.Println("enter handle")
	var (
		handlerData = &msgqueue.HandlerData{}
		pod         = &MockPod{}
	)
	json.Unmarshal(data, handlerData)
	if len(handlerData.Body) > 0 {
		json.Unmarshal(handlerData.Body, pod)
	}

	fmt.Println(time.Now().Unix(), " ", handlerData.ResourceType, " ", string(handlerData.Body), " ", pod.Name)

	return nil
}

func pub(queue msgqueue.MessageQueue) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	podData := []msgqueue.PodInfo{
		{
			MetaData: msgqueue.MetaData{
				ClusterID:    "1",
				NameSpace:    "default",
				ResourceType: "Pod",
				Name:         "pod1",
			},
			Data: &MockPod{
				Name:      "pod1 world",
				Namespace: "default",
			},
		},
		{
			MetaData: msgqueue.MetaData{
				ClusterID:    "1",
				NameSpace:    "kube-system",
				ResourceType: "Pod",
				Name:         "pod2",
			},
			Data: &MockPod{

				Name:      "pod2 world",
				Namespace: "kube-system",
			},
		},
		{
			MetaData: msgqueue.MetaData{
				ClusterID:    "2",
				NameSpace:    "kube-qwer",
				ResourceType: "Pod",
				Name:         "pod3",
			},
			Data: &MockPod{
				Name:      "pod3 world",
				Namespace: "kube-qwer",
			},
		},
		{
			MetaData: msgqueue.MetaData{
				ClusterID:    "2",
				NameSpace:    "kube-asdf",
				ResourceType: "Pod",
				Name:         "pod4",
			},
			Data: &MockPod{
				Name:      "pod4 world",
				Namespace: "kube-asdf",
			},
		},
	}

	var messages []*broker.Message
	for i := range podData {
		msg := &broker.Message{
			Header: map[string]string{
				"clusterId":    podData[i].ClusterID,
				"resourceType": common.DataJobQueue,
				"Namespace":    podData[i].NameSpace,
				"EventType":    podData[i].Event,
				"ResourceName": podData[i].Name,
			},
		}
		msg.Body, _ = json.Marshal(podData[i].Data)
		messages = append(messages, msg)
	}

	cnt := 0
	for range ticker.C {
		if len(messages) > 0 {
			randNum, _ := rand.Int(rand.Reader, big.NewInt(int64(len(messages))))
			msg := messages[randNum.Int64()]
			// fmt.Println(msg)
			err := queue.Publish(msg)
			cnt++
			// fmt.Println(cnt)
			if err != nil {
				_ = fmt.Errorf("publish message failed: %v", err)
				continue
			}
		}
	}
}
