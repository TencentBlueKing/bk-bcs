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

package business

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// UpdateCloudKubeConfig update cloud kubeconfig
func UpdateCloudKubeConfig(kubeConfig string, opt *cloudprovider.UpdateCloudKubeConfigOption) (bool, error) {
	if kubeConfig == "" {
		// 获取集群是内网还是外网
		client, err := api.NewTkeClient(&opt.CommonOption)
		if err != nil {
			return false, err
		}

		isExtranet := true
		status, err := client.GetClusterEndpointStatus(opt.Cluster.SystemID, false)
		if err != nil {
			return false, err
		}

		blog.Infof("cluster endpoint status: %s", status)

		if status.Created() {
			isExtranet = false
		}

		// 获取集群的kubeconfig
		kubeConfig, err = client.GetTKEClusterKubeConfig(opt.Cluster.SystemID, isExtranet)
		if err != nil {
			return false, err
		}
	}

	kubeRet, err := encrypt.Encrypt(nil, kubeConfig)
	if err != nil {
		return false, err
	}

	opt.Cluster.KubeConfig = kubeRet
	cloudprovider.GetStorageModel().UpdateCluster(context.Background(), opt.Cluster)

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
