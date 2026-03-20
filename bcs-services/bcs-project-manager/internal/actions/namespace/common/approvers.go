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
	"context"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	configm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
)

// GetBcsApprovers get bcs approvers
func GetBcsApprovers() string {
	itsmConf := config.GlobalConf.ITSM
	if itsmConf.AutoRegister {
		approvers, err := store.GetModel().GetConfig(context.Background(),
			configm.ConfigKeyNamespaceItsmApprovers)
		if err != nil {
			blog.Warnf("GetBcsApprovers GetConfig error:[%v]", err)
			return ""
		}
		return approvers
	}

	return itsmConf.Approvers
}

// GetClusterApprovers get cluster approvers
func GetClusterApprovers(cluster *clustermanager.Cluster) string {
	clsApprover := []string{
		cluster.GetCreator(), cluster.GetUpdater(),
	}

	clsApprover = stringx.RemoveDuplicateAndEmptyValues(clsApprover)

	return strings.Join(clsApprover, ",")
}
