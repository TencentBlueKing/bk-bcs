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

package clusterconfig

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	dbutils "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
)

// db 方法

// GetData 查询数据
func GetData(ctx context.Context, resourceType string, opt *lib.StoreGetOption) ([]operator.M, error) {
	return dbutils.GetData(&dbutils.DBOperate{
		GetOpt:       opt,
		Context:      ctx,
		DBConfig:     dbConfig,
		ResourceType: resourceType,
	})
}

// PutData 插入数据
func PutData(ctx context.Context, resourceType string, data operator.M, opt *lib.StorePutOption) error {
	//return dbutils.PutData(ctx, dbConfig, resourceType, data, opt)
	return dbutils.PutData(&dbutils.DBOperate{
		PutOpt:       opt,
		Context:      ctx,
		Data:         data,
		DBConfig:     dbConfig,
		ResourceType: resourceType,
	})
}

/*
	业务方法
*/

// GenerateData 生成数据
func GenerateData(ctx context.Context, opt *lib.StoreGetOption, clsConfig []operator.M, service string) (config *types.DeployConfig, err error) {
	var stableVersion string
	var svcConfigSet *types.ConfigSet
	var clsConfigSet []types.ClusterSet

	if svcConfigSet, err = getSvcSet(ctx, tableSvc, opt); err != nil {
		return &types.DeployConfig{}, err
	}

	if clsConfigSet, err = getClsSet(clsConfig); err != nil {
		return &types.DeployConfig{}, err
	}

	if stableVersion, err = GetStableSvcVersion(ctx, opt); err != nil {
		return &types.DeployConfig{}, err
	}

	config = &types.DeployConfig{
		Service:       service,
		ServiceConfig: *svcConfigSet,
		Clusters:      clsConfigSet,
		StableVersion: stableVersion,
	}
	return config, nil
}

// GetStableSvcVersion get stable version
func GetStableSvcVersion(ctx context.Context, opt *lib.StoreGetOption) (string, error) {
	mList, err := GetData(ctx, tableVer, opt)
	if err != nil {
		return "", err
	}
	vs, _ := mList[0][dataTag]
	s, ok := vs.(string)
	if !ok {
		err = storageErr.StableVersionInvalid
		blog.Errorf("%v", err)
		return "", err
	}
	return s, nil
}

// GetTemplate get template
func GetTemplate(ctx context.Context, cond *operator.Condition) (string, error) {
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)
	opt := &lib.StoreGetOption{
		Cond: cond,
		Sort: map[string]int{
			versionTag: -1,
		},
	}
	mList, err := store.Get(ctx, tableTpl, opt)
	if err != nil {
		return "", err
	}
	if len(mList) == 0 {
		err := storageErr.ConfigTemplateNoFound
		blog.Errorf("%s", err.Error())
		return "", err
	}
	vs, _ := mList[0][dataTag]
	s, ok := vs.(string)
	if !ok {
		err := storageErr.ConfigTemplateInvalid
		blog.Errorf("%s", err.Error())
		return "", err
	}
	return s, nil
}

// GetClusterInfo get cluster info
func GetClusterInfo(ctx context.Context, cond *operator.Condition) ([]operator.M, error) {
	// 获取db连接
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)
	// option
	opt := &lib.StoreGetOption{
		Cond: cond,
	}
	return store.Get(ctx, tableCls, opt)
}

// SaveClusterInfoConfig save cluster info config
func SaveClusterInfoConfig(ctx context.Context, data operator.M, opt *lib.StorePutOption) error {
	return PutData(ctx, tableCls, data, opt)
}

// SaveStableVersion save stable version
func SaveStableVersion(ctx context.Context, service, version string) error {
	// 参数
	data := operator.M{constants.DataTag: version}
	// option
	opt := &lib.StorePutOption{
		Cond: operator.NewLeafCondition(operator.Eq, operator.M{constants.ServiceTag: service}),
	}
	return PutData(ctx, tableVer, data, opt)
}
