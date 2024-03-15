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
	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PatchNodeAnnotation patch annotation to node
func PatchNodeAnnotation(ctx context.Context, k8sCli client.Client, node *k8scorev1.Node,
	patchAnno map[string]interface{}) error {
	patchStruct := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": patchAnno,
		},
	}
	patchBytes, err := json.Marshal(patchStruct)
	if err != nil {
		return errors.Wrapf(err, "marshal patchAnno for node '%s' failed", node.Name)
	}
	rawPatch := client.RawPatch(k8stypes.MergePatchType, patchBytes)
	updateNode := &k8scorev1.Node{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name: node.Name,
		},
	}
	if err = k8sCli.Patch(ctx, updateNode, rawPatch, &client.PatchOptions{}); err != nil {
		return errors.Wrapf(err, "patch node '%s' annotation failed, patcheStruct: %s", node.Name, string(patchBytes))
	}

	return nil
}
