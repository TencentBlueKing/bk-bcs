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

package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

func importClusterCredential(ctx context.Context, data *cloudprovider.CloudDependBasicInfo) error { // nolint
	kubeConfig := data.Cluster.KubeConfig
	if len(kubeConfig) == 0 {
		client, err := api.NewEksClient(data.CmOption)
		if err != nil {
			return fmt.Errorf("create eks client failed, %v", err)
		}

		eksCluster, err := client.GetEksCluster(data.Cluster.SystemID)
		if err != nil {
			return fmt.Errorf("get eks cluster failed, %v", err)
		}

		// generate kube config
		kubeConfig, err = api.GetClusterKubeConfig(data.CmOption, eksCluster)
		if err != nil {
			return fmt.Errorf("get cluster kubeconfig failed, %v", err)
		}
	}

	// decrypt kube config
	configByte, err := encrypt.Decrypt(nil, kubeConfig)
	if err != nil {
		return fmt.Errorf("failed to decrypt kubeconfig, %v", err)
	}

	typesConfig := &types.Config{}
	err = json.Unmarshal([]byte(configByte), typesConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal kubeconfig, %v", err)
	}

	err = cloudprovider.UpdateClusterCredentialByConfig(data.Cluster.ClusterID, typesConfig)
	if err != nil {
		return err
	}

	return nil
}
