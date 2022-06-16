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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
)

// Consumer for subscribe handler interface
type Consumer interface {
	Consume(queue msgqueue.MessageQueue) error
	Stop() error
}

// Consumers consumer type
type Consumers struct {
	ctx       context.Context
	cancel    context.CancelFunc
	msgQueue  msgqueue.MessageQueue
	consumers []Consumer
}

// NewConsumers init consumers
func NewConsumers(consumers []Consumer, queue msgqueue.MessageQueue) *Consumers {
	c := &Consumers{
		msgQueue:  queue,
		consumers: consumers,
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	return c
}

// Run run consumers
func (c *Consumers) Run() {
	if c == nil {
		blog.Fatalf("consumer is nil")
	}
	for idx := range c.consumers {
		go func(consumer Consumer) {
			defer func() {
				if r := recover(); r != nil {
					blog.Errorf("[monitor][panic] consumer panic: %v\n", r)
				}
			}()
			if err := consumer.Consume(c.msgQueue); err != nil {
				blog.Errorf("consume job error:%v", err)
			}
		}(c.consumers[idx])
	}
}

// Stop stop subscribe & close queue
func (c *Consumers) Stop() {
	blog.Info("receive term signal")
	for idx := range c.consumers {
		if err := c.consumers[idx].Stop(); err != nil {
			blog.Errorf("stop consumer err:%v", err)
		}
	}
}
