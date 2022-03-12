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

package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	eventQueueBackoffBaseDuration = 1 * time.Second
	eventQueueBackoffMaxDuration  = 5 * time.Minute
)

// EventHandler handler for listener event
type EventHandler struct {
	ctx context.Context

	lbID   string
	region string

	lbClient cloud.LoadBalance

	k8sCli client.Client

	listenerEventer record.EventRecorder

	eventRecvCache *EventCache
	queueLock      sync.Mutex
	eventQueue     workqueue.RateLimitingInterface

	// if use bulk cloud interface
	isBulkMode bool
}

// NewEventHandler create event handler
func NewEventHandler(opt EventHandlerOption) *EventHandler {
	return &EventHandler{
		ctx:             context.Background(),
		lbID:            opt.LbID,
		region:          opt.Region,
		lbClient:        opt.LbClient,
		k8sCli:          opt.K8sCli,
		isBulkMode:      opt.IsBulkMode,
		listenerEventer: opt.ListenerEventer,
		eventRecvCache:  NewEventCache(),
		// does a simple eventQueueBackoffBaseDuration*2^<failures> limit
		eventQueue: workqueue.NewRateLimitingQueue(
			workqueue.NewItemExponentialFailureRateLimiter(
				eventQueueBackoffBaseDuration,
				eventQueueBackoffMaxDuration)),
	}
}

// PushQueue push item into queue
func (h *EventHandler) PushQueue(nsName k8stypes.NamespacedName) {
	h.eventQueue.Forget(nsName)
	h.eventQueue.Add(nsName)
}

func (h *EventHandler) handleQueue() bool {
	// CAUTION: Done() function must be called at last for object Get() from eventQueue
	obj, shutDown := h.eventQueue.Get()
	if shutDown {
		blog.Warnf("event queue of lb %s was shut down", h.lbID)
		return false
	}
	nsName, ok := obj.(k8stypes.NamespacedName)
	if !ok {
		h.eventQueue.Forget(obj)
		h.eventQueue.Done(obj)
		blog.Warnf("invalid queue item, type %T, value %v", obj, obj)
		return true
	}
	h.queueLock.Lock()
	defer h.queueLock.Unlock()

	listener := &networkextensionv1.Listener{}
	err := h.k8sCli.Get(context.Background(), nsName, listener)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			h.eventQueue.Forget(obj)
			h.eventQueue.Done(obj)
			return true
		}
		blog.Warnf("get listener of %s failed, err %s", nsName.String(), err.Error())
		h.eventQueue.Forget(obj)
		h.eventQueue.Done(obj)
		return true
	}

	copiedListener := listener.DeepCopy()
	blog.Infof("add listener %s/%s to processing cache", copiedListener.GetName(), copiedListener.GetNamespace())
	if copiedListener.DeletionTimestamp != nil {
		deleteEvent := NewListenerEvent(
			EventDelete,
			copiedListener.GetName(),
			copiedListener.GetNamespace(),
			copiedListener,
		)
		h.eventRecvCache.Set(deleteEvent.Key(), deleteEvent)
	} else {
		updateEvent := NewListenerEvent(
			EventUpdate,
			copiedListener.GetName(),
			copiedListener.GetNamespace(),
			copiedListener,
		)
		h.eventRecvCache.Set(updateEvent.Key(), updateEvent)
	}
	return true
}

