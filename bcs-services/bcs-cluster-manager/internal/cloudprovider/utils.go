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
	"math/rand"
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
	TaskID    = "taskID"
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

// GetCredential get specified credential information according Project configuration, if Project conf is nil, try Cloud.
// @return CommonOption: option can be nil if no credential conf in project and cloud or when cloudprovider don't support authentication
func GetCredential(project *proto.Project, cloud *proto.Cloud) (*CommonOption, error) {
	if project == nil {
		return nil, fmt.Errorf("lost Project information")
	}
	if cloud == nil {
		return nil, fmt.Errorf("lost cloud information")
	}
	option := &CommonOption{}
	if len(project.Credentials) != 0 {
		if cred, ok := project.Credentials[cloud.CloudID]; ok {
			option.Key = cred.Key
			option.Secret = cred.Secret
		}
	}
	if len(option.Key) == 0 && cloud.CloudCredential != nil {
		// try to get credential in cloud
		option.Key = cloud.CloudCredential.Key
		option.Secret = cloud.CloudCredential.Secret
	}
	// set cloud basic confInfo
	option.CommonConf = CloudConf{
		CloudInternalEnable: cloud.ConfInfo.CloudInternalEnable,
		CloudDomain:         cloud.ConfInfo.CloudDomain,
		MachineDomain:       cloud.ConfInfo.MachineDomain,
	}

	return option, nil
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

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

// RandomString get n length random string.
// implementation comes from
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go .
func RandomString(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

var (
	nums        = "0123456789"
	lower       = "abcdefghijklmnopqrstuvwxyz"
	upper       = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialChar = "@#+_-[]{}"
)

func getLenRandomString(str string, length int) string {
	bytes := []byte(str)

	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(str))])
	}
	return string(result)
}

// BuildInstancePwd build instance init passwd
func BuildInstancePwd() string {
	randomStr := []string{lower, upper, nums, specialChar}

	totalRandomList := ""
	for i := range randomStr {
		totalRandomList += getLenRandomString(randomStr[i], 3)
	}

	byteRandom := []byte(totalRandomList)
	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(byteRandom), func(i, j int) { byteRandom[i], byteRandom[j] = byteRandom[j], byteRandom[i] })

	return "Bcs#" + string(byteRandom)
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
	Cluster  *proto.Cluster
	// Cloud info
	Cloud    *proto.Cloud
	// Project info
	Project  *proto.Project
	// CmOption option
	CmOption *CommonOption
}

// GetClusterDependBasicInfo get cluster and cloud depend info
func GetClusterDependBasicInfo(clusterID string, cloudID string) (*CloudDependBasicInfo, error) {
	cluster, err := GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		return nil, err
	}

	cloud, project, err := actions.GetProjectAndCloud(GetStorageModel(), cluster.ProjectID, cloudID)
	if err != nil {
		return nil, err
	}

	cmOption, err := GetCredential(project, cloud)
	if err != nil {
		return nil, err
	}
	cmOption.Region = cluster.Region

	return &CloudDependBasicInfo{cluster, cloud, project, cmOption}, nil
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
		server = ""
		caCertData = ""
		token = ""
		clientCert = ""
		clientKey = ""
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
		ServerKey:            clusterID,
		ClusterID:            clusterID,
		ClientModule:         modules.BCSModuleKubeagent,
		ServerAddress:        server,
		CaCertData:           caCertData,
		UserToken:            token,
		ConnectMode:          modules.BCSConnectModeDirect,
		CreateTime:           now,
		UpdateTime:           now,
		ClientKey:            clientKey,
		ClientCert:           clientCert,
	})
	if err != nil {
		return err
	}

	return nil
}
