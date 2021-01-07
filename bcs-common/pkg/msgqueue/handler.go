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
	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/micro/go-micro/v2/broker"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"time"
)

// consumer subscribe resource type
const (
	PodSubscribeType         = "Pod"
	EventSubscribeType       = "Event"
	DeploymentSubscribeType  = "Deployment"
	StatefulSetSubscribeType = "StatefulSet"
)

// default handler context timeout
const (
	HandleTimeout = 10 * time.Second
)

// Handler handle event to data and inject consumer
type Handler interface {
	Name() string
	Handle(ctx context.Context, data interface{}) error
}

// HandlerWrap function for Handler interface
func HandlerWrap(name string, f func(ctx context.Context, data interface{}) error) *HandlerWrapper {
	return &HandlerWrapper{name, f}
}

// HandlerWrapper for Handler
type HandlerWrapper struct {
	NameValue string
	Impl      func(ctx context.Context, data interface{}) error
}

// Handle of hw
func (hw *HandlerWrapper) Handle(ctx context.Context, data interface{}) error {
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

	glog.Infof("handler[%s] deal resourceType[%s] data", object.handler.Name(), object.resourceType)

	var dataObject interface{}
	switch object.resourceType {
	case PodSubscribeType:
		dataObject = corev1.Pod{}
	case EventSubscribeType:
		dataObject = corev1.Event{}
	case DeploymentSubscribeType:
		dataObject = appsv1.Deployment{}
	case StatefulSetSubscribeType:
		dataObject = appsv1.StatefulSet{}
	default:
		return nil
	}

	err := json.Unmarshal(data, &dataObject)
	if err != nil {
		glog.Infof("unmarshal pod data failed: %v", err)
		return nil
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), HandleTimeout)
	defer cancel()

	err = object.handler.Handle(timeoutCtx, dataObject)
	if err != nil {
		glog.Errorf("external handler data failed: %v", err)
		return err
	}

	return nil
}
