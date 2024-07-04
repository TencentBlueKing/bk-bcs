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

package webhookserver

import (
	"context"
	"fmt"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/utils"
	netextv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func (s *Server) validateDeletePortPool(pool *netextv1.PortPool) error {
	for _, itemStatus := range pool.Status.PoolItemStatuses {
		// check whether there is port bind object related to this port pool item
		set := k8slabels.Set(map[string]string{
			utils.GenPortBindingLabel(pool.GetName(), pool.GetNamespace()): itemStatus.ItemName,
		})
		selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(set))
		if err != nil {
			return fmt.Errorf("get selector from set %v failed, err %s", set, err.Error())
		}
		bindingList := &netextv1.PortBindingList{}
		if err = s.k8sClient.List(
			context.Background(), bindingList, &client.ListOptions{LabelSelector: selector}); err != nil {
			return fmt.Errorf("failed to list port bind list, err %s", err.Error())
		}
		if len(bindingList.Items) != 0 {
			return fmt.Errorf("port binding object found! cannot delete port pool item %s of pool %s/%s",
				itemStatus.ItemName, pool.GetNamespace(), pool.GetName())
		}
	}

	return nil
}
