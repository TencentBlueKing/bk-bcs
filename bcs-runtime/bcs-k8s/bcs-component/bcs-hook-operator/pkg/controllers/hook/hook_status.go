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

package hook

import (
	"context"
	utildiff "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/util/diff"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	patchtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

func (hc *HookController) updateHookRunStatus(orig *v1alpha1.HookRun, newStatus v1alpha1.HookRunStatus) error {
	patch, modified, err := utildiff.CreateTwoWayMergePatch(
		&v1alpha1.HookRun{
			Status: orig.Status,
		},
		&v1alpha1.HookRun{
			Status: newStatus,
		}, v1alpha1.HookRun{})
	if err != nil {
		klog.Errorf("HookRun %s/%s: Error constructing HookRun status patch: %v", orig.Namespace, orig.Name, err)
		return err
	}
	if !modified {
		klog.Infof("HookRun %s/%s: No status changes. Skipping patch", orig.Namespace, orig.Name)
		return nil
	}
	klog.Infof("HookRun %s/%s Patch: %s", orig.Namespace, orig.Name, patch)
	_, err = hc.tkexClient.TkexV1alpha1().HookRuns(orig.Namespace).Patch(context.TODO(),
		orig.Name, patchtypes.MergePatchType, patch, metav1.PatchOptions{}, "status")
	if err != nil {
		klog.Warningf("HookRun %s/%s: error updating HookRun: %v", orig.Namespace, orig.Name, err)
		return err
	}
	klog.Infof("HookRun %s/%s: Patch status successfully", orig.Namespace, orig.Name)
	return nil
}
