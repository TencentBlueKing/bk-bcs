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

package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"k8s.io/client-go/util/retry"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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
// 上层为每个LB ID维护一个EventHandler， 处理该LB相关的所有监听器事件
// 当ListenerController收到Listener Event后，交由对应LB的EventHandler处理
// EventHandler将event存在Cache中，通过Ticker每三分钟处理一次 （为了使用batch接口提升处理效率）
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
	blog.Infof("listener '%s' was pushed in event queue", nsName.String())
	h.eventQueue.Forget(nsName)
	h.eventQueue.Add(nsName)
}

// handleQueue get listener from eventQueue and store in EventCache
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

	blog.Infof("listener '%s' got from event queue", nsName.String())

	listener := &networkextensionv1.Listener{}
	err := h.k8sCli.Get(context.Background(), nsName, listener)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Infof("not found listener '%s', forget", nsName.String())
			h.eventQueue.Forget(obj)
			h.eventQueue.Done(obj)
			return true
		}
		blog.Warnf("get listener of %s failed, err %s", nsName.String(), err.Error())
		h.eventQueue.Forget(obj)
		h.eventQueue.Done(obj)
		return true
	}

	blog.Infof("add listener %s/%s to processing cache", listener.GetName(), listener.GetNamespace())
	if listener.DeletionTimestamp != nil {
		deleteEvent := NewListenerEvent(
			EventDelete,
			listener.GetName(),
			listener.GetNamespace(),
		)
		h.eventRecvCache.Set(deleteEvent.Key(), deleteEvent)
	} else {
		updateEvent := NewListenerEvent(
			EventUpdate,
			listener.GetName(),
			listener.GetNamespace(),
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
			h.lbID, event.Type, event.Name, event.Namespace)
		// 实际处理时再获取listener，避免listener短时间内的多次变化被拆分到多次处理中
		listener := &networkextensionv1.Listener{}
		nsName := k8stypes.NamespacedName{
			Namespace: event.Namespace,
			Name:      event.Name,
		}
		if err := h.k8sCli.Get(context.TODO(), nsName, listener); err != nil {
			if k8serrors.IsNotFound(err) {
				blog.Infof("listener '%s' is deleted, skip", nsName.String())
				h.eventQueue.Forget(nsName)
				h.eventQueue.Done(nsName)
				continue
			}

			blog.Errorf("get listener '%s' failed, err: %s", nsName.String(), err.Error())
			h.eventQueue.AddRateLimited(nsName)
			h.eventQueue.Done(nsName)
			continue
		}

		// 根据事件类型区分处理
		switch event.Type {
		case EventAdd, EventUpdate:
			if listener.Spec.EndPort > 0 {
				segListenerEnsureList = append(segListenerEnsureList, listener)
				continue
			}
			listenerEnsureList = append(listenerEnsureList, listener)
		case EventDelete:
			listenerDeleteList = append(listenerDeleteList, listener)
		}
	}
	// 取出监听器后清空缓存
	h.eventRecvCache.Clean()
	h.queueLock.Unlock()

	if h.isBulkMode {
		if len(listenerDeleteList) > 0 {
			if err := h.deleteMultiListeners(listenerDeleteList); err != nil {
				blog.Warnf("delete multiple listeners failed, err %s", err.Error())
			}
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

	startTime := time.Now()
	var listenerIDMap map[string]cloud.Result
	var err error
	// 通过EndPort区分端口段监听器
	if listeners[0].Spec.EndPort > 0 {
		listenerIDMap, err = h.lbClient.EnsureMultiSegmentListeners(h.region, h.lbID, listeners)
	} else {
		listenerIDMap, err = h.lbClient.EnsureMultiListeners(h.region, h.lbID, listeners)
	}

	// 返回error时认为这批监听器全部处理失败
	if err != nil {
		for _, li := range listeners {
			obj := k8stypes.NamespacedName{
				Namespace: li.GetNamespace(),
				Name:      li.GetName(),
			}
			h.recordListenerFailedEvent(li, err)
			if inErr := h.patchListenerStatus(li, "", networkextensionv1.ListenerStatusNotSynced,
				err.Error()); inErr != nil {
				blog.Warnf("patch listener id of %s/%s failed, err %s", li.GetName(), li.GetNamespace(), inErr.Error())
			}
			metrics.ReportHandleListenerMetric(len(listeners), h.isBulkMode, metrics.ListenerMethodEnsureListener,
				err, startTime)
			h.eventQueue.AddRateLimited(obj)
			h.eventQueue.Done(obj)
		}
		return err
	}
	// 不发success事件，避免Listener量大时Event事件影响etcd性能
	for _, li := range listeners {
		obj := k8stypes.NamespacedName{
			Namespace: li.GetNamespace(),
			Name:      li.GetName(),
		}
		listenerResult, ok := listenerIDMap[li.GetName()]
		if !ok || listenerResult.IsError {
			blog.Warnf("ensure listener %s/%s failed, requeue", li.GetName(), li.GetNamespace())
			msg := "ensure multi listener failed"
			if ok {
				msg = listenerResult.Err.Error()
			}
			h.recordListenerFailedEvent(li, errors.New(msg))
			metrics.ReportHandleListenerMetric(len(listeners), h.isBulkMode, metrics.ListenerMethodEnsureListener,
				errors.New(msg), startTime)
			if inErr := h.patchListenerStatus(li, "", networkextensionv1.ListenerStatusNotSynced, msg); inErr != nil {
				blog.Warnf("patch listener id of %s/%s failed, err %s", li.GetName(), li.GetNamespace(), inErr.Error())
			}
			h.eventQueue.AddRateLimited(obj)
			h.eventQueue.Done(obj)
			continue
		}
		if inErr := h.patchListenerStatus(li, listenerResult.Res, networkextensionv1.ListenerStatusSynced,
			"multi ensure success"); inErr != nil {
			blog.Warnf("patch listener id of %s/%s failed, err %s", li.GetName(), li.GetNamespace(), inErr.Error())
			metrics.ReportHandleListenerMetric(len(listeners), h.isBulkMode, metrics.ListenerMethodEnsureListener,
				inErr, startTime)
			h.eventQueue.AddRateLimited(obj)
			h.eventQueue.Done(obj)
			continue
		}
		blog.V(3).Infof("ensure listener %s/%s from cloud successfully", li.GetName(), li.GetNamespace())
		metrics.ReportHandleListenerMetric(len(listeners), h.isBulkMode, metrics.ListenerMethodEnsureListener,
			nil, startTime)
		h.eventQueue.Forget(obj)
		h.eventQueue.Done(obj)
	}
	return nil
}

func (h *EventHandler) deleteMultiListeners(listeners []*networkextensionv1.Listener) error {
	if len(listeners) == 0 {
		return nil
	}

	startTime := time.Now()
	var protocolLayer string
	var err error
	switch listeners[0].Spec.Protocol {
	case constant.ProtocolUDP, constant.ProtocolTCP:
		protocolLayer = constant.ProtocolLayerTransport
	case constant.ProtocolHTTP, constant.ProtocolHTTPS:
		protocolLayer = constant.ProtocolLayerApplication
	}
	// 删除时判断LB是否存在,避免监听器删除接口调用失败，进而导致相关资源无法释放
	if h.lbClient.IsNamespaced() {
		_, err = h.lbClient.DescribeLoadBalancerWithNs(listeners[0].Namespace, h.region, h.lbID, "", protocolLayer)
	} else {
		_, err = h.lbClient.DescribeLoadBalancer(h.region, h.lbID, "", protocolLayer)
	}
	if err != nil && err != cloud.ErrLoadbalancerNotFound {
		blog.Errorf("cloud lb client DescribeLoadBalancer failed, err %s", err.Error())
		for _, li := range listeners {
			obj := k8stypes.NamespacedName{
				Namespace: li.GetNamespace(),
				Name:      li.GetName(),
			}
			h.recordListenerDeleteFailedEvent(li, err)
			metrics.ReportHandleListenerMetric(len(listeners), h.isBulkMode,
				metrics.ListenerMethodDeleteListener, err, startTime)
			h.eventQueue.AddRateLimited(obj)
			h.eventQueue.Done(obj)
		}
		return err
	}
	// 只有当LB未被删除时，才需要处理删除监听器
	if err == nil {
		err = h.lbClient.DeleteMultiListeners(h.region, h.lbID, listeners)
		if err != nil {
			blog.Warnf("delete listeners failed, requeue listeners, err: %s", err.Error())
			for _, li := range listeners {
				obj := k8stypes.NamespacedName{
					Namespace: li.GetNamespace(),
					Name:      li.GetName(),
				}
				h.recordListenerDeleteFailedEvent(li, err)
				metrics.ReportHandleListenerMetric(len(listeners), h.isBulkMode,
					metrics.ListenerMethodDeleteListener, err, startTime)
				h.eventQueue.AddRateLimited(obj)
				h.eventQueue.Done(obj)
			}
			return err
		}
	}

	for _, li := range listeners {
		obj := k8stypes.NamespacedName{
			Namespace: li.GetNamespace(),
			Name:      li.GetName(),
		}

		if err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			listener := &networkextensionv1.Listener{}
			if err = h.k8sCli.Get(context.TODO(), obj, listener); err != nil {
				if k8serrors.IsNotFound(err) {
					return nil
				}
				return err
			}
			cpListener := listener.DeepCopy()
			cpListener.Finalizers = common.RemoveString(cpListener.Finalizers, constant.FinalizerNameBcsIngressController)
			return h.k8sCli.Update(context.Background(), cpListener)
		}); err != nil {
			blog.Warnf("failed to remove finalizer from listener %s/%s, err %s",
				li.GetNamespace(), li.GetName(), err.Error())

			metrics.ReportHandleListenerMetric(len(listeners), h.isBulkMode,
				metrics.ListenerMethodDeleteListener, err, startTime)
			h.eventQueue.AddRateLimited(obj)
			h.eventQueue.Done(obj)
			continue
		}

		metrics.ReportHandleListenerMetric(len(listeners), h.isBulkMode,
			metrics.ListenerMethodDeleteListener, nil, startTime)
		blog.V(3).Infof("delete listener %s/%s from cloud successfully", li.GetName(), li.GetNamespace())
		h.eventQueue.Forget(obj)
		h.eventQueue.Done(obj)
	}
	return nil
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
	startTime := time.Now()
	var listenerID string
	var err error
	// 通过EndPort区分端口段监听器
	if li.Spec.EndPort > 0 {
		listenerID, err = h.lbClient.EnsureSegmentListener(h.region, li)
		if err != nil {
			h.recordListenerFailedEvent(li, err)
			metrics.ReportHandleListenerMetric(1, h.isBulkMode, metrics.ListenerMethodEnsureListener, err, startTime)
			blog.Errorf("cloud lb client EnsureSegmentListener failed, err %s", err.Error())
			if inErr := h.patchListenerStatus(li, "", networkextensionv1.ListenerStatusNotSynced,
				err.Error()); inErr != nil {
				blog.Warnf("patch listener id of %s/%s failed, err %s", li.GetName(), li.GetNamespace(), inErr.Error())
			}
			return fmt.Errorf("cloud lb client EnsureSegmentListener failed, err %s", err.Error())
		}
	} else {
		listenerID, err = h.lbClient.EnsureListener(h.region, li)
		if err != nil {
			h.recordListenerFailedEvent(li, err)
			metrics.ReportHandleListenerMetric(1, h.isBulkMode, metrics.ListenerMethodEnsureListener, err, startTime)
			blog.Errorf("cloud lb client EnsureListener failed, err %s", err.Error())
			if inErr := h.patchListenerStatus(li, "", networkextensionv1.ListenerStatusNotSynced,
				err.Error()); inErr != nil {
				blog.Warnf("patch listener id of %s/%s failed, err %s", li.GetName(), li.GetNamespace(), inErr.Error())
			}
			return fmt.Errorf("cloud lb client EnsureListener failed, err %s", err.Error())
		}
	}
	if err = h.patchListenerStatus(li, listenerID, networkextensionv1.ListenerStatusSynced,
		"ensure success"); err != nil {
		metrics.ReportHandleListenerMetric(1, h.isBulkMode, metrics.ListenerMethodEnsureListener, err, startTime)
		blog.Errorf("patch listener id %s to k8s apiserver failed, err %s", listenerID, err.Error())
		return fmt.Errorf("update listener id %s to k8s apiserver failed, err %s", listenerID, err.Error())
	}
	metrics.ReportHandleListenerMetric(1, h.isBulkMode, metrics.ListenerMethodEnsureListener, nil, startTime)
	return nil
}

