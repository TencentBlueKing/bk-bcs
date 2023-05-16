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

// ParsePV ...
func ParsePV(manifest map[string]interface{}) map[string]interface{} {
	pv := model.PV{}
	common.ParseMetadata(manifest, &pv.Metadata)
	ParsePVSpec(manifest, &pv.Spec)
	return structs.Map(pv)
}

// ParsePVSpec ...
func ParsePVSpec(manifest map[string]interface{}, spec *model.PVSpec) {
	spec.SCName = mapx.GetStr(manifest, "spec.storageClassName")
	spec.StorageSize = util.ConvertStorageUnit(mapx.GetStr(manifest, "spec.capacity.storage"))
	if accessModes := mapx.GetList(manifest, "spec.accessModes"); len(accessModes) != 0 {
		for _, am := range accessModes {
			spec.AccessModes = append(spec.AccessModes, am.(string))
		}
	}
	if local := mapx.GetMap(manifest, "spec.local"); len(local) != 0 {
		spec.Type = resCsts.PVTypeLocalVolume
		spec.LocalPath = mapx.GetStr(local, "path")
	} else if hp := mapx.GetMap(manifest, "spec.hostPath"); len(hp) != 0 {
		spec.Type = resCsts.PVTypeHostPath
		spec.HostPath = mapx.GetStr(hp, "path")
		spec.HostPathType = mapx.GetStr(hp, "type")
	} else if nfs := mapx.GetMap(manifest, "spec.nfs"); len(nfs) != 0 {
		spec.Type = resCsts.PVTypeNFS
		spec.NFSPath = mapx.GetStr(nfs, "path")
		spec.NFSServer = mapx.GetStr(nfs, "server")
		spec.NFSReadOnly = mapx.GetBool(nfs, "readOnly")
	}
}
