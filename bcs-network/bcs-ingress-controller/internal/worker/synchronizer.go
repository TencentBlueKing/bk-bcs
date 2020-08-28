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
	"time"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/cloud"
)

// EventHandler handler for listener event
type EventHandler struct {
	ctx context.Context

	lbID   string
	region string

	lbClient cloud.LoadBalance

	k8sCli client.Client

	eventRecvCache  *EventCache
	eventDoingCache *EventCache

	needHandle bool
	isDoing    bool
}

// NewEventHandler create event handler
func NewEventHandler(region, lbID string, lbClient cloud.LoadBalance, k8sCli client.Client) *EventHandler {
	return &EventHandler{
		ctx:             context.Background(),
		lbID:            lbID,
		region:          region,
		lbClient:        lbClient,
		k8sCli:          k8sCli,
		eventRecvCache:  NewEventCache(),
		eventDoingCache: NewEventCache(),
	}
}

// PushEvent push event to event handler
func (h *EventHandler) PushEvent(e *ListenerEvent) {
	h.eventRecvCache.Set(e.Key(), e)
	h.needHandle = true
}

// Run run event handler loop
func (h *EventHandler) Run() {
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			if !h.needHandle || h.isDoing {
				continue
			}
			h.needHandle = false
			h.isDoing = true
			// if has error
			if h.doHandle() {
				h.needHandle = true
				time.Sleep(2 * time.Second)
			}
			h.isDoing = false
		case <-h.ctx.Done():
			blog.Infof("EventHandler for %s run loop exit", h.lbID)
			return
		}
	}
}

func (h *EventHandler) doHandle() bool {
	h.eventRecvCache.Drain(h.eventDoingCache)
	hasError := false
	for _, event := range h.eventDoingCache.List() {
		blog.Infof("[worker %s] eventType: %s, listener: %s/%s",
			h.lbID, event.Type, event.Listener.GetName(), event.Listener.GetNamespace())
		switch event.Type {
		case EventAdd, EventUpdate:
			err := h.ensureListener(&event)
			if err != nil {
				blog.Warnf("ensureListener listener event %s failed", event)
				hasError = true
				continue
			}
			h.eventDoingCache.Delete(event.Key())
		case EventDelete:
			err := h.deleteListener(&event)
			if err != nil {
				blog.Warnf("deleteListener listener event %s failed", event)
				hasError = true
				continue
			}
			h.eventDoingCache.Delete(event.Key())
		default:
			blog.Warnf("[worker %s] invalid evenType: %s, listener: %s/%s",
				h.lbID, event.Type, event.Listener.GetName(), event.Listener.GetNamespace())
		}
	}
	return hasError
}

func (h *EventHandler) ensureListener(e *ListenerEvent) error {
	var listenerID string
	var err error
	if e.Listener.Spec.EndPort > 0 {
		listenerID, err = h.lbClient.EnsureSegmentListener(h.region, &e.Listener)
		if err != nil {
			blog.Errorf("cloud lb client EnsureSegmentListener failed, err %s", err.Error())
			return fmt.Errorf("cloud lb client EnsureSegmentListener failed, err %s", err.Error())
		}
	} else {
		listenerID, err = h.lbClient.EnsureListener(h.region, &e.Listener)
		if err != nil {
			blog.Errorf("cloud lb client EnsureListener failed, err %s", err.Error())
			return fmt.Errorf("cloud lb client EnsureListener failed, err %s", err.Error())
		}
	}

	rawPatch := client.RawPatch(k8stypes.MergePatchType, []byte("{\"status\":{\"listenerID\":\""+listenerID+"\"}}"))
	updateListener := &networkextensionv1.Listener{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      e.Listener.Name,
			Namespace: e.Listener.Namespace,
		},
	}
	err = h.k8sCli.Patch(context.TODO(), updateListener, rawPatch, &client.PatchOptions{})
	if err != nil {
		blog.Errorf("patch listener id %s to k8s apiserver failed, err %s", listenerID, err.Error())
		return fmt.Errorf("update listener id %s to k8s apiserver failed, err %s", listenerID, err.Error())
	}
	return nil
}

func (h *EventHandler) deleteListener(e *ListenerEvent) error {
	var err error
	if e.Listener.Spec.EndPort > 0 {
		err = h.lbClient.DeleteSegmentListener(h.region, &e.Listener)
		if err != nil {
			blog.Errorf("cloud lb client DeleteSegmentListener failed, err %s", err.Error())
			return fmt.Errorf("cloud lb client DeleteSegmentListener failed, err %s", err.Error())
		}
	} else {
		err = h.lbClient.DeleteListener(h.region, &e.Listener)
		if err != nil {
			blog.Errorf("cloud lb client DeleteListener failed, err %s", err.Error())
			return fmt.Errorf("cloud lb client DeleteListener failed, err %s", err.Error())
		}
	}
	return nil
}

// Stop stop event handler
func (h *EventHandler) Stop() {
	h.ctx.Done()
}