func (h *EventHandler) deleteListener(li *networkextensionv1.Listener) error {
	startTime := time.Now()
	var protocolLayer string
	var err error
	switch li.Spec.Protocol {
	case constant.ProtocolUDP, constant.ProtocolTCP:
		protocolLayer = constant.ProtocolLayerTransport
	case constant.ProtocolHTTP, constant.ProtocolHTTPS:
		protocolLayer = constant.ProtocolLayerApplication
	}
	// 删除时判断LB是否存在,避免监听器删除接口调用失败，进而导致相关资源无法释放
	if h.lbClient.IsNamespaced() {
		_, err = h.lbClient.DescribeLoadBalancerWithNs(li.Namespace, h.region, h.lbID, "", protocolLayer)
	} else {
		_, err = h.lbClient.DescribeLoadBalancer(h.region, h.lbID, "", protocolLayer)
	}
	if err != nil && err != cloud.ErrLoadbalancerNotFound {
		h.recordListenerDeleteFailedEvent(li, err)
		metrics.ReportHandleListenerMetric(1, h.isBulkMode, metrics.ListenerMethodDeleteListener, err, startTime)
		blog.Errorf("cloud lb client DescribeLoadBalancer failed, err %s", err.Error())
		return fmt.Errorf("cloud lb client DescribeLoadBalancer failed, err %s", err.Error())
	}
	// 只有当LB未被删除时，才需要处理删除监听器
	if err == nil {
		if li.Spec.EndPort > 0 {
			err = h.lbClient.DeleteSegmentListener(h.region, li)
			if err != nil {
				h.recordListenerDeleteFailedEvent(li, err)
				metrics.ReportHandleListenerMetric(1, h.isBulkMode, metrics.ListenerMethodDeleteListener, err, startTime)
				blog.Errorf("cloud lb client DeleteSegmentListener failed, err %s", err.Error())
				return fmt.Errorf("cloud lb client DeleteSegmentListener failed, err %s", err.Error())
			}
		} else {
			err = h.lbClient.DeleteListener(h.region, li)
			if err != nil {
				h.recordListenerDeleteFailedEvent(li, err)
				metrics.ReportHandleListenerMetric(1, h.isBulkMode, metrics.ListenerMethodDeleteListener, err, startTime)
				blog.Errorf("cloud lb client DeleteListener failed, err %s", err.Error())
				return fmt.Errorf("cloud lb client DeleteListener failed, err %s", err.Error())
			}
		}
	}

	if err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		listener := &networkextensionv1.Listener{}
		if err = h.k8sCli.Get(context.TODO(), k8stypes.NamespacedName{Namespace: li.GetNamespace(), Name: li.GetName()},
			listener); err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		cpListener := listener.DeepCopy()
		cpListener.Finalizers = common.RemoveString(cpListener.Finalizers, constant.FinalizerNameBcsIngressController)
		return h.k8sCli.Update(context.Background(), cpListener)
	}); err != nil {
		blog.Errorf("failed to remove finalizer from listener %s/%s, err %s",
			li.GetNamespace(), li.GetName(), err.Error())
		metrics.ReportHandleListenerMetric(1, h.isBulkMode, metrics.ListenerMethodDeleteListener, err, startTime)
		return fmt.Errorf("failed to remove finalizer from listener %s/%s, err %s",
			li.GetNamespace(), li.GetName(), err.Error())
	}

	metrics.ReportHandleListenerMetric(1, h.isBulkMode, metrics.ListenerMethodDeleteListener, nil, startTime)
	return nil
}

// Stop stop event handler
func (h *EventHandler) Stop() {
	h.eventQueue.ShutDown()
	h.ctx.Done()
}
