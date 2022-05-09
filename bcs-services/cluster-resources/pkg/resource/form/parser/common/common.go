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

package common

import (
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseAPIVersionKind 解析 apiVersion && kind
func ParseAPIVersionKind(manifest map[string]interface{}) (string, string) {
	return mapx.GetStr(manifest, "apiVersion"), mapx.GetStr(manifest, "kind")
}

// ParseMetadata ...
func ParseMetadata(manifest map[string]interface{}, metadata *model.Metadata) {
	metadata.Name = mapx.GetStr(manifest, "metadata.name")
	metadata.Namespace = mapx.GetStr(manifest, "metadata.namespace")
	ParseLabels(manifest, &metadata.Labels)
	ParseAnnotations(manifest, &metadata.Annotations)
}

// ParseLabels ...
func ParseLabels(manifest map[string]interface{}, labels *[]model.Label) {
	if ls, _ := mapx.GetItems(manifest, "metadata.labels"); ls != nil {
		for k, v := range ls.(map[string]interface{}) {
			*labels = append(*labels, model.Label{Key: k, Value: v.(string)})
		}
	}
}

// ParseAnnotations ...
func ParseAnnotations(manifest map[string]interface{}, annotations *[]model.Annotation) {
	if as, _ := mapx.GetItems(manifest, "metadata.annotations"); as != nil {
		for k, v := range as.(map[string]interface{}) {
			*annotations = append(*annotations, model.Annotation{Key: k, Value: v.(string)})
		}
	}
}