func (h *EventHandler) doHandleMulti() error {
	h.queueLock.Lock()
	var listenerEnsureList []*networkextensionv1.Listener
	var listenerDeleteList []*networkextensionv1.Listener
	var segListenerEnsureList []*networkextensionv1.Listener
	for _, event := range h.eventRecvCache.List() {
		blog.Infof("[worker %s] eventType: %s, listener: %s/%s",
			h.lbID, event.Type, event.Listener.GetName(), event.Listener.GetNamespace())
		switch event.Type {
		case EventAdd, EventUpdate:
			if event.Listener.Spec.EndPort > 0 {
				segListenerEnsureList = append(segListenerEnsureList, event.Listener)
				continue
			}
			listenerEnsureList = append(listenerEnsureList, event.Listener)
		case EventDelete:
			listenerDeleteList = append(listenerDeleteList, event.Listener)
		}
	}
	h.eventRecvCache.Clean()
	h.queueLock.Unlock()

	if h.isBulkMode {
		if len(listenerDeleteList) > 0 {
			h.deleteMultiListeners(listenerDeleteList)
		}
		if len(segListenerEnsureList) > 0 {
			if err := h.ensureMultiListeners(segListenerEnsureList); err != nil {
				blog.Warnf("ensure multiple segment listeners failed, err %s", err.Error())
			}
		}
		if len(listenerEnsureList) > 0 {
			if err := h.ensureMultiListeners(listenerEnsureList); err != nil {
				blog.Warnf("ensure multiple listeners failed, err %s", err.Error())
			}
		}
	} else {
		for _, delLi := range listenerDeleteList {
			obj := k8stypes.NamespacedName{
				Namespace: delLi.GetNamespace(),
				Name:      delLi.GetName(),
			}
			if err := h.deleteListener(delLi); err != nil {
				blog.Warnf("delete listener %s failed, requeue later", obj.String())
				h.eventQueue.AddRateLimited(obj)
				h.eventQueue.Done(obj)
			} else {
				blog.V(3).Infof("delete listener %s successfully", obj.String())
				h.eventQueue.Forget(obj)
				h.eventQueue.Done(obj)
			}
		}
		for _, segLi := range segListenerEnsureList {
			obj := k8stypes.NamespacedName{
				Namespace: segLi.GetNamespace(),
				Name:      segLi.GetName(),
			}
			if err := h.ensureListener(segLi); err != nil {
				blog.Warnf("ensure segment listener %s failed, requeue later", obj.String())
				h.eventQueue.AddRateLimited(obj)
				h.eventQueue.Done(obj)
			} else {
				blog.V(3).Infof("ensure segment listener %s successfully", obj.String())
				h.eventQueue.Forget(obj)
				h.eventQueue.Done(obj)
			}
		}
		for _, li := range listenerEnsureList {
			obj := k8stypes.NamespacedName{
				Namespace: li.GetNamespace(),
				Name:      li.GetName(),
			}
			if err := h.ensureListener(li); err != nil {
				blog.Warnf("ensure listener %s failed, requeue later", obj.String())
				h.eventQueue.AddRateLimited(obj)
				h.eventQueue.Done(obj)
			} else {
				blog.V(3).Infof("ensure listener %s successfully", obj.String())
				h.eventQueue.Forget(obj)
				h.eventQueue.Done(obj)
			}
		}
	}

	return nil
}

func (h *EventHandler) ensureMultiListeners(listeners []*networkextensionv1.Listener) error {
	if len(listeners) == 0 {
		return fmt.Errorf("ensureMultiListeners listener list cannot be empty")
	}
	var listenerIDMap map[string]string
	var err error
	if listeners[0].Spec.EndPort > 0 {
		listenerIDMap, err = h.lbClient.EnsureMultiSegmentListeners(h.region, h.lbID, listeners)
	} else {
		listenerIDMap, err = h.lbClient.EnsureMultiListeners(h.region, h.lbID, listeners)
	}

	if err != nil {
		for _, li := range listeners {
			obj := k8stypes.NamespacedName{
				Namespace: li.GetNamespace(),
				Name:      li.GetName(),
			}
			h.eventQueue.AddRateLimited(obj)
			h.eventQueue.Done(obj)
		}
		return err
	}
	for _, li := range listeners {
		obj := k8stypes.NamespacedName{
			Namespace: li.GetNamespace(),
			Name:      li.GetName(),
		}
		listenerID, ok := listenerIDMap[li.GetName()]
		if !ok {
			blog.Warnf("ensure listener %s/%s failed, requeue", li.GetName(), li.GetNamespace())
			h.eventQueue.AddRateLimited(obj)
			h.eventQueue.Done(obj)
			continue
		}
		h.recordListenerSuccessEvent(li, listenerID)
		if err := h.patchListenerID(li, listenerID); err != nil {
			blog.Warnf("patch listener id of %s/%s failed, err %s", li.GetName(), li.GetNamespace(), err.Error())
			h.eventQueue.AddRateLimited(obj)
			h.eventQueue.Done(obj)
			continue
		}
		blog.V(3).Infof("ensure listener %s/%s from cloud successfully", li.GetName(), li.GetNamespace())
		h.eventQueue.Forget(obj)
		h.eventQueue.Done(obj)
	}
	return nil
}

