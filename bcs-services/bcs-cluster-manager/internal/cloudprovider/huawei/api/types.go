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

// Package api xxx
package api

import (
	"fmt"
	"sort"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	evsModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/model"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	v1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ClusterFilterCond 集群列表过滤条件
type ClusterFilterCond struct {
	// Status 集群状态
	Status string
	// Type 集群类型(VirtualMachine:CCE集群 ARM64:鲲鹏集群)
	Type string
	// Version 集群版本过滤
	Version string
}

// GetClusterStatus get cluster status
func GetClusterStatus(status string) *model.ListClustersRequestStatus {
	statusEnum := model.GetListClustersRequestStatusEnum()

	switch status {
	case Available:
		return &statusEnum.AVAILABLE
	case Unavailable:
		return &statusEnum.UNAVAILABLE
	case Deleting:
		return &statusEnum.DELETING
	case Creating:
		return &statusEnum.CREATING
	}

	return &statusEnum.EMPTY
}

// GetClusterType get cluster type
func GetClusterType(clusterType string) *model.ListClustersRequestType {
	typeEnum := model.GetListClustersRequestTypeEnum()

	switch clusterType {
	case ARM64:
		return &typeEnum.ARM64
	case VirtualMachine:
		return &typeEnum.VIRTUAL_MACHINE
	}

	return &typeEnum.VIRTUAL_MACHINE
}

// UpdateClusterEipRequest xxx
type UpdateClusterEipRequest struct {
	// Action: bind & unbind
	Action string
	// ElasticNicId 弹性网卡Id
	ElasticNicId string
}

func (uc UpdateClusterEipRequest) validate() error {
	if !utils.StringInSlice(uc.Action, []string{model.GetMasterEipRequestSpecActionEnum().BIND.Value(),
		model.GetMasterEipRequestSpecActionEnum().UNBIND.Value()}) {
		return fmt.Errorf("action[%s] invalid", uc.Action)
	}

	if uc.Action == model.GetMasterEipRequestSpecActionEnum().BIND.Value() && uc.ElasticNicId == "" {
		return fmt.Errorf("ElasticNicId must not empty when action bind")
	}

	return nil
}

func (uc UpdateClusterEipRequest) trans2ClusterEipRequest(clsId string) *model.UpdateClusterEipRequest {
	specSpec := &model.MasterEipRequestSpecSpec{
		Id: &uc.ElasticNicId,
	}
	actionSpec := func() model.MasterEipRequestSpecAction {
		switch uc.Action {
		case model.GetMasterEipRequestSpecActionEnum().BIND.Value():
			return model.GetMasterEipRequestSpecActionEnum().BIND
		case model.GetMasterEipRequestSpecActionEnum().UNBIND.Value():
			return model.GetMasterEipRequestSpecActionEnum().UNBIND
		}

		return model.MasterEipRequestSpecAction{}
	}()
	specBody := &model.MasterEipRequestSpec{
		Action: &actionSpec,
		Spec:   specSpec,
	}

	request := &model.UpdateClusterEipRequest{
		ClusterId: clsId,
	}
	request.Body = &model.MasterEipRequest{Spec: specBody}

	return request
}

// UpdateClusterRequest update cluster request
type UpdateClusterRequest struct {
	// Name 更新集群名称
	Name string
	// Desc 更新集群描述
	Desc string
	// CustomSan 更新集群自定义证书SAN
	CustomSan []string
	// ContainerCidrs 添加集群容器网段
	ContainerCidrs []string
	// NodeSecurityGroup 修改集群默认节点安全组
	NodeSecurityGroup string
	// SubnetIds IPv4子网ID列表。只允许新增子网,不可删除已有子网,请谨慎选择。
	SubnetIds []string
}

func (uc UpdateClusterRequest) trans2ClusterRequest(clsId string) *model.UpdateClusterRequest {
	clusterInfoSpec := &model.ClusterInformationSpec{}
	clusterInfoMeta := &model.ClusterMetadataForUpdate{}

	if len(uc.Name) > 0 {
		clusterInfoMeta.Alias = common.StringPtr(uc.Name)
	}
	if len(uc.Desc) > 0 {
		clusterInfoSpec.Description = common.StringPtr(uc.Desc)
	}
	if len(uc.CustomSan) > 0 {
		clusterInfoSpec.CustomSan = &uc.CustomSan
	}
	if len(uc.ContainerCidrs) > 0 {
		clusterInfoSpec.ContainerNetwork = &model.ContainerNetworkUpdate{
			Cidrs: func() *[]model.ContainerCidr {
				containerCidrs := make([]model.ContainerCidr, 0)
				for i := range uc.ContainerCidrs {
					containerCidrs = append(containerCidrs, model.ContainerCidr{Cidr: uc.ContainerCidrs[i]})
				}
				return &containerCidrs
			}(),
		}
	}
	if len(uc.NodeSecurityGroup) > 0 {
		clusterInfoSpec.HostNetwork = &model.ClusterInformationSpecHostNetwork{
			SecurityGroup: &uc.NodeSecurityGroup,
		}
	}
	if len(uc.SubnetIds) > 0 {
		clusterInfoSpec.EniNetwork = &model.EniNetworkUpdate{
			Subnets: func() *[]model.NetworkSubnet {
				subnetIds := make([]model.NetworkSubnet, 0)
				for i := range uc.SubnetIds {
					subnetIds = append(subnetIds, model.NetworkSubnet{SubnetID: uc.SubnetIds[i]})
				}
				return &subnetIds
			}(),
		}
	}

	request := &model.UpdateClusterRequest{
		ClusterId: clsId,
	}
	request.Body = &model.ClusterInformation{
		Spec:     clusterInfoSpec,
		Metadata: clusterInfoMeta,
	}

	return request
}

// Login node login info
type Login struct {
	SshKey   string
	UserName string
	Passwd   string
}

// RemoveNodesRequest remove nodes
type RemoveNodesRequest struct {
	NodeIds []string
	Login   *Login
}

func (rn RemoveNodesRequest) trans2RemoveNodesRequest(clsId string) (*model.RemoveNodeRequest, error) {
	if len(rn.NodeIds) == 0 || clsId == "" {
		return nil, fmt.Errorf("RemoveNodesRequest nodeIds or clusterId empty")
	}

	listNodesSpec := make([]model.NodeItem, 0)
	for i := range rn.NodeIds {
		listNodesSpec = append(listNodesSpec, model.NodeItem{Uid: rn.NodeIds[i]})
	}
	loginSpec := Login2ModelLogin(rn.Login)

	request := &model.RemoveNodeRequest{
		ClusterId: clsId,
	}
	request.Body = &model.RemoveNodesTask{
		Spec: &model.RemoveNodesSpec{
			Login: loginSpec,
			Nodes: listNodesSpec,
		},
	}

	return request, nil
}

// UpdateNodePoolRequest update node pool request
type UpdateNodePoolRequest struct {
	// Name 节点池名称
	Name string
	// DesiredNodeCount 节点池期望节点数
	DesiredNodeCount int32
	// Labels 节点池labels
	Labels map[string]string
	// Taints 节点池taints
	Taints []v1.Taint
	// NodeSchedule 节点池调度状态，true 可调度；false 不可调度
	NodeSchedule bool
	// PreScript 前置脚本
	PreScript string
	// PostScript 后置脚本
	PostScript string
	// Login 节点密码不允许用户更新

	// AsgConfig 自动扩缩容配置
	AsgConfig *AutoScalingConfig
}

func taintEffectValueTrans(effect v1.TaintEffect) model.TaintEffect {
	enum := model.GetTaintEffectEnum()
	switch effect {
	case v1.TaintEffectNoSchedule:
		return enum.NO_SCHEDULE
	case v1.TaintEffectNoExecute:
		return enum.NO_EXECUTE
	case v1.TaintEffectPreferNoSchedule:
		return enum.PREFER_NO_SCHEDULE
	}

	return enum.NO_SCHEDULE
}

// AutoScalingConfig 自动扩缩容配置
type AutoScalingConfig struct {
	// Enable 默认false
	Enable bool
	// MinNodeCount 节点下限
	MinNodeCount int32
	// MaxNodeCount 节点上限
	MaxNodeCount int32
}

func (un UpdateNodePoolRequest) trans2ModelUpdateNodePoolRequest(clsId string, nodePoolId string) (
	*model.UpdateNodePoolRequest, error) {
	if clsId == "" || nodePoolId == "" {
		return nil, fmt.Errorf("clusterId or nodePoolId empty")
	}

	nodePoolMeta := &model.NodePoolMetadataUpdate{
		Name: un.Name,
	}

	nodeSpec := &model.NodeSpecUpdate{
		Taints:  Taint2ModelTaint(un.Taints),
		K8sTags: un.Labels,
	}

	enableAutoscaling := false
	autoscalingSpec := &model.NodePoolNodeAutoscaling{
		Enable:       &enableAutoscaling,
		MinNodeCount: &un.AsgConfig.MinNodeCount,
		MaxNodeCount: &un.AsgConfig.MaxNodeCount,
	}

	specbody := &model.NodePoolSpecUpdate{
		NodeTemplate:     nodeSpec,
		InitialNodeCount: un.DesiredNodeCount,
		Autoscaling:      autoscalingSpec,
	}

	request := &model.UpdateNodePoolRequest{
		ClusterId:  clsId,
		NodepoolId: nodePoolId,
	}

	request.Body = &model.NodePoolUpdate{
		Spec:     specbody,
		Metadata: nodePoolMeta,
	}

	return request, nil
}

// AddNodesRequest xxx
type AddNodesRequest struct {
	ServerIds []string

	// nodeSpec
	Os    string
	Login *Login

	Labels  map[string]string
	Taints  []v1.Taint
	MaxPods int32

	// PreScript 前置脚本
	PreScript string
	// PostScript 后置脚本
	PostScript string

	// 默认是 privateIp / cceNodeName
	HostNameConfig string
}

// Taint2ModelTaint trans taint
func Taint2ModelTaint(taints []v1.Taint) []model.Taint {
	mTaints := make([]model.Taint, 0)

	for i := range taints {
		mTaints = append(mTaints, model.Taint{
			Key:    taints[i].Key,
			Value:  &taints[i].Value,
			Effect: taintEffectValueTrans(taints[i].Effect),
		})
	}

	return mTaints
}

// Login2ModelLogin trans login
func Login2ModelLogin(login *Login) *model.Login {
	return &model.Login{
		SshKey: func() *string {
			if login != nil && len(login.SshKey) > 0 {
				return &login.SshKey
			}

			return nil
		}(),
		UserPassword: func() *model.UserPassword {
			if login != nil && len(login.Passwd) > 0 {
				if login.UserName == "" {
					// default root name
					login.UserName = "root"
				}

				return &model.UserPassword{
					Username: &login.UserName,
					Password: func() string {
						pwd, _ := Crypt(login.Passwd)
						return pwd
					}(),
				}
			}

			return nil
		}(),
	}
}

// Trans2AddNodesRequest trans nodes request
func (ar AddNodesRequest) Trans2AddNodesRequest(clsId string, opt *cloudprovider.CommonOption) *model.AddNodeRequest {
	addNodesList := make([]model.AddNode, 0)

	for i := range ar.ServerIds {
		taints := Taint2ModelTaint(ar.Taints)

		addNodesList = append(addNodesList, model.AddNode{
			ServerID: ar.ServerIds[i],
			Spec: &model.ReinstallNodeSpec{
				Os:    ar.Os,
				Login: Login2ModelLogin(ar.Login),
				VolumeConfig: func() *model.ReinstallVolumeConfig {
					vConfig, err := GetAddNodesReinstallVolumeConfig(ar.ServerIds[i], opt)
					if err != nil {
						return nil
					}
					return vConfig
				}(),
				// RuntimeConfig:         nil,
				K8sOptions: func() *model.ReinstallK8sOptionsConfig {
					return &model.ReinstallK8sOptionsConfig{
						Labels:  ar.Labels,
						Taints:  &taints,
						MaxPods: &ar.MaxPods,
					}
				}(),
				Lifecycle: func() *model.NodeLifecycleConfig {
					life := &model.NodeLifecycleConfig{}
					if len(ar.PreScript) > 0 {
						life.PreInstall = common.StringPtr(ar.PreScript)
					}
					if len(ar.PostScript) > 0 {
						life.PostInstall = common.StringPtr(ar.PostScript)
					}

					return life
				}(),
				HostnameConfig: &model.HostnameConfig{Type: model.GetHostnameConfigTypeEnum().PRIVATE_IP},
			},
		})
	}

	return &model.AddNodeRequest{
		ClusterId: clsId,
		Body: &model.AddNodeList{
			ApiVersion: "v3",
			Kind:       "List",
			NodeList:   addNodesList,
		},
	}
}

// Volume for disk mount
type Volume struct {
	VolumeId    string
	ServerId    string
	MountPoint  string
	VolumeType  string
	Size        int32 // GB
	StorageType string

	IsCceDisk      bool
	Name           string
	RuntimeSize    int
	KubernetesSize int
	MountPath      string
}

// VolumeSlice volume info
type VolumeSlice []*Volume

// Len len()
func (v VolumeSlice) Len() int {
	return len(v)
}

// Less less()
func (v VolumeSlice) Less(i, j int) bool {
	return v[i].Size > v[j].Size
}

// Swap swap()
func (v VolumeSlice) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// GetServerVolumeInfo get server volume info
func GetServerVolumeInfo(serverId string, opt *cloudprovider.CommonOption) (*Volume, []*Volume, error) {
	ecs, err := NewEcsClient(opt)
	if err != nil {
		return nil, nil, err
	}

	blockDevices, err := ecs.ListServerBlockDevices(serverId)
	if err != nil {
		return nil, nil, err
	}

	// 无任何盘(系统盘/数据盘)
	if len(*blockDevices) == 0 {
		return nil, nil, nil
	}

	var (
		systemDisk *Volume
		dataDisks  []*Volume
	)

	devices := *blockDevices

	for i := range devices {

		vDetail, errLocal := GetVolumeDetailInfo(*devices[i].VolumeId, opt)
		if errLocal != nil {
			return nil, nil, errLocal
		}

		if devices[i].BootIndex != nil && *devices[i].BootIndex == 0 {
			systemDisk = &Volume{
				VolumeId:    *devices[i].VolumeId,
				MountPoint:  *devices[i].Device,
				ServerId:    *devices[i].ServerId,
				Size:        *devices[i].Size,
				VolumeType:  vDetail.VolumeType,
				StorageType: "evs",
			}
		} else {
			dataDisks = append(dataDisks, &Volume{
				VolumeId:    *devices[i].VolumeId,
				MountPoint:  *devices[i].Device,
				ServerId:    *devices[i].ServerId,
				Size:        *devices[i].Size,
				VolumeType:  vDetail.VolumeType,
				StorageType: "evs",
			})
		}
	}
	if len(dataDisks) == 0 {
		return nil, nil, nil
	}

	return systemDisk, dataDisks, nil
}

// GetVolumeDetailInfo get volume detailed info
func GetVolumeDetailInfo(volumeId string, opt *cloudprovider.CommonOption) (*evsModel.VolumeDetail, error) {
	evs, err := NewEvsClient(opt)
	if err != nil {
		return nil, err
	}

	return evs.ShowVolume(volumeId)
}

// GetStorageConfigByVolume xxx
func GetStorageConfigByVolume(v *Volume) (model.StorageSelectors, model.StorageGroups) {
	selector := model.StorageSelectors{
		Name:        v.Name,
		StorageType: v.StorageType,
		MatchLabels: &model.StorageSelectorsMatchLabels{
			Size:       common.StringPtr(fmt.Sprintf("%v", v.Size)),
			VolumeType: &v.VolumeType,
			Count:      common.StringPtr(fmt.Sprintf("%v", 1)),
		},
	}

	group := model.StorageGroups{
		Name: func() string {
			if v.IsCceDisk {
				return "vgpaas"
			}
			return utils.RandomString(8)
		}(),
		CceManaged: func() *bool {
			if v.IsCceDisk {
				return common.BoolPtr(true)
			}
			return nil
		}(),
		SelectorNames: []string{v.Name},
		/*
			virtualSpace的名称，当前仅支持三种类型：kubernetes、runtime、user。

			kubernetes：k8s空间配置，需配置lvmConfig；
			runtime：运行时空间配置，需配置runtimeConfig；
			user：用户空间配置，需配置lvmConfig
		*/
		VirtualSpaces: func() []model.VirtualSpace {
			var spaces = make([]model.VirtualSpace, 0)

			if v.IsCceDisk {
				spaces = append(spaces, model.VirtualSpace{
					Name: "runtime",
					Size: fmt.Sprintf("%d%%", v.RuntimeSize),
					// LvmConfig: &model.LvmConfig{LvType: "linear"},
					// runtime配置管理，适用于运行时空间配置。 需要注意：一个virtualSpace仅支持一个config配置。
					RuntimeConfig: &model.RuntimeConfig{LvType: "linear"},
				})
				spaces = append(spaces, model.VirtualSpace{
					Name: "kubernetes",
					Size: fmt.Sprintf("%d%%", v.KubernetesSize),
					// lvm配置管理，适用于kubernetes和user空间配置。 需要注意：一个virtualSpace仅支持一个config配置。
					LvmConfig: &model.LvmConfig{LvType: "linear"},
					//RuntimeConfig: &model.RuntimeConfig{LvType: "linear"},
				})

				return spaces
			}

			spaces = append(spaces, model.VirtualSpace{
				Name: "user",
				Size: "100%",
				LvmConfig: &model.LvmConfig{
					LvType: "linear",
					Path:   common.StringPtr(v.MountPath),
				},
			})
			return spaces
		}(),
	}

	return selector, group
}

func getServerFirstVolume(volumes []*Volume) *Volume {
	for i := range volumes {
		if volumes[i].MountPoint == "/dev/vdb" {
			return volumes[i]
		}
	}
	// 如果没有找到第一块磁盘，则选择最大的数据盘作为 k8s使用的数据盘
	volumesLocal := VolumeSlice(volumes)
	sort.Sort(&volumesLocal)

	return volumesLocal[0]
}

// GetAddNodesReinstallVolumeConfig 纳管节点时只操作第一块数据云盘 /dev/vdb
func GetAddNodesReinstallVolumeConfig(serverId string, opt *cloudprovider.CommonOption) (*model.ReinstallVolumeConfig, error) {
	_, dataDisks, err := GetServerVolumeInfo(serverId, opt)
	if err != nil || (len(dataDisks) == 0) {
		return nil, err
	}

	for i := range dataDisks {
		fmt.Println(*dataDisks[i])
	}

	// 获取server操作的数据盘
	v := getServerFirstVolume(dataDisks)

	var (
		storageSelector = make([]model.StorageSelectors, 0)
		storageGroups   = make([]model.StorageGroups, 0)
	)

	v.Name = "cce"
	v.IsCceDisk = true
	v.RuntimeSize = 70
	v.KubernetesSize = 30
	selectorLocal, groupLocal := GetStorageConfigByVolume(v)
	storageSelector = append(storageSelector, selectorLocal)
	storageGroups = append(storageGroups, groupLocal)

	volumeConfig := &model.ReinstallVolumeConfig{
		Storage: &model.Storage{
			StorageSelectors: storageSelector,
			StorageGroups:    storageGroups,
		},
	}

	fmt.Println(*volumeConfig)

	return volumeConfig, nil
}
