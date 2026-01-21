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

// Package types pod types
package types

// Cluster 集群信息
type Cluster struct {
	ProjectID       string           `json:"projectID"`
	ClusterID       string           `json:"clusterID"`
	ClusterName     string           `json:"clusterName"`
	Provider        string           `json:"provider"`
	Region          string           `json:"region"`
	BKBizID         string           `json:"businessID"`
	Status          string           `json:"status"`
	IsShared        bool             `json:"is_shared"`
	ClusterType     string           `json:"clusterType"`
	VpcID           string           `json:"vpcID"`
	NetworkSettings *NetworkSettings `json:"networkSettings"`
	ExtraInfo       *ExtraInfo       `json:"extraInfo"`
}

// NetworkSettings 网络设置
type NetworkSettings struct {
	MaxNodePodNum int           `json:"maxNodePodNum"`
	MaxServiceNum int           `json:"maxServiceNum"`
	EnableVPCCni  bool          `json:"enableVPCCni"`
	EniSubnetIDs  []string      `json:"eniSubnetIDs"`
	SubnetSource  *SubnetSource `json:"subnetSource"`
}

// ExtraInfo 额外信息
type ExtraInfo struct {
	NamespaceInfo   string `json:"namespaceInfo"`
	Provider        string `json:"provider"`
	VclusterNetwork string `json:"vclusterNetwork"`
}

// SubnetSource VPC-CNI网络模式下申请subnet或使用已存在的subnet
type SubnetSource struct {
	New     []*NewSubnet      `json:"new"`
	Existed *ExistedSubnetIDs `json:"existed"`
}

// NewSubnet VPC-CNI网络模式下申请subnet
type NewSubnet struct {
	Mask  uint32 `json:"mask"`
	Zone  string `json:"zone"`
	IpCnt uint32 `json:"ipCnt"`
}

// ExistedSubnetIDs VPC-CNI网络模式下使用已存在的subnet
type ExistedSubnetIDs struct {
	IDs []string `json:"ids"`
}

const (
	// VirtualClusterType vcluster
	VirtualClusterType = "virtual"
)

// IsVirtual check cluster is vcluster
func (c *Cluster) IsVirtual() bool {
	return c.ClusterType == VirtualClusterType
}
