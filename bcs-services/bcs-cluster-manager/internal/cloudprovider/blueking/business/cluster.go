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

// Package business xxx
package business

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// UpdateCloudKubeConfig update cloud kubeconfig
func UpdateCloudKubeConfig(kubeConfig string, opt *cloudprovider.UpdateCloudKubeConfigOption) (bool, error) {
	kubeRet, err := encrypt.Encrypt(nil, kubeConfig)
	if err != nil {
		return false, err
	}

	opt.Cluster.KubeConfig = kubeRet
	err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), opt.Cluster)
	if err != nil {
		return false, err
	}

	config, err := types.GetKubeConfigFromYAMLBody(false, types.YamlInput{
		YamlContent: kubeRet,
	})
	if err != nil {
		return false, err
	}

	err = cloudprovider.UpdateClusterCredentialByConfig(opt.Cluster.ClusterID, config)
	if err != nil {
		return false, err
	}

	return true, nil
}
