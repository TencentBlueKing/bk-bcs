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

// Package formatter xxx
package formatter

import (
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// FormatCRD ...
func FormatCRD(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonFormatRes(manifest)
	ret["name"] = mapx.GetStr(manifest, "metadata.name")
	ret["scope"] = mapx.Get(manifest, "spec.scope", "N/A")
	ret["kind"] = mapx.Get(manifest, "spec.names.kind", "N/A")
	ret["apiVersion"] = parseCObjAPIVersion(manifest)
	ret["addColumns"] = parseCRDAdditionalColumns(manifest)
	return ret
}

// FormatCObj ...
func FormatCObj(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonFormatRes(manifest)
	ret["podLabelSelector"] = parsePodLabelSelector(manifest)
	return ret
}

// FormatGWorkload ...
func FormatGWorkload(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonFormatRes(manifest)
	ret["images"] = parseContainerImages(manifest, "spec.template.spec.containers")
	ret["podLabelSelector"] = parsePodLabelSelector(manifest)
	return ret
}

// FormatGDeploy ...
func FormatGDeploy(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatGWorkload(manifest)
	// 注：前端针对自定义资源，会根据 AdditionalColumns 自动获取字段数据并展示，
	// 如果使用 status 字段，则会冲突导致数据覆盖（普通资源不会存在该问题）
	ret["resStatus"] = newGDeployStatusParser(manifest).Parse()
	return ret
}

// FormatGSTS ...
func FormatGSTS(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatGWorkload(manifest)
	ret["resStatus"] = newGSTSStatusParser(manifest).Parse()
	return ret
}

// GDeployStatusChecker GameDeployment 状态检查器
type GDeployStatusChecker struct{}

// IsNormal ...
func (c *GDeployStatusChecker) IsNormal(manifest map[string]interface{}) bool {
	return slice.AllInt64Equal([]int64{
		mapx.GetInt64(manifest, "spec.replicas"),
		mapx.GetInt64(manifest, "status.readyReplicas"),
		mapx.GetInt64(manifest, "status.updatedReplicas"),
		mapx.GetInt64(manifest, "status.updatedReadyReplicas"),
	})
}

// GSTSStatusChecker GameStatefulSet 状态检查器
type GSTSStatusChecker struct{}

// IsNormal ...
func (c *GSTSStatusChecker) IsNormal(manifest map[string]interface{}) bool {
	return slice.AllInt64Equal([]int64{
		mapx.GetInt64(manifest, "spec.replicas"),
		mapx.GetInt64(manifest, "status.readyReplicas"),
		mapx.GetInt64(manifest, "status.currentReplicas"),
		mapx.GetInt64(manifest, "status.updatedReplicas"),
		mapx.GetInt64(manifest, "status.updatedReadyReplicas"),
	})
}

func newGDeployStatusParser(manifest map[string]interface{}) *WorkloadStatusParser {
	return &WorkloadStatusParser{&GDeployStatusChecker{}, manifest}
}

func newGSTSStatusParser(manifest map[string]interface{}) *WorkloadStatusParser {
	return &WorkloadStatusParser{&GSTSStatusChecker{}, manifest}
}

// parseCObjAPIVersion 根据 CRD 配置解析 cobj ApiVersion
// nolint
// ref: https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/#specify-multiple-versions
func parseCObjAPIVersion(manifest map[string]interface{}) string {
	group := mapx.GetStr(manifest, "spec.group")
	versions := mapx.GetList(manifest, "spec.versions")

	if len(versions) != 0 {
		for _, ver := range versions {
			v, _ := ver.(map[string]interface{})
			if v["served"].(bool) {
				return fmt.Sprintf("%s/%s", group, v["name"])
			}
		}
		return fmt.Sprintf("%s/%s", group, versions[0].(map[string]interface{})["name"])
	}

	version := mapx.GetStr(manifest, "spec.version")
	if version != "" {
		return fmt.Sprintf("%s/%s", group, version)
	}
	return fmt.Sprintf("%s/v1alpha1", group)
}

// parseCRDAdditionalColumns xxx
func parseCRDAdditionalColumns(manifest map[string]interface{}) (addColumns []interface{}) {
	// CRD v1beta1 spec.additionalPrinterColumns
	rawAddColumns := mapx.GetList(manifest, "spec.additionalPrinterColumns")
	if len(rawAddColumns) == 0 {
		// CRD v1 spec.versions[0].additionalPrinterColumns
		versions := mapx.GetList(manifest, "spec.versions")
		if len(versions) != 0 {
			rawAddColumns = mapx.GetList(versions[0].(map[string]interface{}), "additionalPrinterColumns")
		}
	}
	for _, column := range rawAddColumns {
		col, _ := column.(map[string]interface{})
		// 存在时间会统一处理，因此此处直接过滤掉
		if strings.ToLower(col["name"].(string)) == "age" {
			continue
		}
		// BCS 不同版本 CRD jsonPath 参数 Key 值不一致，有 jsonPath, JSONPath 等多个版本
		// 前端统一使用 jsonPath，因此这里做一次检查，若 jsonPath 不存在，则赋予 JSONPath 的值
		if _, exists := col["jsonPath"]; !exists {
			col["jsonPath"] = col["JSONPath"]
		}
		addColumns = append(addColumns, col)
	}
	return addColumns
}

// parsePodLabelSelector 获取 Pod Label Selector
func parsePodLabelSelector(manifest map[string]interface{}) string {
	labels := mapx.GetMap(manifest, "spec.template.metadata.labels")
	for k, v := range labels {
		if k == constants.PodLabelSelector {
			if value, ok := v.(string); ok {
				return fmt.Sprintf("%s=%s", constants.PodLabelSelector, value)
			}
		}
	}
	return ""
}
