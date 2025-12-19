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

// Package templateconfig xxx
package templateconfig

import (
	"context"
	"encoding/json"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/templateconfig"
)

// ConfigType xxx
var ConfigType = map[string]struct{}{
	common.CloudConfigType:    {},
	common.TaskTimeConfigType: {},
}

func checkTemplateConfigType(confType string) bool {
	if _, ok := ConfigType[confType]; ok {
		return true
	}
	return false
}

// getTemplateConfigInfos list all templateConfigInfos
func getTemplateConfigInfos(ctx context.Context, model store.ClusterManagerModel, businessID, projectID, clusterID,
	provider, confType string, opt *options.ListOption) ([]*proto.TemplateConfigInfo, error) {
	condM := make(operator.M)
	//! we don't setting bson tag in proto file
	//! all fields are in lowcas
	if businessID != "" {
		condM[templateconfig.BusinessIDKey] = businessID
	}
	if projectID != "" {
		condM[templateconfig.ProjectIDKey] = projectID
	}
	if clusterID != "" {
		condM[templateconfig.ClusterIDKey] = clusterID
	}
	if provider != "" {
		condM[templateconfig.ProviderKey] = provider
	}
	if confType != "" {
		condM[templateconfig.ConfigTypeKey] = confType
	}
	cond := operator.NewLeafCondition(operator.Eq, condM)

	if opt == nil {
		opt = &options.ListOption{}
	}

	templateConfigs, err := model.ListTemplateConfigs(ctx, cond, opt)
	if err != nil {
		return nil, err
	}

	cloudTemplateConfigs := make([]*proto.TemplateConfigInfo, 0)
	for _, templateConfig := range templateConfigs {
		var cloudConfig *proto.CloudTemplateConfig
		if err := json.Unmarshal([]byte(templateConfig.ConfigContent), &cloudConfig); err != nil {
			blog.Errorf("unmarshal cloud config content failed: %v", err)
			continue
		}

		cloudTemplateConfigs = append(cloudTemplateConfigs, &proto.TemplateConfigInfo{
			TemplateConfigID:    templateConfig.TemplateConfigID,
			BusinessID:          templateConfig.BusinessID,
			ProjectID:           templateConfig.ProjectID,
			ClusterID:           templateConfig.ClusterID,
			Provider:            templateConfig.Provider,
			ConfigType:          templateConfig.ConfigType,
			CloudTemplateConfig: cloudConfig,
			Creator:             templateConfig.Creator,
			Updater:             templateConfig.Updater,
			CreateTime:          templateConfig.CreateTime,
			UpdateTime:          templateConfig.UpdateTime,
		})
	}

	return cloudTemplateConfigs, nil
}
