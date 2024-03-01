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

package ingresscontroller

import (
	"fmt"
	"reflect"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8scorev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/ingresscache"
)

func deduplicateIngresses(ingresses []ingresscache.IngressMeta) []ingresscache.IngressMeta {
	var retList []ingresscache.IngressMeta
	ingressMap := make(map[string]struct{})
	for index, meta := range ingresses {
		key := fmt.Sprintf("%s/%s", meta.Namespace, meta.Name)
		if _, ok := ingressMap[key]; !ok {
			ingressMap[key] = struct{}{}
			retList = append(retList, ingresses[index])
		}
	}
	return retList
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

	if !reflect.DeepEqual(oldPod.Labels, newPod.Labels) {
		return true
	}

	if !reflect.DeepEqual(oldPod.Status, newPod.Status) {
		return true
	}

	if !reflect.DeepEqual(oldPod.Spec, newPod.Spec) {
		return true
	}

	if oldPod.Annotations == nil || newPod.Annotations == nil {
		return true
	}
	if oldPod.Annotations[networkextensionv1.AnnotationKeyForLoadbalanceWeight] != newPod.
		Annotations[networkextensionv1.AnnotationKeyForLoadbalanceWeight] {
		return true
	}

	return false
}
