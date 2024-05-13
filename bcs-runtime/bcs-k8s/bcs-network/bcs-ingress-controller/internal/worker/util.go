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
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/pkg/errors"

	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ReasonEnsureListenerSuccess event reason field for ensuring listener successfully
	ReasonEnsureListenerSuccess = "ensure success"
	// ReasonEnsureListenerDeleteSuccess event reason field for delete listener successfully
	ReasonEnsureListenerDeleteSuccess = "delete success"
	// ReasonEnsureListenerFailed event reason field for ensuring listener failed
	ReasonEnsureListenerFailed = "ensure fail"
	// ReasonEnsureListenerDeleteFailed event reason field for delete listener failed
	ReasonEnsureListenerDeleteFailed = "delete fail"
	// ReasonBackendUnhealthy event reason for listener unhealthy backends
	ReasonBackendUnhealthy = "unhealthy listener backends"
	// MsgEnsureListenerSuccess msg ensure listener successfully
	MsgEnsureListenerSuccess = "ensured success, listener id %s"
	// MsgEnsureListenerFailed msg ensure listener failed
	MsgEnsureListenerFailed = "ensure falied, err %s"
	// MsgEnsureListenerDeleteSuccess msg delete listener successfully
	MsgEnsureListenerDeleteSuccess = "delete success, listener id %s"
	// MsgEnsureListenerDeleteFailed msg delete listener failed
	MsgEnsureListenerDeleteFailed = "delete falied, err %s"
	// MsgBackendUnhealthy msg which show listener backend is unhealthy
	MsgBackendUnhealthy = "listener %s, port %d has unhealthy backend %+v"
)

func (h *EventHandler) recordListenerEvent(lis *networkextensionv1.Listener, eType, reason, msg string) {
	if h.listenerEventer == nil {
		return
	}
	h.listenerEventer.Event(lis, eType, reason, msg)
}

func (h *EventHandler) recordListenerSuccessEvent(lis *networkextensionv1.Listener, lid string) {
	h.recordListenerEvent(lis, k8scorev1.EventTypeNormal, ReasonEnsureListenerSuccess,
		fmt.Sprintf(MsgEnsureListenerSuccess, lid))
}

func (h *EventHandler) recordListenerFailedEvent(lis *networkextensionv1.Listener, err error) {
	h.recordListenerEvent(lis, k8scorev1.EventTypeWarning, ReasonEnsureListenerFailed,
		fmt.Sprintf(MsgEnsureListenerFailed, err.Error()))
}

func (h *EventHandler) recordListenerDeleteSuccessEvent(lis *networkextensionv1.Listener, lid string) {
	h.recordListenerEvent(lis, k8scorev1.EventTypeNormal, ReasonEnsureListenerDeleteSuccess,
		fmt.Sprintf(MsgEnsureListenerDeleteSuccess, lid))
}

func (h *EventHandler) recordListenerDeleteFailedEvent(lis *networkextensionv1.Listener, err error) {
	h.recordListenerEvent(lis, k8scorev1.EventTypeWarning, ReasonEnsureListenerDeleteFailed,
		fmt.Sprintf(MsgEnsureListenerDeleteFailed, err.Error()))
}

func (h *EventHandler) recordBackendUnhealthyEvent(lis *networkextensionv1.Listener, port int, backends []string) {
	h.recordListenerEvent(lis, k8scorev1.EventTypeWarning, ReasonBackendUnhealthy,
		fmt.Sprintf(MsgBackendUnhealthy, lis.GetName(), port, backends))
}

func (h *EventHandler) patchListenerStatus(lis *networkextensionv1.Listener, lid string, status string, msg string) error {
	listenerStatus := &networkextensionv1.ListenerStatus{
		ListenerID: lid,
		Status:     status,
		Msg:        msg,
	}
	patchStruct := map[string]interface{}{
		"status": listenerStatus,
	}
	patchBytes, err := json.Marshal(patchStruct)
	if err != nil {
		blog.Errorf("marshal listener status failed, lid: %s, status: %s, msg: %s, err: %s", lid, status, msg, err)
		return errors.Wrapf(err, "marshal listener status failed, lid: %s, status: %s, msg: %s", lid, status, msg)
	}
	rawPatch := client.RawPatch(k8stypes.MergePatchType, patchBytes)
	updateListener := &networkextensionv1.Listener{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      lis.GetName(),
			Namespace: lis.GetNamespace(),
		},
	}
	err = h.k8sCli.Patch(context.Background(), updateListener, rawPatch, &client.PatchOptions{})
	if err != nil {
		blog.Errorf("patch listener id %s to k8s apiserver failed, err %s", lid, err.Error())
		return errors.Wrapf(err, "update listener id %s to k8s apiserver failed", lid)
	}
	return nil
}
