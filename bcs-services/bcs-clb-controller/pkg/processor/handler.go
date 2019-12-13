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
	"bk-bcs/bcs-common/common/blog"
	ingressType "bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/clb/v1"
	cloudListenerType "bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1"
	"reflect"
)

type appServiceHandler struct {
	processor *Processor
}

func newAppServiceHandler() *appServiceHandler {
	return &appServiceHandler{}
}

func (handler *appServiceHandler) RegisterProcessor(p *Processor) {
	handler.processor = p
}

func (handler *appServiceHandler) OnAdd(obj interface{}) {
	handler.processor.SetUpdated()
}

func (handler *appServiceHandler) OnUpdate(objOld, objNew interface{}) {
	handler.processor.SetUpdated()
}

func (handler *appServiceHandler) OnDelete(obj interface{}) {
	handler.processor.SetUpdated()
}

type ingressHandler struct {
	processor *Processor
}

func newIngressHandler() *ingressHandler {
	return &ingressHandler{}
}

func (handler *ingressHandler) RegisterProcessor(p *Processor) {
	handler.processor = p
}

func (handler *ingressHandler) OnAdd(obj interface{}) {
	ingress, ok := obj.(*ingressType.ClbIngress)
	if ok {
		blog.V(5).Infof("sync clb ingress add event: %s", ingress.ToString())
	} else {
		blog.Errorf("get object add %v, no a clb ingress object", obj)
		return
	}
	handler.processor.SetUpdated()
}

func (handler *ingressHandler) OnUpdate(objOld, objNew interface{}) {
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

func (handler *ingressHandler) OnDelete(obj interface{}) {
	ingress, ok := obj.(*ingressType.ClbIngress)
	if ok {
		blog.V(5).Infof("sync clb ingress delete event: %s", ingress.ToString())
	} else {
		blog.Errorf("get object delete %v, no a clb ingress object", obj)
		return
	}
	handler.processor.SetUpdated()
}

type nodeHandler struct {
	processor *Processor
}

func newNodeHandler() *nodeHandler {
	return &nodeHandler{}
}

func (handler *nodeHandler) RegisterProcessor(p *Processor) {
	handler.processor = p
}

func (handler *nodeHandler) OnAdd(obj interface{}) {
	handler.processor.SetUpdated()
}

func (handler *nodeHandler) OnUpdate(objOld, objNew interface{}) {
	handler.processor.SetUpdated()
}

func (handler *nodeHandler) OnDelete(obj interface{}) {
	handler.processor.SetUpdated()
}

type listenerHandler struct{}

func newListenerHandler() *listenerHandler {
	return &listenerHandler{}
}

func (handler *listenerHandler) OnAdd(obj interface{}) {
	listener, ok := obj.(*cloudListenerType.CloudListener)
	if ok {
		blog.V(5).Infof("sync listener add event: %s", listener.ToString())
	} else {
		blog.Errorf("get object add %v, no a listener object", obj)
	}
}

func (handler *listenerHandler) OnUpdate(objOld, objNew interface{}) {
	listenerNew, okNew := objNew.(*cloudListenerType.CloudListener)
	listenerOld, okOld := objOld.(*cloudListenerType.CloudListener)
	if okNew && okOld {
		blog.V(5).Infof("sync listener update event: %s, old %s", listenerNew.ToString(), listenerOld.ToString())
	} else {
		blog.Errorf("get object update %v, old %v, no a listener object", objNew, objOld)
	}
}

func (handler *listenerHandler) OnDelete(obj interface{}) {
	listener, ok := obj.(*cloudListenerType.CloudListener)
	if ok {
		blog.V(5).Infof("sync listener delete event: %s", listener.ToString())
	} else {
		blog.Errorf("get object delete %v, no a listener object", obj)
	}
}
