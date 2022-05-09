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

package iam

import (
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm"
	clusterAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/cluster"
	nsAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/namespace"
)

// NewNSPerm ...
func NewNSPerm(projectID, clusterID string) perm.Perm {
	if runtime.RunMode == runmode.Dev || runtime.RunMode == runmode.UnitTest {
		return &MockPerm{}
	}
	return nsAuth.NewPerm(projectID, clusterID)
}

// NewNSScopedPerm ...
func NewNSScopedPerm(projectID, clusterID string) perm.Perm {
	if runtime.RunMode == runmode.Dev || runtime.RunMode == runmode.UnitTest {
		return &MockPerm{}
	}
	return nsAuth.NewScopedPerm(projectID, clusterID)
}

// NewClusterPerm ...
func NewClusterPerm(projectID string) perm.Perm {
	if runtime.RunMode == runmode.Dev || runtime.RunMode == runmode.UnitTest {
		return &MockPerm{}
	}
	return clusterAuth.NewPerm(projectID)
}

// NewClusterScopedPerm ...
func NewClusterScopedPerm(projectID string) perm.Perm {
	if runtime.RunMode == runmode.Dev || runtime.RunMode == runmode.UnitTest {
		return &MockPerm{}
	}
	return clusterAuth.NewScopedPerm(projectID)
}
