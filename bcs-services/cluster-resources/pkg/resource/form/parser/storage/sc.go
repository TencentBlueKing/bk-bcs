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

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseSC ...
func ParseSC(manifest map[string]interface{}) map[string]interface{} {
	sc := model.SC{}
	common.ParseMetadata(manifest, &sc.Metadata)
	ParseSCSpec(manifest, &sc.Spec)
	return structs.Map(sc)
}

// ParseSCSpec ...
func ParseSCSpec(manifest map[string]interface{}, spec *model.SCSpec) {
	scDefaultFlagPath := []string{"metadata", "annotations", "storageclass.kubernetes.io/is-default-class"}
	if mapx.GetStr(manifest, scDefaultFlagPath) == "true" {
		spec.SetAsDefault = true
	}
	spec.Provisioner = mapx.GetStr(manifest, "provisioner")
	spec.VolumeBindingMode = mapx.GetStr(manifest, "volumeBindingMode")
	spec.ReclaimPolicy = mapx.GetStr(manifest, "reclaimPolicy")
	for key, value := range mapx.GetMap(manifest, "parameters") {
		spec.Params = append(spec.Params, model.SCParam{Key: key, Value: value.(string)})
	}
	for _, mOpt := range mapx.GetList(manifest, "mountOptions") {
		spec.MountOpts = append(spec.MountOpts, mOpt.(string))
	}
}
