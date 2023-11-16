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

package actions

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/stringx"
)

type basePermInfo struct {
	username       string
	projectCode    string
	projectID      string
	clusterID      string
	isShardCluster bool
}

func checkReleaseAccess(manifest []release.SimpleHead, resources []*metav1.APIResourceList,
	basePermInfo basePermInfo) error {
	// check has cluster scope resource
	clusterScope := false
	// check has namespace
	createNamespace := false
	// check has namespace scope resource
	namespaceScope := make([]string, 0)
	for _, v := range manifest {
		if v.Kind == "Namespace" {
			createNamespace = true
			continue
		}
		for _, resource := range resources {
			if resource.GroupVersion != v.Version {
				continue
			}
			for _, item := range resource.APIResources {
				if item.Kind != v.Kind {
					continue
				}
				if item.Namespaced {
					namespaceScope = append(namespaceScope, v.Metadata.Namespace)
				} else {
					clusterScope = true
				}
			}
		}
	}
	namespaceScope = stringx.RemoveDuplicateValues(namespaceScope)
	allow, url, _, err := auth.ReleaseResourcePermCheck(basePermInfo.projectCode,
		basePermInfo.clusterID, createNamespace, clusterScope, namespaceScope)
	if err != nil {
		return err
	}
	if !allow {
		return fmt.Errorf("You do not have permission to operate cluster resources. "+
			"Please apply through the provided link. %s", url)
	}
	return nil
}
