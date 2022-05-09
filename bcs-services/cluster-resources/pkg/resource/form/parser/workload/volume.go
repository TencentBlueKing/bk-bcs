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

package workload

import (
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseWorkloadVolume ...
func ParseWorkloadVolume(manifest map[string]interface{}, volume *model.WorkloadVolume) {
	if volumes, _ := mapx.GetItems(manifest, "spec.template.spec.volumes"); volumes != nil { // nolint:nestif
		for _, vol := range volumes.([]interface{}) {
			v, _ := vol.(map[string]interface{})
			if _, ok := v["configMap"]; ok {
				volume.ConfigMap = append(volume.ConfigMap, model.CMVolume{
					Name:        v["name"].(string),
					DefaultMode: mapx.GetInt64(v, "configMap.defaultMode"),
					CMName:      mapx.GetStr(v, "configMap.name"),
					Items:       parseVolumeItems(v, "configMap.items"),
				})
			} else if _, ok := v["secret"]; ok {
				volume.Secret = append(volume.Secret, model.SecretVolume{
					Name:        v["name"].(string),
					DefaultMode: mapx.GetInt64(v, "secret.defaultMode"),
					SecretName:  mapx.GetStr(v, "secret.secretName"),
					Items:       parseVolumeItems(v, "secret.items"),
				})
			} else if _, ok := v["hostPath"]; ok {
				volume.HostPath = append(volume.HostPath, model.HostPathVolume{
					Name: v["name"].(string),
					Path: mapx.GetStr(v, "hostPath.path"),
					Type: mapx.GetStr(v, "hostPath.type"),
				})
			} else if _, ok := v["persistentVolumeClaim"]; ok {
				volume.PVC = append(volume.PVC, model.PVCVolume{
					Name:     v["name"].(string),
					PVCName:  mapx.GetStr(v, "persistentVolumeClaim.claimName"),
					ReadOnly: mapx.GetBool(v, "persistentVolumeClaim.readOnly"),
				})
			} else if _, ok := v["emptyDir"]; ok {
				volume.EmptyDir = append(volume.EmptyDir, model.EmptyDirVolume{
					Name: v["name"].(string),
				})
			} else if _, ok := v["nfs"]; ok {
				volume.NFS = append(volume.NFS, model.NFSVolume{
					Name:     v["name"].(string),
					Path:     mapx.GetStr(v, "nfs.path"),
					Server:   mapx.GetStr(v, "nfs.server"),
					ReadOnly: mapx.GetBool(v, "nfs.readOnly"),
				})
			}
		}
	}
}

// parseVolumeItems 解析 ConfigMap/SecretVolume Key-Path 信息
func parseVolumeItems(vol map[string]interface{}, paths string) []model.KeyToPath {
	items := []model.KeyToPath{}
	for _, item := range mapx.GetList(vol, paths) {
		it, _ := item.(map[string]interface{})
		items = append(items, model.KeyToPath{Key: it["key"].(string), Path: it["path"].(string)})
	}
	return items
}
