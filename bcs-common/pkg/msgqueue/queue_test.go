/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package msgqueuev4

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"go-micro.dev/v4/broker"
)

// MockPod data
type MockPod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func connectQueue() (MessageQueue, error) {
	var (
		resourceToQueue = map[string]string{
			"Pod":   "Pod",
			"Event": "Event",
		}
		queueOptions []QueueOption
	)

	commonOpts := CommonOpts(
		&CommonOptions{
			QueueFlag:       true,
			QueueKind:       RABBITMQ,
			ResourceToQueue: resourceToQueue,
			Address:         DefaultRabbitURL,
		})

	exchangeOpts := Exchange(
		&ExchangeOptions{
			Name:           "bcs-storage-test",
			Durable:        true,
			PrefetchCount:  0,
			PrefetchGlobal: false,
		})

	publishOpts := PublishOpts(
		&PublishOptions{
			TopicName:    "Pod",
			DeliveryMode: 2,
		})

	subscribeOpts := SubscribeOpts(&SubscribeOptions{
		TopicName:      "Pod",
		QueueName:      "Pod",
		DisableAutoAck: true,
		Durable:        true,
		AckOnSuccess:   true,
		RequeueOnError: true,
		QueueArguments: map[string]interface{}{
			"x-message-ttl": 60000,
		},
	})
	queueOptions = append(queueOptions, commonOpts, exchangeOpts, publishOpts, subscribeOpts)

	queue, err := NewMsgQueue(queueOptions...)
	if err != nil {
		return nil, err
	}

	return queue, err
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

	sub, err := q.Subscribe(&testHandler{name: "hello world"}, []Filter{
		&DefaultNamespaceFilter{
			FilterKind: "namespace",
			Namespace:  "default",
		},
		&DefaultClusterFilter{
			FilterKind: "clusterId",
			ClusterID:  "1",
		},
	}, PodSubscribeType)
	if err != nil {
		t.Fatal(err)
	}
	defer func(sub UnSub) {
		_ = sub.Unsubscribe()
	}(sub)

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

	// go pub(q)
	sub, err := q.Subscribe(&testHandler{name: "hello world"}, []Filter{
		&DefaultClusterFilter{
			FilterKind: "clusterId",
			ClusterID:  "1",
		},
	}, PodSubscribeType)
	if err != nil {
		t.Fatal(err)
	}
	defer func(sub UnSub) {
		_ = sub.Unsubscribe()
	}(sub)

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
	var (
		handlerData = &HandlerData{}
		pod         = &MockPod{}
	)
	_ = json.Unmarshal(data, handlerData)
	if len(handlerData.Body) > 0 {
		_ = json.Unmarshal(handlerData.Body, pod)
	}

	fmt.Println(time.Now().Unix(), " ", handlerData.ResourceType, " ", string(handlerData.Body), " ", pod.Name)

	return nil
}

func pub(queue MessageQueue) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	podData := []PodInfo{
		{
			MetaData: MetaData{
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
			MetaData: MetaData{
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
			MetaData: MetaData{
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
			MetaData: MetaData{
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
				filterToString(ClusterID):    podData[i].ClusterID,
				filterToString(ResourceType): podData[i].ResourceType,
				filterToString(Namespace):    podData[i].NameSpace,
				filterToString(EventType):    podData[i].Event,
				filterToString(ResourceName): podData[i].Name,
			},
		}
		msg.Body, _ = json.Marshal(podData[i].Data)
		messages = append(messages, msg)
	}

	cnt := 0
	rand.New(rand.NewSource(time.Now().Unix())) // nolint
	for range ticker.C {
		if len(messages) > 0 {
			msg := messages[rand.Intn(len(messages))] // nolint
			fmt.Println(msg)
			err := queue.Publish(msg)
			cnt++
			fmt.Println(cnt)
			if err != nil {
				_ = fmt.Errorf("publish message failed: %v", err)
				continue
			}
		}
	}
}

func filterToString(filterType FilterType) string {
	return string(filterType)
}
