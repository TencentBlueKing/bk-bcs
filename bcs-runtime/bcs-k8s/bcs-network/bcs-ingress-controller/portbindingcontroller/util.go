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

package portbindingcontroller

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/pkg/errors"
	k8scorev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	netpkgcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
)

// if port binding is expired
func isPortBindingExpired(portBinding *networkextensionv1.PortBinding) (bool, error) {
	keepTimeStr, ok := portBinding.Annotations[networkextensionv1.PortPoolBindingAnnotationKeyKeepDuration]
	if !ok {
		// always return true when no keeptime annotation
		return true, nil
	}
	keepDuration, err := time.ParseDuration(keepTimeStr)
	if err != nil {
		return false, fmt.Errorf("parse keep duration string %s failed, err %s", keepTimeStr, err.Error())
	}
	updateTime, err := netpkgcommon.ParseTimeString(portBinding.Status.UpdateTime)
	if err != nil {
		return false, fmt.Errorf("parse update time string %s failed, err %s",
			portBinding.Status.UpdateTime, err.Error())
	}
	if time.Now().After(updateTime.Add(keepDuration)) {
		return true, nil
	}
	return false, nil
}

// check for pod annotation
func checkPortPoolAnnotation(annotations map[string]string) bool {
	_, ok := annotations[constant.AnnotationForPortPool]
	return ok
}

// return portBindingItemList in pod annotation
func parsePoolBindingsAnnotation(pod *k8scorev1.Pod) ([]*networkextensionv1.PortBindingItem, error) {
	poolBindingItemList := make([]*networkextensionv1.PortBindingItem, 0)
	poolBindingItemsStr, ok := pod.Annotations[constant.AnnotationForPortPoolBindings]
	if !ok {
		return poolBindingItemList, nil
	}
	if err := json.Unmarshal([]byte(poolBindingItemsStr), &poolBindingItemList); err != nil {
		return poolBindingItemList, errors.Wrapf(err, "unmarshal annotation[%s]='%s' for pod '%s/%s' failed",
			constant.AnnotationForPortPoolBindings, poolBindingItemsStr, pod.Namespace, pod.Name)
	}

	return poolBindingItemList, nil
}

// return unique ID of portBindingItem
func genUniqueIDOfPortBindingItem(item *networkextensionv1.PortBindingItem) string {
	if item == nil {
		return ""
	}

	return fmt.Sprintf("%s/%s/%s", item.PoolNamespace, item.PoolName, item.GetKey())
}

func checkPodNeedReconcile(oldPod, newPod *k8scorev1.Pod) bool {
	if oldPod == nil || newPod == nil {
		return true
	}
	if oldPod.Namespace != newPod.Namespace || oldPod.Name != newPod.Name {
		return true
	}

	if oldPod.DeletionTimestamp != newPod.DeletionTimestamp {
		return true
	}

	if oldPod.Annotations == nil || newPod.Annotations == nil {
		return true
	}

	// 允许用户通过删除该注解实现Pod解绑
	if oldPod.Annotations[constant.AnnotationForPortPool] != newPod.Annotations[constant.
		AnnotationForPortPool] {
		return true
	}
	// 允许用户更新PortBinding保留时间
	if oldPod.Annotations[networkextensionv1.PortPoolBindingAnnotationKeyKeepDuration] != newPod.
		Annotations[networkextensionv1.PortPoolBindingAnnotationKeyKeepDuration] {
		return true
	}

	// Pod状态/IP等变化时需要触发PortBinding调谐
	if !reflect.DeepEqual(oldPod.Status, newPod.Status) {
		return true
	}

	if !reflect.DeepEqual(oldPod.Spec, newPod.Spec) {
		return true
	}

	return false
}
