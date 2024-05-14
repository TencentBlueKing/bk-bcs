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
	"encoding/json"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	monitorextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
)

// PatchAppMonitorAnnotation patch annotation to app monitor
func PatchAppMonitorAnnotation(ctx context.Context, k8sCli client.Client,
	monitor *monitorextensionv1.AppMonitor, patchAnno map[string]interface{}) error {
	patchStruct := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": patchAnno,
		},
	}
	patchBytes, err := json.Marshal(patchStruct)
	if err != nil {
		return errors.Wrapf(err, "marshal patchStruct for app monitor '%s/%s' failed", monitor.GetNamespace(),
			monitor.GetName())
	}
	rawPatch := client.RawPatch(k8stypes.MergePatchType, patchBytes)
	updateAppMonitor := &monitorextensionv1.AppMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      monitor.GetName(),
			Namespace: monitor.GetNamespace(),
		},
	}
	if inErr := k8sCli.Patch(ctx, updateAppMonitor, rawPatch, &client.PatchOptions{}); inErr != nil {
		return errors.Wrapf(err, "patch app monitor %s/%s annotation failed, patcheStruct: %s",
			monitor.GetNamespace(), monitor.GetName(), string(patchBytes))
	}
	return nil
}
