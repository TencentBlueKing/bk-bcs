/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package web

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenPVCMountTips(t *testing.T) {
	ctx := context.TODO()
	pvcMountInfo := map[string][]string{
		"pvc-1": {"po-11", "po-12", "po-13"},
		"pvc-2": {"po-21", "po-22"},
	}
	assert.Equal(
		t, "无法删除 PersistentVolumeClaim，原因：已经被 po-11 等共计 3 个 Pod 挂载",
		genPVCMountTips(ctx, "pvc-1", pvcMountInfo),
	)
	assert.Equal(
		t, "无法删除 PersistentVolumeClaim，原因：已被 Pods [po-21 po-22] 挂载",
		genPVCMountTips(ctx, "pvc-2", pvcMountInfo),
	)
	assert.Equal(t, "", genPVCMountTips(ctx, "pvc-3", pvcMountInfo))
}
