/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package watchbus

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

// EventBus dispatch event
type EventBus struct {
	cond           *operator.Condition
	subscribers    map[string]map[string]chan *drivers.WatchEvent
	sublock        sync.RWMutex
	topicListeners map[string]chan *drivers.WatchEvent
	db             drivers.DB
}

// NewEventBus create event bus
func NewEventBus(db drivers.DB) *EventBus {
	return &EventBus{
		subscribers:    make(map[string]map[string]chan *drivers.WatchEvent),
		topicListeners: make(map[string]chan *drivers.WatchEvent),
		db:             db,
	}
}

// SetCondition set condition
func (eb *EventBus) SetCondition(cond *operator.Condition) {
	eb.cond = cond
}

// create listener of topic
func (eb *EventBus) createListener(topic string) (chan *drivers.WatchEvent, error) {
	var conditionList []*operator.Condition
	if eb.cond != nil {
		conditionList = append(conditionList, eb.cond)
	}
	listenerCh, err := eb.db.Table(topic).Watch(conditionList).
		WithFullContent(true).
		DoWatch(context.Background())
	if err != nil {
		return nil, fmt.Errorf("start listener of topic %s failed, err %s", topic, err.Error())
	}
	go func() {
		for {
			select {
			case event := <-listenerCh:
				eb.dispatch(topic, event)
				if event.Type == drivers.EventError || event.Type == drivers.EventClose {
					eb.sublock.Lock()
					delete(eb.topicListeners, topic)
					eb.sublock.Unlock()
					return
				}
			}
		}
	}()
	return listenerCh, nil
}

// Subscribe subscribe event for certain topic, here topic is database table
func (eb *EventBus) Subscribe(topic, uuid string, ch chan *drivers.WatchEvent) error {
	eb.sublock.Lock()
	defer eb.sublock.Unlock()
	if topicChanMap, ok := eb.subscribers[topic]; ok {
		if _, idDup := topicChanMap[uuid]; idDup {
			return fmt.Errorf("uuid %s of topic %s is duplicated", uuid, topic)
		}
		topicChanMap[uuid] = ch
	} else {
		if _, listenerFound := eb.topicListeners[topic]; !listenerFound {
			listenerCh, err := eb.createListener(topic)
			if err != nil {
				return err
			}
			eb.topicListeners[topic] = listenerCh
		}
		topicChanMap := make(map[string]chan *drivers.WatchEvent)
		topicChanMap[uuid] = ch
		eb.subscribers[topic] = topicChanMap
	}
	return nil
}

// Unsubscribe unsubscribe topic
func (eb *EventBus) Unsubscribe(topic, uuid string) error {
	eb.sublock.Lock()
	defer eb.sublock.Unlock()
	if topicChanMap, ok := eb.subscribers[topic]; ok {
		if _, idFound := topicChanMap[uuid]; !idFound {
			return fmt.Errorf("no uuid %s of topic %s to unsubscribe", uuid, topic)
		}

		delete(topicChanMap, uuid)
	}
	return fmt.Errorf("no topic %s to subscribe", topic)
}

// dispatch event to subscribers
func (eb *EventBus) dispatch(topic string, e *drivers.WatchEvent) {
	eb.sublock.RLock()
	defer eb.sublock.RUnlock()
	if topicChanMap, ok := eb.subscribers[topic]; ok {
		// copy the slice, because we send event to chan slice in another goroutine
		var chanList []chan *drivers.WatchEvent
		for _, ch := range topicChanMap {
			chanList = append(chanList, ch)
		}
		go func(data *drivers.WatchEvent, channels []chan *drivers.WatchEvent) {
			for _, ch := range channels {
				ch <- data
			}
		}(e, chanList)
	}
}
