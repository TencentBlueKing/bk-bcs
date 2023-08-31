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

package utils

// EnvType for vcluster env
type EnvType string

// String xxx
func (et EnvType) String() string {
	return string(et)
}

var (
	// IDC for vcluster idc env
	IDC EnvType = "idc"
	// DEVNET for vcluster devnet env
	DEVNET EnvType = "devnet"
)

// NetworkType for network
type NetworkType string

// String trans NetworkType to string
func (nt NetworkType) String() string {
	return string(nt)
}

var (
	// GlobalRouter for globalrouter
	GlobalRouter NetworkType = "globalrouter"
	// DirectRouter for directrouter
	DirectRouter NetworkType = "directrouter"
)

const (
	// ClusterCIDR cluster cidr
	ClusterCIDR = "ClusterCIDR"
	// MultiClusterCIDR cluster multi cidr
	MultiClusterCIDR = "MultiClusterCIDR"
)

const (
	// NodeGroupLabelKey for CA nodes common label
	NodeGroupLabelKey = "bkbcs.tencent.com/nodegroupid"
	// RegionLabelKey for node region label
	RegionLabelKey = "topology.bkbcs.tencent.com/region"
	// BusinessIDLabelKey for businessID label
	BusinessIDLabelKey = "bkcmdb.tencent.com/bk-biz-id"
	// SubZoneIDLabelKey for cc sub zone id
	SubZoneIDLabelKey = "bkbcs.tencent.com/szoneID"
	// RegionBcsLabelKey for device region label
	RegionBcsLabelKey = "node.bkbcs.tencent.com/region"
	// DrainDelayKey for device delay label
	DrainDelayKey = "node.bkbcs.tencent.com/drain-delay"
	// DeviceLabelFlag for device labels flag
	DeviceLabelFlag = "bkbcs.tencent.com"
)

const (
	// TencentCloud qcloud
	TencentCloud = "tencentCloud"
	// ProjectCode project
	ProjectCode = "io.tencent.bcs.projectcode"
	// NamespaceCreator creator
	NamespaceCreator = "io.tencent.bcs.creator"
	// NamespaceVcluster vcluster
	NamespaceVcluster = "io.tencent.bcs.vcluster"
)

// namespace
const (
	// BkSystem namespace
	BkSystem = "bk-system"
	// BCSSystem namespace
	BCSSystem = "bcs-system"
)

// cloud account type
const (
	// STANDARD 标准用户
	STANDARD = "STANDARD"
	// LEGACY 传统用户
	LEGACY = "LEGACY"
)
