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

package processor

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	ingressType "github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/clb/v1"
	cloudListenerType "github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1"
	"reflect"
)

type HandlerProcessor interface {
	SetUpdated()
}

type AppServiceHandler struct {
	processor HandlerProcessor
}

func NewAppServiceHandler() *AppServiceHandler {
	return &AppServiceHandler{}
}

func (handler *AppServiceHandler) RegisterProcessor(p HandlerProcessor) {
	handler.processor = p
}

func (handler *AppServiceHandler) OnAdd(obj interface{}) {
	handler.processor.SetUpdated()
}

func (handler *AppServiceHandler) OnUpdate(objOld, objNew interface{}) {
	handler.processor.SetUpdated()
}

func (handler *AppServiceHandler) OnDelete(obj interface{}) {
	handler.processor.SetUpdated()
}

type IngressHandler struct {
	processor HandlerProcessor
}

func NewIngressHandler() *IngressHandler {
	return &IngressHandler{}
}

func (handler *IngressHandler) RegisterProcessor(p HandlerProcessor) {
	handler.processor = p
}

func (handler *IngressHandler) OnAdd(obj interface{}) {
	ingress, ok := obj.(*ingressType.ClbIngress)
	if ok {
		blog.V(5).Infof("sync clb ingress add event: %s", ingress.ToString())
	} else {
		blog.Errorf("get object add %v, no a clb ingress object", obj)
		return
	}
	handler.processor.SetUpdated()
}

func (handler *IngressHandler) OnUpdate(objOld, objNew interface{}) {
	ingressNew, okNew := objNew.(*ingressType.ClbIngress)
	ingressOld, okOld := objOld.(*ingressType.ClbIngress)
	if okNew && okOld {
		blog.V(5).Infof("sync clb ingress update event: %s, old %s", ingressNew.ToString(), ingressOld.ToString())
	} else {
		blog.Errorf("get object update %v, old %v, no a listener object", objNew, objOld)
		return
	}
	if reflect.DeepEqual(ingressNew.Spec, ingressOld.Spec) {
		blog.V(5).Infof("clb ingress new %s has no change, no need to call updater", ingressNew.ToString())
		return
	}
	handler.processor.SetUpdated()
}

func (handler *IngressHandler) OnDelete(obj interface{}) {
	ingress, ok := obj.(*ingressType.ClbIngress)
	if ok {
		blog.V(5).Infof("sync clb ingress delete event: %s", ingress.ToString())
	} else {
		blog.Errorf("get object delete %v, no a clb ingress object", obj)
		return
	}
	handler.processor.SetUpdated()
}

type NodeHandler struct {
	processor HandlerProcessor
}

func NewNodeHandler() *NodeHandler {
	return &NodeHandler{}
}

func (handler *NodeHandler) RegisterProcessor(p HandlerProcessor) {
	handler.processor = p
}

func (handler *NodeHandler) OnAdd(obj interface{}) {
	handler.processor.SetUpdated()
}

func (handler *NodeHandler) OnUpdate(objOld, objNew interface{}) {
	handler.processor.SetUpdated()
}

func (handler *NodeHandler) OnDelete(obj interface{}) {
	handler.processor.SetUpdated()
}

type ListenerHandler struct{}

func NewListenerHandler() *ListenerHandler {
	return &ListenerHandler{}
}

func (handler *ListenerHandler) OnAdd(obj interface{}) {
	listener, ok := obj.(*cloudListenerType.CloudListener)
	if ok {
		blog.V(5).Infof("sync listener add event: %s", listener.ToString())
	} else {
		blog.Errorf("get object add %v, no a listener object", obj)
	}
}

func (handler *ListenerHandler) OnUpdate(objOld, objNew interface{}) {
	listenerNew, okNew := objNew.(*cloudListenerType.CloudListener)
	listenerOld, okOld := objOld.(*cloudListenerType.CloudListener)
	if okNew && okOld {
		blog.V(5).Infof("sync listener update event: %s, old %s", listenerNew.ToString(), listenerOld.ToString())
	} else {
		blog.Errorf("get object update %v, old %v, no a listener object", objNew, objOld)
	}
}

func (handler *ListenerHandler) OnDelete(obj interface{}) {
	listener, ok := obj.(*cloudListenerType.CloudListener)
	if ok {
		blog.V(5).Infof("sync listener delete event: %s", listener.ToString())
	} else {
		blog.Errorf("get object delete %v, no a listener object", obj)
	}
}
