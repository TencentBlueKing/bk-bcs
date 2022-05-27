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
 *
 */

package msgqueue

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/asim/go-micro/v3/broker"
)

// consumer subscribe resource type
const (
	// PodSubscribeType for resource 'Pod'
	PodSubscribeType = "Pod"
	// EventSubscribeType for resource 'Event'
	EventSubscribeType = "Event"
	// DeploymentSubscribeType for resource 'Deployment'
	DeploymentSubscribeType = "Deployment"
	// StatefulSetSubscribeType for resource 'StatefulSet'
	StatefulSetSubscribeType = "StatefulSet"
	// DaemonSetSubscribeType for resource 'DaemonSet'
	DaemonSetSubscribeType = "DaemonSet"
)

// default handler context timeout
const (
	HandleTimeout = 10 * time.Second
)

// Handler handle event to data and inject consumer
type Handler interface {
	Name() string
	Handle(ctx context.Context, data []byte) error
}

// HandlerWrap function for Handler interface
func HandlerWrap(name string, f func(ctx context.Context, data []byte) error) *HandlerWrapper {
	return &HandlerWrapper{name, f}
}

// HandlerWrapper for Handler
type HandlerWrapper struct {
	NameValue string
	Impl      func(ctx context.Context, data []byte) error
}

// Handle of hw
func (hw *HandlerWrapper) Handle(ctx context.Context, data []byte) error {
	if hw == nil {
		return errors.New("nil handler")
	}
	return hw.Impl(ctx, data)
}

// Name of the handler
func (hw *HandlerWrapper) Name() string {
	if hw == nil {
		return "nil"
	}
	return hw.NameValue
}

// objectHandler inject data to subscriber by filter and handler
type objectHandler struct {
	resourceType string
	handler      Handler
	filter       []Filter
}

// HandlerData return data for external subscriber
type HandlerData struct {
	ResourceType string            `json:"resourceType"`
	Meta         map[string]string `json:"meta"`
	Body         []byte            `json:"body"`
}

func (object *objectHandler) selfHandler(b broker.Event) error {
	if object == nil || object.handler == nil {
		return nil
	}

	// headers: metaData; data: originData
	headers := b.Message().Header
	data := b.Message().Body

	if len(headers) == 0 {
		return nil
	}

	_, okID := headers[string(ClusterID)]
	if !okID {
		return nil
	}

	// filter validate data
	for _, filter := range object.filter {
		if !filter.Filter(headers) {
			return nil
		}
	}

	glog.V(4).Infof("handler[%s] deal resourceType[%s] data", object.handler.Name(), object.resourceType)

	dataObject := &HandlerData{
		ResourceType: object.resourceType,
		Meta:         headers,
		Body:         data,
	}

	dataByte, err := json.Marshal(dataObject)
	if err != nil {
		glog.Errorf("marshal dataObject failed: %v", err)
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), HandleTimeout)
	defer cancel()

	err = object.handler.Handle(timeoutCtx, dataByte)
	if err != nil {
		glog.Errorf("external handler data failed: %v", err)
		return err
	}

	return nil
}
