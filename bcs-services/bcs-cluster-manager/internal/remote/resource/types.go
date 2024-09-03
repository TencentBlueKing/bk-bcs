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

package resource

const (
	// ResourcePoolType pool type
	ResourcePoolType = "resourcePoolType"
	// DevicePoolIds pool id
	DevicePoolIds = "devicePoolIds"
	// SelfPool 自建资源池
	SelfPool = "self"
	// YunTiPool yunti资源池
	YunTiPool = "yunti"
	// CrPool cr资源池
	CrPool = "cr"
	// BcsResourcePool bcs pool
	BcsResourcePool = "bcs_resource_pool"
)

// NodeType instance type
type NodeType string

// String toString
func (nt NodeType) String() string {
	return string(nt)
}

var (
	// CVM cloud instance
	CVM NodeType = "CVM"
	// IDC xxx
	IDC NodeType = "IDC"
)

// DataDisk disk define
type DataDisk struct {
	DiskType string
	DiskSize string
}

// SubnetZone subnet&zone info
type SubnetZone struct {
	Subnet string
	Zone   string
}

// ImageInfo image info
type ImageInfo struct {
	ImageID   string
	ImageName string
}

// GPUInfo gpu info
type GPUInfo struct {
	// 实例GPU个数。值小于1代表VGPU类型，大于1代表GPU直通类型。
	GPUCount float64
	// 实例GPU地址。
	GPUId []string
	// 实例GPU类型。
	GPUType string
}

// InternetAccessible xxx
type InternetAccessible struct {
	// InternetChargeType 网络计费类型
	InternetChargeType string
	// InternetMaxBandwidthOut 公网出带宽上限，单位：Mbps。默认值：0Mbps。
	InternetMaxBandwidthOut int64
	// PublicIpAssigned 是否分配公网IP
	PublicIpAssigned bool
	// BandwidthPackageId 带宽包ID
	BandwidthPackageId string
}

// LoginSettings xxx
type LoginSettings struct {
	// Password 实例登录密码。不同操作系统类型密码复杂度限制不一样
	Password string
}

// EnhancedService 增强服务
type EnhancedService struct {
	// SecurityService 开启云安全服务
	SecurityService bool
	// MonitorService 开启云监控服务
	MonitorService bool
}

// ApplyInstanceReq xxx
type ApplyInstanceReq struct {
	// NodeType instanceType
	NodeType NodeType
	// 地域子网信息
	Region     string
	VpcID      string
	ZoneList   []string // ap-nanjing-3
	SubnetList []string // 子网ID
	// 实例信息
	InstanceType string
	// 实例的CPU核数，单位：核。
	CPU    uint32
	Memory uint32
	Gpu    uint32
	// 实例计费模式. `PREPAID`：表示预付费，即包年包月; `POSTPAID_BY_HOUR`：表示后付费，即按量计费
	InstanceChargeType string
	// 实例系统盘信息
	SystemDisk DataDisk
	// 实例数据盘信息。
	DataDisks []DataDisk
	// 镜像信息
	Image *ImageInfo
	// 公网访问
	InternetAccess *InternetAccessible
	// 登录信息
	LoginInfo *LoginSettings
	// 实例所属安全组
	SecurityGroupIds []string `json:"SecurityGroupIds,omitempty" name:"SecurityGroupIds"`
	// 增强服务
	EnhancedService *EnhancedService `json:"EnhancedService,omitempty" name:"EnhancedService"`
	// UserData 实例执行脚本
	UserData string

	// PoolID resourcePool id
	PoolID string
	// Operator resource applicants
	Operator string

	// Selector labels match
	Selector map[string]string
}

// ApplyInstanceResp return async task bu orderID or return instanceIDs and check instance status
type ApplyInstanceResp struct {
	OrderID     string   `json:"orderID"`
	InstanceIDs []string `json:"instanceIDs"`
	InstanceIPs []string `json:"instanceIPs"`
}

// DestroyInstanceReq destroy instances request
type DestroyInstanceReq struct {
	Region      string
	PoolID      string
	SystemID    string
	InstanceIDs []string
	Operator    string
}

// DestroyInstanceResp check instance status by orderID or instances
type DestroyInstanceResp struct {
	OrderID     string   `json:"orderID"`
	InstanceIDs []string `json:"instanceIDs"`
	InstanceIPs []string `json:"instanceIPs"`
}

// OrderInstanceList order instanceInfo
type OrderInstanceList struct {
	OrderStatus bool
	Message     string
	InstanceIDs []string
	InstanceIPs []string
	ExtraIDs    []string
}

// InstanceType instance type info
type InstanceType struct {
	NodeType          string
	TypeName          string
	NodeFamily        string
	Cpu               uint32
	Memory            uint32
	Gpu               uint32
	Status            string
	UnitPrice         float32
	Zones             []string
	Provider          string
	ResourcePoolID    string
	SystemDisk        *DataDisk
	DataDisks         []*DataDisk
	OversoldAvailable int32
}

// InstanceSpec size
type InstanceSpec struct {
	BizID        string
	Provider     string
	Cpu          uint32
	Mem          uint32
	ResourceType string
}

// ResourcePoolInfo resource pool info
type ResourcePoolInfo struct { // nolint
	Name               string
	Provider           string
	ClusterID          string
	RelativeDevicePool []string
	PoolID             []string
	Operator           string
}

// BuildResourcePoolLabels build labels for resource pool
func BuildResourcePoolLabels(poolID string) map[string]string {
	return map[string]string{
		ResourcePoolID.String(): poolID,
	}
}

// LabelKey for resourcePoolInfo labels
type LabelKey string

// String xxx
func (lk LabelKey) String() string {
	return string(lk)
}

var (
	// ResourcePoolID label
	ResourcePoolID LabelKey = "resourcePoolID"
	// ProductionName label
	ProductionName LabelKey = "productionName"
	// ProductionID label
	ProductionID LabelKey = "productionID"
)

// DeviceInfo device detailed info
type DeviceInfo struct {
	DeviceID           string
	Provider           string
	Labels             map[string]string
	Annotations        map[string]string
	Status             string
	DevicePoolID       string
	Instance           string
	InnerIP            string
	InstanceType       string
	Cpu                uint32
	Mem                uint32
	Gpu                uint32
	Vpc                string
	Region             string
	Zone               string
	LastConsumerId     string
	LastRecordId       string
	LastReturnRecordId string
}

// DevicePoolInfo 资源池信息
type DevicePoolInfo struct {
	PoolId       string
	PoolName     string
	Region       string
	Zone         string // 如果是多个, 通过;隔开
	InstanceType string

	// 节点池实际使用情况
	Total     int32
	Used      int32
	Available int32

	OversoldTotal     int32
	OversoldAvailable int32

	Status string // 节点池状态

	// 节点池被占用情况
	GroupUsed  int
	GroupQuota int
}

// PoolUsage device pool usage
type PoolUsage struct {
	Total     int32
	Used      int32
	Available int32

	OversoldTotal     int32
	OversoldAvailable int32
}

// UsageRatio ratio
type UsageRatio struct {
	QuotaRatio *int // 空闲率
	QuotaCount *int // 空闲资源数
}
