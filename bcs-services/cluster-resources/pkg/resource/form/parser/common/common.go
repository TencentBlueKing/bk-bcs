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

// Package common xxx
package common

import (
	"sort"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseMetadata xxx
func ParseMetadata(manifest map[string]interface{}, metadata *model.Metadata) {
	metadata.APIVersion = mapx.GetStr(manifest, "apiVersion")
	metadata.Kind = mapx.GetStr(manifest, "kind")
	metadata.Name = mapx.GetStr(manifest, "metadata.name")
	metadata.Namespace = mapx.GetStr(manifest, "metadata.namespace")
	lables, _ := mapx.GetItems(manifest, "metadata.labels")
	ParseLabels(lables, &metadata.Labels)
	ParseAnnotations(manifest, &metadata.Annotations)
	metadata.ResVersion = mapx.GetStr(manifest, "metadata.resourceVersion")
}

// ParseLabels xxx
func ParseLabels(manifest interface{}, labels *[]model.Label) {
	if manifest == nil {
		return
	}
	for k, v := range manifest.(map[string]interface{}) {
		*labels = append(*labels, model.Label{Key: k, Value: v.(string)})
	}
	if labels == nil || len(*labels) == 0 {
		return
	}
	sort.Slice(*labels, func(i, j int) bool {
		return (*labels)[i].Key < (*labels)[j].Key
	})
}

// ParseAnnotations xxx
func ParseAnnotations(manifest map[string]interface{}, annotations *[]model.Annotation) {
	if as, _ := mapx.GetItems(manifest, "metadata.annotations"); as != nil {
		for k, v := range as.(map[string]interface{}) {
			*annotations = append(*annotations, model.Annotation{Key: k, Value: v.(string)})
		}
	}
	if annotations == nil || len(*annotations) == 0 {
		return
	}
	sort.Slice(*annotations, func(i, j int) bool {
		return (*annotations)[i].Key < (*annotations)[j].Key
	})
}
