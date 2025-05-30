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

// Package config xxx
package config

import (
	"github.com/fatih/structs"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseBscpConfig BscpConfig manifest -> formData
func ParseBscpConfig(manifest map[string]interface{}) map[string]interface{} {
	cm := model.BscpConfig{}
	common.ParseMetadata(manifest, &cm.Metadata)
	ParseBscpConfigSpec(manifest, &cm.Spec)
	return structs.Map(cm)
}

// ParseBscpConfigSpec ...
func ParseBscpConfigSpec(manifest map[string]interface{}, spec *model.BscpConfigSpec) {
	configSyncer := mapx.GetList(manifest, "spec.configSyncer")
	for _, v := range configSyncer {
		if vv, ok := v.(map[string]interface{}); ok {
			cs := model.ConfigSyncer{
				ConfigmapName:    mapx.GetStr(vv, "configmapName"),
				AssociationRules: "matchConfigs",
				ResourceType:     "configmap",
				SecretName:       mapx.GetStr(vv, "secretName"),
				SecretType:       mapx.GetStr(vv, "type"),
			}
			// 默认值
			if cs.SecretType == "" {
				cs.SecretType = "Opaque"
			}
			if mapx.GetStr(vv, "configmapName") == "" {
				cs.ResourceType = "secret"
			}
			if len(mapx.GetList(vv, "data")) != 0 {
				cs.AssociationRules = "data"
			}
			for _, mc := range mapx.GetList(vv, "matchConfigs") {
				if mcV, ok := mc.(string); ok {
					cs.MatchConfigs = append(cs.MatchConfigs, model.MatchConfigs{Value: mcV})
				}
			}
			for _, dataV := range mapx.GetList(vv, "data") {
				if dv, ok := dataV.(map[string]interface{}); ok {
					cs.ConfigData = append(cs.ConfigData, model.ConfigSyncerData{
						Key:       mapx.GetStr(dv, "key"),
						RefConfig: mapx.GetStr(dv, "refConfig"),
					})
				}
			}
			spec.ConfigSyncer = append(spec.ConfigSyncer, cs)
		}
	}

	spec.Provider.App = mapx.GetStr(manifest, "spec.provider.app")
	spec.Provider.Biz = mapx.GetInt64(manifest, "spec.provider.biz")
	spec.Provider.FeedAddr = mapx.GetStr(manifest, "spec.provider.feedAddr")
	spec.Provider.Token = mapx.GetStr(manifest, "spec.provider.token")
}
