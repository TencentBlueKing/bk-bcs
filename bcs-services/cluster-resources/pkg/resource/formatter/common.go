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

package formatter

import (
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/timex"
)

// CommonFormatRes 通用资源格式化
func CommonFormatRes(manifest map[string]interface{}) map[string]interface{} {
	rawCreateTime := mapx.GetStr(manifest, "metadata.creationTimestamp")
	createSource, immutable := parseCreateSource(manifest)
	ret := map[string]interface{}{
		"namespace":  mapx.GetStr(manifest, []string{"metadata", "namespace"}),
		"age":        timex.CalcAge(rawCreateTime),
		"createTime": rawCreateTime,
		"editMode": mapx.Get(
			manifest, []string{"metadata", "annotations", resCsts.EditModeAnnoKey}, resCsts.EditModeYaml,
		),
		"creator":         mapx.GetStr(manifest, []string{"metadata", "annotations", resCsts.CreatorAnnoKey}),
		"updater":         mapx.GetStr(manifest, []string{"metadata", "annotations", resCsts.UpdaterAnnoKey}),
		"immutable":       immutable,
		"createSource":    createSource,
		"templateName":    mapx.GetStr(manifest, []string{"metadata", "annotations", resCsts.TemplateNameAnnoKey}),
		"templateVersion": mapx.GetStr(manifest, []string{"metadata", "annotations", resCsts.TemplateVersionAnnoKey}),
		"chart":           mapx.GetStr(manifest, []string{"metadata", "labels", resCsts.HelmChartAnnoKey}),
	}
	return ret
}

// GetFormatFunc 获取资源对应 FormatFunc
func GetFormatFunc(kind string, apiVersion string) func(manifest map[string]interface{}) map[string]interface{} {
	// 自定义Ingress，按照通用资源格式化
	if kind == resCsts.Ing && apiVersion == resCsts.BCSNetworkApiVersion {
		kind = ""
	}
	formatFunc, ok := Kind2FormatFuncMap[kind]
	if !ok {
		// 若指定资源类型没有对应的，则当作自定义资源处理
		return FormatCObj
	}
	return formatFunc
}

// GetPruneFunc 获取资源对应 PruneFunc
func GetPruneFunc(kind string) func(manifest map[string]interface{}) map[string]interface{} {
	pruneFunc, ok := Kind2PruneFuncMap[kind]
	if !ok {
		return DefaultPruneFunc
	}
	return pruneFunc
}

// 解析创建来源，主要有：Template/Helm/Client/Web
func parseCreateSource(manifest map[string]interface{}) (string, bool) {
	labels := mapx.GetMap(manifest, "metadata.labels")
	// Helm创建来源：app.kubernetes.io/managed-by: Helm
	if mapx.GetStr(labels, []string{resCsts.HelmSourceType}) == resCsts.HelmCreateSource {
		return resCsts.HelmCreateSource, true
	}

	annotations := mapx.GetMap(manifest, "metadata.annotations")
	// template创建来源：io.tencent.paas.source_type: template
	if mapx.GetStr(annotations, []string{resCsts.TemplateSourceType}) == resCsts.TemplateSourceTypeValue ||
		mapx.GetStr(labels, []string{resCsts.TemplateSourceType}) == resCsts.TemplateSourceTypeValue {
		return resCsts.TemplateCreateSource, false
	}

	// web 创建来源
	if mapx.GetStr(annotations, []string{resCsts.CreatorAnnoKey}) != "" ||
		mapx.GetStr(labels, []string{resCsts.UpdaterAnnoKey}) != "" {
		return resCsts.WebCreateSource, false
	}

	return resCsts.ClientCreateSource, false
}
