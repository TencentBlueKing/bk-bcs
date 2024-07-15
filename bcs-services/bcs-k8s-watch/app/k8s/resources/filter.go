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

package resources

import (
	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/options"
)

// ResourceFilter filters unwanted resources or specifies wanted resources
type ResourceFilter struct {
	filterConfig    *options.FilterConfig
	blackListFilter map[string]map[string]struct{}
	whiteListFilter map[string]map[string]struct{}
}

// NewResourceFilter create resource filter
func NewResourceFilter(filterConfig *options.FilterConfig) *ResourceFilter {
	blackListFilter := make(map[string]map[string]struct{})
	whiteListFilter := make(map[string]map[string]struct{})
	if filterConfig != nil {
		for _, gv := range filterConfig.APIResourceException {
			blackListFilter[gv.GroupVersion] = make(map[string]struct{})
			for _, resource := range gv.ResourceKinds {
				blackListFilter[gv.GroupVersion][resource] = struct{}{}
			}
		}
		for _, gv := range filterConfig.APIResourceSpecification {
			whiteListFilter[gv.GroupVersion] = make(map[string]struct{})
			for _, resource := range gv.ResourceKinds {
				whiteListFilter[gv.GroupVersion][resource] = struct{}{}
			}
		}
	}
	return &ResourceFilter{
		blackListFilter: blackListFilter,
		whiteListFilter: whiteListFilter,
		filterConfig:    filterConfig,
	}
}

// IsBanned return true resource is banned
func (rf *ResourceFilter) IsBanned(
	groupVersion string, apiResource options.APIResource) bool {
	if apiResource.Kind != "Namespace" {
		// check black list.
		resourceFiltered, resourceFilterOK := rf.blackListFilter[groupVersion]
		if resourceFilterOK && len(resourceFiltered) == 0 {
			glog.Warnf("filter has banned all resource in groupversion %s", groupVersion)
			return true
		}
		if _, filtered := resourceFiltered[apiResource.Kind]; filtered && resourceFilterOK {
			glog.Warnf("filter has banned resource kind %s in groupversion %s", apiResource.Kind, groupVersion)
			return true
		}
		// check white list. if white list is not empty, then do filter
		if len(rf.whiteListFilter) != 0 {
			resourceSpecified, resourceSpecifyOk := rf.whiteListFilter[groupVersion]
			if !resourceSpecifyOk {
				return true
			}
			if len(resourceSpecified) != 0 {
				if _, isSpecified := resourceSpecified[apiResource.Kind]; !isSpecified {
					return true
				}
			}
		}
	}
	if apiResource.Kind == "ComponentStatus" ||
		apiResource.Kind == "Binding" ||
		apiResource.Kind == "ReplicationControllerDummy" {
		// 这几种类型的资源无法watch，跳过
		return true
	}
	if groupVersion == StorageV1GroupVersion && apiResource.Kind != "StorageClass" {
		// 1.12版本的 VolumeAttachment在v1beta1下，但1.14版本放到了v1下，为了避免list报错，暂时只同步StorageClass
		return true
	}
	return false
}
