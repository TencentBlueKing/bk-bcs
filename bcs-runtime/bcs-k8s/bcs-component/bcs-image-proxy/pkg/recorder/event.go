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

// Package recorder xxx
package recorder

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/utils"
)

// EventType defines the type of event
type EventType string

const (
	// Normal type of event
	Normal EventType = "Normal"
	// Warning type of event
	Warning EventType = "Warning"
)

// Event defines the event object
type Event struct {
	RequestID string `json:"requestID"`
	Registry  string `json:"registry"`
	Repo      string `json:"repo"`
	Tag       string `json:"tag"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`

	EventType EventType `json:"eventType"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

// EventRecorder defines the event recorder instance
type EventRecorder struct {
	op *options.ImageProxyOption

	ch chan *Event
}

var (
	globalRecorder *EventRecorder
)

// GlobalRecorder return the global recorder instance
func GlobalRecorder() *EventRecorder {
	if globalRecorder != nil {
		return globalRecorder
	}
	globalRecorder = &EventRecorder{
		op: options.GlobalOptions(),
		ch: make(chan *Event, 1000),
	}
	return globalRecorder
}

var (
	objKey = "ctxobj"
)

// SetManifestRequest set manifest request to context
func (r *EventRecorder) SetManifestRequest(ctx context.Context, registry, repo, tag string) context.Context {
	return context.WithValue(ctx, objKey, &Event{
		RequestID: logctx.RequestID(ctx),
		Registry:  registry,
		Repo:      repo,
		Tag:       tag,
	})
}

// SetLayerRequest set layer request to context
func (r *EventRecorder) SetLayerRequest(ctx context.Context, registry, repo, digest string) context.Context {
	return context.WithValue(ctx, objKey, &Event{
		RequestID: logctx.RequestID(ctx),
		Registry:  registry,
		Repo:      repo,
		Digest:    digest,
	})
}

// SendObjEvent send obj event
func (r *EventRecorder) SendObjEvent(ctx context.Context, eventType EventType, message string) {
	v := ctx.Value(objKey)
	if v == nil {
		return
	}
	manifestEvent := v.(*Event)
	newEvent := &Event{}
	if err := utils.DeepCopyStruct(manifestEvent, newEvent); err != nil {
		logctx.Errorf(ctx, "deep-copy struct failed: %v", err)
		return
	}
	newEvent.EventType = eventType
	newEvent.Message = message
	newEvent.CreatedAt = time.Now()
	r.SendEvent(newEvent)
}

// SendEvent send event
func (r *EventRecorder) SendEvent(event *Event) {
	r.ch <- event
}

// Run write the event to local file
func (r *EventRecorder) Run(ctx context.Context) error {
	file, err := os.OpenFile(r.op.EventFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0755)
	if err != nil && !os.IsExist(err) {
		return errors.Wrapf(err, "open event file '%s' failed", r.op.EventFile)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()
	for {
		select {
		case mes := <-r.ch:
			var bs []byte
			bs, err = json.Marshal(mes)
			if err != nil {
				blog.Warnf("marshal event message failed: %s", err.Error())
				continue
			}
			if _, err = writer.Write(bs); err != nil {
				blog.Errorf("write event message failed: %s", err.Error())
			}
			writer.WriteString("\n")
			writer.Flush()
		case <-ctx.Done():
			return nil
		}
	}
}
