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

package cloudprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/modules"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

var (
	// BKSOPTask bk-sops common job
	BKSOPTask = "bksopsjob"
	// TaskID inject taskID into ctx
	TaskID = "taskID"
)

// GetTaskIDFromContext
func GetTaskIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(TaskID).(string); ok {
		return id
	}

	return ""
}

// WithTaskIDForContext will return a new context wrapped taskID flag around the original ctx
func WithTaskIDForContext(ctx context.Context, taskID string) context.Context {
	return context.WithValue(ctx, TaskID, taskID)
}

// CredentialData dependency data
type CredentialData struct {
	// Cloud cloud
	Cloud *proto.Cloud
	// Cluster cluster
	AccountID string
}

// GetCredential get specified credential information according Cloud configuration, if Cloud conf is nil, try Cluster Account.
// @return CommonOption: option can be nil if no credential conf in cloud or cluster account or when cloudprovider don't support authentication
// GetCredential get cloud credential by cloud or cluster
func GetCredential(data *CredentialData) (*CommonOption, error) {
	if data.Cloud == nil && data.AccountID == "" {
		return nil, fmt.Errorf("lost cloud/account information")
	}

	option := &CommonOption{}
	// get credential from cloud
	if data.Cloud.CloudCredential != nil {
		option.Key = data.Cloud.CloudCredential.Key
		option.Secret = data.Cloud.CloudCredential.Secret
	}

	// if credential not exist cloud, get from cluster account
	if len(option.Key) == 0 && data.AccountID != "" {
		// try to get credential in cluster
		account, err := GetStorageModel().GetCloudAccount(context.Background(), data.Cloud.CloudID, data.AccountID)
		if err != nil {
			return nil, fmt.Errorf("GetCloudAccount failed: %v", err)
		}
		option.Key = account.Account.SecretID
		option.Secret = account.Account.SecretKey
	}

	// set cloud basic confInfo
	option.CommonConf = CloudConf{
		CloudInternalEnable: data.Cloud.ConfInfo.CloudInternalEnable,
		CloudDomain:         data.Cloud.ConfInfo.CloudDomain,
		MachineDomain:       data.Cloud.ConfInfo.MachineDomain,
	}

	// check cloud credential info
	err := checkCloudCredentialValidate(data.Cloud, option)
	if err != nil {
		return nil, fmt.Errorf("checkCloudCredentialValidate %s failed: %v", data.Cloud.CloudProvider, err)
	}

	return option, nil
}

func checkCloudCredentialValidate(cloud *proto.Cloud, option *CommonOption) error {
	validate, err := GetCloudValidateMgr(cloud.CloudProvider)
	if err != nil {
		return err
	}
	err = validate.ImportCloudAccountValidate(&proto.Account{
		SecretID:  option.Key,
		SecretKey: option.Secret,
	})
	if err != nil {
		return err
	}

	return nil
}

// GetCredentialByCloudID get credentialInfo by cloudID
func GetCredentialByCloudID(cloudID string) (*CommonOption, error) {
	cloud, err := GetStorageModel().GetCloud(context.Background(), cloudID)
	if err != nil {
		return nil, fmt.Errorf("GetCredentialByCloudID getCloud failed: %v", err)
	}

	option := &CommonOption{}
	option.Key = cloud.CloudCredential.Key
	option.Secret = cloud.CloudCredential.Secret

	return option, nil
}

// TaskType taskType
type TaskType string

// String() toString
func (tt TaskType) String() string {
	return string(tt)
}

var (
	// CreateCluster task
	CreateCluster TaskType = "CreateCluster"
	// ImportCluster task
	ImportCluster TaskType = "ImportCluster"
	// DeleteCluster task
	DeleteCluster TaskType = "DeleteCluster"
	// AddNodesToCluster task
	AddNodesToCluster TaskType = "AddNodesToCluster"
	// RemoveNodesFromCluster task
	RemoveNodesFromCluster TaskType = "RemoveNodesFromCluster"
)

// GetTaskType getTaskType by cloud
func GetTaskType(cloud string, taskName TaskType) string {
	return fmt.Sprintf("%s-%s", cloud, taskName.String())
}

// CloudDependBasicInfo cloud depend cluster info
type CloudDependBasicInfo struct {
	// Cluster info
	Cluster *proto.Cluster
	// Cloud info
	Cloud *proto.Cloud
	// CmOption option
	CmOption *CommonOption
}

// GetClusterDependBasicInfo get cluster and cloud depend info
func GetClusterDependBasicInfo(clusterID string, cloudID string) (*CloudDependBasicInfo, error) {
	cluster, err := GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		return nil, err
	}

	cloud, err := actions.GetCloudByCloudID(GetStorageModel(), cloudID)
	if err != nil {
		return nil, err
	}

	cmOption, err := GetCredential(&CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		return nil, err
	}
	cmOption.Region = cluster.Region

	return &CloudDependBasicInfo{cluster, cloud, cmOption}, nil
}

// UpdateClusterStatus set cluster status
func UpdateClusterStatus(clusterID string, status string) error {
	cluster, err := GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		return err
	}

	cluster.Status = status
	err = GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		return err
	}

	return nil
}

// UpdateClusterStatus set cluster status
func UpdateCluster(cluster *proto.Cluster) error {
	err := GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		return err
	}

	return nil
}

// UpdateClusterCredentialByConfig update clusterCredential by kubeConfig
func UpdateClusterCredentialByConfig(clusterID string, config *types.Config) error {
	// first import cluster need to auto generate clusterCredential info, subsequently kube-agent report to update
	// currently, bcs only support token auth, kubeConfigList length greater 0, get zeroth kubeConfig
	var (
		server     = ""
		caCertData = ""
		token      = ""
		clientCert = ""
		clientKey  = ""
	)
	if len(config.Clusters) > 0 {
		server = config.Clusters[0].Cluster.Server
		caCertData = string(config.Clusters[0].Cluster.CertificateAuthorityData)
	}
	if len(config.AuthInfos) > 0 {
		token = config.AuthInfos[0].AuthInfo.Token
		clientCert = string(config.AuthInfos[0].AuthInfo.ClientCertificateData)
		clientKey = string(config.AuthInfos[0].AuthInfo.ClientKeyData)
	}

	if server == "" || caCertData == "" || (token == "" && (clientCert == "" || clientKey == "")) {
		return fmt.Errorf("importClusterCredential parse kubeConfig failed: %v", "[server|caCertData|token] null")
	}

	now := time.Now().Format(time.RFC3339)
	err := GetStorageModel().PutClusterCredential(context.Background(), &proto.ClusterCredential{
		ServerKey:     clusterID,
		ClusterID:     clusterID,
		ClientModule:  modules.BCSModuleKubeagent,
		ServerAddress: server,
		CaCertData:    caCertData,
		UserToken:     token,
		ConnectMode:   modules.BCSConnectModeDirect,
		CreateTime:    now,
		UpdateTime:    now,
		ClientKey:     clientKey,
		ClientCert:    clientCert,
	})
	if err != nil {
		return err
	}

	return nil
}
