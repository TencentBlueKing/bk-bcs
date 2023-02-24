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

package check

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

type listenerChecker struct {
	cli client.Client
}

func NewListenerChecker(cli client.Client) *listenerChecker {
	return &listenerChecker{
		cli: cli,
	}
}

func (l *listenerChecker) Run() {
	listenerList := &networkextensionv1.ListenerList{}
	if err := l.cli.List(context.TODO(), listenerList); err != nil {
		blog.Errorf("list listener failed, err: %s", err.Error())
		return
	}

	cntMap := make(map[string]int)
	for _, listener := range listenerList.Items {
		status := listener.Status.Status

		targetGroupType := networkextensionv1.LabelValueForTargetGroupNormal
		if listener.Spec.TargetGroup == nil || len(listener.Spec.TargetGroup.Backends) == 0 {
			targetGroupType = networkextensionv1.LabelValueForTargetGroupEmpty
		}

		cntMap[buildKey(status, targetGroupType)] = cntMap[buildKey(status, targetGroupType)] + 1

		label := listener.GetLabels()
		value, ok := label[networkextensionv1.LabelKetForTargetGroupType]
		if !ok || value != targetGroupType {
			patchStruct := map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]string{
						networkextensionv1.LabelKetForTargetGroupType: targetGroupType,
					},
				},
			}
			patchData, err := json.Marshal(patchStruct)
			if err != nil {
				blog.Errorf("marshal listener failed, err: %s", err.Error())
				continue
			}
			updatePod := &networkextensionv1.Listener{
				ObjectMeta: metav1.ObjectMeta{
					Name:      listener.Name,
					Namespace: listener.Namespace,
				},
			}
			err = l.cli.Patch(context.TODO(), updatePod, client.RawPatch(types.MergePatchType, patchData))
			if err != nil {
				blog.Errorf("patch listener failed, err: %s", err.Error())
				continue
			}
		}
	}

	metrics.ListenerTotal.Reset()
	for key, cnt := range cntMap {
		status, targetGroupType := transKey(key)
		metrics.ListenerTotal.WithLabelValues(status, targetGroupType).Set(float64(cnt))
	}
}

func buildKey(status, targetGroupType string) string {
	return fmt.Sprintf("%s/%s", status, targetGroupType)
}

// return status, targetGroup
func transKey(key string) (string, string) {
	splits := strings.Split(key, "/")
	if len(splits) != 2 {
		return "", ""
	}
	return splits[0], splits[1]
}