func (h *EventHandler) deleteMultiListeners(listeners []*networkextensionv1.Listener) {
	err := h.lbClient.DeleteMultiListeners(h.region, h.lbID, listeners)
	if err != nil {
		blog.Warnf("delete listeners failed, requeue listeners")
		for _, li := range listeners {
			h.eventQueue.AddRateLimited(k8stypes.NamespacedName{
				Namespace: li.GetNamespace(),
				Name:      li.GetName(),
			})
		}
		return
	}
	for _, li := range listeners {
		li.Finalizers = common.RemoveString(li.Finalizers, constant.FinalizerNameBcsIngressController)
		err := h.k8sCli.Update(context.Background(), li, &client.UpdateOptions{})
		if err != nil {
			blog.Warnf("failed to remove finalizer from listener %s/%s, err %s",
				li.GetNamespace(), li.GetName(), err.Error())
			h.eventQueue.AddRateLimited(k8stypes.NamespacedName{
				Namespace: li.GetNamespace(),
				Name:      li.GetName(),
			})
		}
	}
	for _, li := range listeners {
		obj := k8stypes.NamespacedName{
			Namespace: li.GetNamespace(),
			Name:      li.GetName(),
		}
		blog.V(3).Infof("delete listener %s/%s from cloud successfully", li.GetName(), li.GetNamespace())
		h.eventQueue.Forget(obj)
		h.eventQueue.Done(obj)
	}
}

// RunQueueRecving run queue receiver
func (h *EventHandler) RunQueueRecving() {
	for h.handleQueue() {
	}
	blog.Infof("queue is shut down")
}

// Run run event handler loop
func (h *EventHandler) Run() {
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C:
			if err := h.doHandleMulti(); err != nil {
				blog.Warnf("handle failed, err %s", err.Error())
			}
		case <-h.ctx.Done():
			blog.Infof("EventHandler for %s run loop exit", h.lbID)
			return
		}
	}
}

func (h *EventHandler) ensureListener(li *networkextensionv1.Listener) error {
	var listenerID string
	var err error
	if li.Spec.EndPort > 0 {
		listenerID, err = h.lbClient.EnsureSegmentListener(h.region, li)
		if err != nil {
			h.recordListenerFailedEvent(li, err)
			blog.Errorf("cloud lb client EnsureSegmentListener failed, err %s", err.Error())
			return fmt.Errorf("cloud lb client EnsureSegmentListener failed, err %s", err.Error())
		}
	} else {
		listenerID, err = h.lbClient.EnsureListener(h.region, li)
		if err != nil {
			h.recordListenerFailedEvent(li, err)
			blog.Errorf("cloud lb client EnsureListener failed, err %s", err.Error())
			return fmt.Errorf("cloud lb client EnsureListener failed, err %s", err.Error())
		}
	}
	h.recordListenerSuccessEvent(li, listenerID)

	rawPatch := client.RawPatch(k8stypes.MergePatchType, []byte("{\"status\":{\"listenerID\":\""+listenerID+"\"}}"))
	updateListener := &networkextensionv1.Listener{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      li.GetName(),
			Namespace: li.GetNamespace(),
		},
	}
	err = h.k8sCli.Patch(context.TODO(), updateListener, rawPatch, &client.PatchOptions{})
	if err != nil {
		blog.Errorf("patch listener id %s to k8s apiserver failed, err %s", listenerID, err.Error())
		return fmt.Errorf("update listener id %s to k8s apiserver failed, err %s", listenerID, err.Error())
	}
	return nil
}

func (h *EventHandler) deleteListener(li *networkextensionv1.Listener) error {
	var err error
	if li.Spec.EndPort > 0 {
		err = h.lbClient.DeleteSegmentListener(h.region, li)
		if err != nil {
			blog.Errorf("cloud lb client DeleteSegmentListener failed, err %s", err.Error())
			return fmt.Errorf("cloud lb client DeleteSegmentListener failed, err %s", err.Error())
		}
	} else {
		err = h.lbClient.DeleteListener(h.region, li)
		if err != nil {
			blog.Errorf("cloud lb client DeleteListener failed, err %s", err.Error())
			return fmt.Errorf("cloud lb client DeleteListener failed, err %s", err.Error())
		}
	}
	li.Finalizers = common.RemoveString(li.Finalizers, constant.FinalizerNameBcsIngressController)
	err = h.k8sCli.Update(context.Background(), li, &client.UpdateOptions{})
	if err != nil {
		blog.Errorf("failed to remove finalizer from listener %s/%s, err %s",
			li.GetNamespace(), li.GetName(), err.Error())
		return fmt.Errorf("failed to remove finalizer from listener %s/%s, err %s",
			li.GetNamespace(), li.GetName(), err.Error())
	}
	return nil
}

// Stop stop event handler
func (h *EventHandler) Stop() {
	h.eventQueue.ShutDown()
	h.ctx.Done()
}
