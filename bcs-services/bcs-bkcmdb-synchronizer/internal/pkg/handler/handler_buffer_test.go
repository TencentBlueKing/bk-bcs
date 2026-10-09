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

package handler

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestMsgBufferReadyToFlush(t *testing.T) {
	now := time.Date(2026, 9, 11, 10, 48, 8, 0, time.UTC)

	tests := []struct {
		name string
		buf  *msgBuffer
		want bool
	}{
		{
			name: "nil buffer",
			buf:  nil,
			want: false,
		},
		{
			name: "empty buffer never flushes",
			buf: &msgBuffer{
				T: now.Add(-time.Hour),
				M: nil,
			},
			want: false,
		},
		{
			name: "below size and within interval",
			buf: &msgBuffer{
				T: now.Add(-5 * time.Second),
				M: make([]amqp.Delivery, 36),
			},
			want: false,
		},
		{
			name: "below size but interval elapsed",
			buf: &msgBuffer{
				T: now.Add(-msgBufferFlushInterval),
				M: make([]amqp.Delivery, 36),
			},
			want: true,
		},
		{
			name: "reach batch size before interval",
			buf: &msgBuffer{
				T: now.Add(-time.Second),
				M: make([]amqp.Delivery, msgBufferFlushSize),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.buf.readyToFlush(now); got != tt.want {
				t.Fatalf("readyToFlush() = %v, want %v", got, tt.want)
			}
		})
	}
}
