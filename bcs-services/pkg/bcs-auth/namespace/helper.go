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
 *
 */

package namespace

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
)

// ResourceTypeIDMap xxx
var ResourceTypeIDMap = map[iam.TypeID]string{
	SysNamespace: "命名空间",
}

const (
	// SysNamespace resource namespace
	SysNamespace iam.TypeID = "namespace"
)

// NamespaceResourcePath  build IAMPath for namespace resource
type NamespaceResourcePath struct {
	ProjectID     string
	ClusterID     string
	IsClusterPerm bool
}

// BuildIAMPath build IAMPath
func (rp NamespaceResourcePath) BuildIAMPath() string {
	// special case to handle create namespace resource
	if rp.IsClusterPerm {
		return fmt.Sprintf("/project,%s/", rp.ProjectID)
	}
	return fmt.Sprintf("/project,%s/cluster,%s/", rp.ProjectID, rp.ClusterID)
}

// NamespaceScopedResourcePath  build IAMPath for namespace scoped resource
type NamespaceScopedResourcePath struct {
	ProjectID string
	ClusterID string
}

// BuildIAMPath build IAMPath
func (rp NamespaceScopedResourcePath) BuildIAMPath() string {
	return fmt.Sprintf("/project,%s/cluster,%s/", rp.ProjectID, rp.ClusterID)
}
