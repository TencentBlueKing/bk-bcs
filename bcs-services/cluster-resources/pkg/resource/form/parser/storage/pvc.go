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

package storage

import (
	"github.com/fatih/structs"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/util"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParsePVC ...
func ParsePVC(manifest map[string]interface{}) map[string]interface{} {
	pvc := model.PVC{}
	common.ParseMetadata(manifest, &pvc.Metadata)
	ParsePVCSpec(manifest, &pvc.Spec)
	return structs.Map(pvc)
}

// ParsePVCSpec ...
func ParsePVCSpec(manifest map[string]interface{}, spec *model.PVCSpec) {
	spec.ClaimType = resCsts.PVCTypeUseExistPV
	spec.PVName = mapx.GetStr(manifest, "spec.volumeName")
	// 如果没有指定 PVName，则认为是要根据 StorageClass 创建
	if spec.PVName == "" {
		spec.ClaimType = resCsts.PVCTypeCreateBySC
	}

	spec.SCName = mapx.GetStr(manifest, "spec.storageClassName")
	spec.StorageSize = util.ConvertStorageUnit(mapx.GetStr(manifest, "spec.resources.requests.storage"))
	if accessModes := mapx.GetList(manifest, "spec.accessModes"); len(accessModes) != 0 {
		for _, am := range accessModes {
			spec.AccessModes = append(spec.AccessModes, am.(string))
		}
	}
}
