/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package mock

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/micro/go-micro/v2/broker"
	"github.com/stretchr/testify/mock"
)

// MockQueue mock queue
type MockQueue struct {
	mock.Mock
}

// Publish mock queue
func (m *MockQueue) Publish(data *broker.Message) error {
	args := m.Called(data)
	return args.Error(0)
}

// Subscribe mock queue
func (m *MockQueue) Subscribe(handler msgqueue.Handler, filters []msgqueue.Filter,
	resourceType string) (msgqueue.UnSub, error) {
	args := m.Called(handler, filters, resourceType)
	return args.Get(0).(msgqueue.UnSub), args.Error(1)
}

// SubscribeWithQueueName mock queue
func (m *MockQueue) SubscribeWithQueueName(handler msgqueue.Handler, filters []msgqueue.Filter, queueName,
	topic string) (msgqueue.UnSub, error) {
	args := m.Called(handler, filters, queueName, topic)
	return args.Get(0).(msgqueue.UnSub), args.Error(1)
}

// String return queue name
func (m *MockQueue) String() (string, error) {
	args := m.Called()
	return args.Get(0).(string), args.Error(1)
}

// Stop the message queue
func (m *MockQueue) Stop() {

}

// NewMockQueue new mock queue
func NewMockQueue() *MockQueue {
	return &MockQueue{}
}
