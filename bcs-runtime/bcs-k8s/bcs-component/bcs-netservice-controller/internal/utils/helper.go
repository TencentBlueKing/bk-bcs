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

package utils

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	netservicev1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/internal/constant"
)

// FixActiveIP 存在某些场景，pod已经删除，但是 bcsnetip 状态没有改变
func FixActiveIP(cli client.Client, netIP *netservicev1.BCSNetIP) error {
	if netIP.Status.Phase != constant.BCSNetIPActiveStatus {
		return nil
	}
	if netIP.Status.IPClaimKey == "" || netIP.Status.PodName == "" || netIP.Status.PodNamespace == "" {
		return fmt.Errorf("IP %s is in Active status, but not bounded", netIP.Name)
	}

	// check pod
	pod := &v1.Pod{}
	err := cli.Get(context.Background(), types.NamespacedName{Name: netIP.Status.PodName,
		Namespace: netIP.Status.PodNamespace}, pod)
	if err == nil {
		return nil
	}
	if !k8serrors.IsNotFound(err) {
		return err
	}

	newNetIP := netIP.DeepCopy()

	// pod is not exist, release IP
	newNetIP.Status.Phase = constant.BCSNetIPReservedStatus
	newNetIP.Status.Host = ""
	newNetIP.Status.ContainerID = ""
	newNetIP.Status.PodNamespace = ""
	newNetIP.Status.PodName = ""
	newNetIP.Status.UpdateTime = metav1.Now()
	if err := cli.Status().Update(context.Background(), newNetIP); err != nil {
		blog.Errorf("update BCSNetIP %s status failed, err %v", newNetIP.Name, err)
		return fmt.Errorf("update BCSNetIP %s status failed, err %v", newNetIP.Name, err)
	}

	return nil
}
