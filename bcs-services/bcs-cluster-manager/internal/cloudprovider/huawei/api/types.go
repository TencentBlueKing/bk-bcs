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
	"math"
	"sort"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	eipModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v2/model"
	evsModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/model"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	v1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
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

// CreateNodePoolRequest create node pool request
type CreateNodePoolRequest struct {
	// ClusterId 集群ID
	ClusterId string
	// Name 节点池名称
	Name string
	// Spec 节点池配置
	Spec CreateNodePoolSpec
}

// CreateNodePoolSpec create node pool spec
type CreateNodePoolSpec struct {
	Template CreateNodePoolTemplate
	// SecurityGroups 安全组ID列表
	SecurityGroups []string
	// SubnetId 子网ID
	SubnetId string
}

// CreateNodePoolTemplate create node pool template
type CreateNodePoolTemplate struct {
	// Flavor 节点规格
	Flavor string
	// Az 可用区
	Az string
	// Os 节点的操作系统类型
	Os string
	// Login 节点登录信息
	Login Login
	// RootVolume 节点系统盘配置
	RootVolume *Volume
	// DataVolumes 节点数据盘配置
	DataVolumes []*Volume
	// Charge 节点计费模式
	Charge ChargePrepaid
	// Taints 节点池taints
	Taints []v1.Taint
	// Labels 节点池labels
	Labels map[string]string
	// ContainerRuntime 节点运行时
	ContainerRuntime string
	// MaxPod 节点最大允许创建的实例数(Pod)
	MaxPod int32
	// PreScript 前置脚本
	PreScript string
	// PostScript 后置脚本
	PostScript string
	// PublicIp 节点公网ip配置
	PublicIp PublicIp
}

type ChargePrepaid struct {
	// ChargeType 节点池计费模式
	ChargeType string
	// Period 节点池计费周期
	Period uint32
	// RenewFlag 节点池自动续费标识
	RenewFlag string
}

type PublicIp struct {
	// ChargeType 节点公网计费模式
	ChangeType string
	// Bandwidth 节点公网带宽
	Bandwidth int32
	// Enable 节点公网是否启用
	Enable bool
}

func GetChargeConfig(charge ChargePrepaid) (billingMode int32, periodType string, periodNum int32,
	isAutoRenew string, isAutoPay string) {
	periodType = "month"
	periodNum = 1
	isAutoRenew = "false"
	isAutoPay = "true"

	periodNum = int32(charge.Period)
	if charge.ChargeType == icommon.PREPAID {
		billingMode = 1
		if charge.Period >= 12 {
			periodType = "year"
			periodNum = int32(charge.Period / 12)
		}
		if charge.RenewFlag == icommon.NOTIFYANDAUTORENEW {
			isAutoRenew = "true"
		}
	}
	return
}

func GetPublicIp(publicIp PublicIp) *model.NodePublicIp {
	if publicIp.Enable {
		shareType := eipModel.GetCreatePublicipBandwidthOptionShareTypeEnum().PER.Value()
		return &model.NodePublicIp{
			Iptype: common.StringPtr("5_bgp"),
			Bandwidth: &model.NodeBandwidth{
				Chargemode: func() *string {
					if publicIp.ChangeType == ChargemodeBandwidth {
						return common.StringPtr(ChargemodeBandwidth)
					}
					return common.StringPtr(ChargemodeTraffic)
				}(),
				Sharetype: &shareType,
				Size: func() *int32 {
					var size int32 = 2
					if publicIp.Bandwidth > 0 {
						size = publicIp.Bandwidth
					}
					return &size
				}(),
			},
		}
	}

	return nil
}

func GetRuntime(containerRuntime string) *model.Runtime {
	runtimeName := model.GetRuntimeNameEnum().CONTAINERD

	if containerRuntime == icommon.DockerContainerRuntime {
		runtimeName = model.GetRuntimeNameEnum().DOCKER
	}

	return &model.Runtime{
		Name: &runtimeName,
	}
}

func (req CreateNodePoolRequest) trans2NodePoolTemplate() *model.CreateNodePoolRequest {
	var (
		NodePoolSpecType = model.GetNodePoolSpecTypeEnum().VM
		taints           = Taint2ModelTaint(req.Spec.Template.Taints)
	)

	dataVolumes, storageSelectors, storageGroups := GetDataVolumeAndStorgeConfig(req.Spec.Template.DataVolumes)
	billingMode, periodType, periodNum, isAutoRenew, isAutoPay := GetChargeConfig(req.Spec.Template.Charge)
	publicIp := GetPublicIp(req.Spec.Template.PublicIp)

	return &model.CreateNodePoolRequest{
		ClusterId: req.ClusterId,
		Body: &model.NodePool{
			Kind:       "NodePool",
			ApiVersion: "v3",
			Metadata: &model.NodePoolMetadata{
				Name: req.Name,
			},
			Spec: &model.NodePoolSpec{
				Type: &NodePoolSpecType,
				NodeTemplate: &model.NodeSpec{
					Flavor: req.Spec.Template.Flavor,
					Az:     req.Spec.Template.Az,
					Os:     &req.Spec.Template.Os,
					Login:  Login2ModelLogin(&req.Spec.Template.Login),
					RootVolume: &model.Volume{
						Size:       req.Spec.Template.RootVolume.Size,
						Volumetype: req.Spec.Template.RootVolume.VolumeType,
					},
					DataVolumes: dataVolumes,
					Storage: &model.Storage{
						StorageSelectors: storageSelectors,
						StorageGroups:    storageGroups,
					},
					BillingMode: &billingMode,
					Taints:      &taints,
					K8sTags:     req.Spec.Template.Labels,
					Runtime:     GetRuntime(req.Spec.Template.ContainerRuntime),
					InitializedConditions: &[]string{
						"NodeInitial", // 新增节点调度策略: 设置为不可调度
					},
					ExtendParam: &model.NodeExtendParam{
						MaxPods:            &req.Spec.Template.MaxPod,
						PeriodType:         &periodType,
						PeriodNum:          &periodNum,
						IsAutoRenew:        &isAutoRenew,
						IsAutoPay:          &isAutoPay,
						AlphaCcePreInstall: &req.Spec.Template.PreScript,
						//AlphaCcePostInstall: &req.Spec.Template.PostScript, 后置脚本由蓝鲸的job任务执行
					},
					NodeNicSpec: &model.NodeNicSpec{
						PrimaryNic: &model.NicSpec{SubnetId: &req.Spec.SubnetId},
					},
					HostnameConfig: &model.HostnameConfig{
						Type: model.GetHostnameConfigTypeEnum().PRIVATE_IP, // 节点名称默认与节点私有ip保持一致
					},
					PublicIP: publicIp,
				},
				CustomSecurityGroups: func() *[]string {
					securityIds := make([]string, 0)
					for _, v := range req.Spec.SecurityGroups {
						id := v
						securityIds = append(securityIds, id)
					}
					return &securityIds
				}(),
			},
		},
	}
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

func GetDataVolumeAndStorgeConfig(volumes []*Volume) ([]model.Volume, []model.StorageSelectors, []model.StorageGroups) {
	var (
		dataVolumes       = make([]model.Volume, 0)
		storageSelectors  = make([]model.StorageSelectors, 0)
		storageGroups     = make([]model.StorageGroups, 0)
		metadataEncrypted = "0"
		matchCount        = "1"
		cceManaged        = true
	)

	for k, v := range volumes {
		dataVolumes = append(dataVolumes, model.Volume{
			Volumetype: v.VolumeType,
			Size:       v.Size,
		})

		selectorName := fmt.Sprintf("selector%d", k)
		size := fmt.Sprintf("%d", v.Size)
		storageSelectors = append(storageSelectors, model.StorageSelectors{
			Name:        selectorName,
			StorageType: "evs",
			MatchLabels: &model.StorageSelectorsMatchLabels{
				Size:              &size,
				VolumeType:        &v.VolumeType,
				MetadataEncrypted: &metadataEncrypted,
				Count:             &matchCount,
			},
		})

		if k == 0 {
			storageGroups = append(storageGroups, model.StorageGroups{
				Name:          "vgpaas", // 当cceManaged=ture时，name必须为：vgpaas
				SelectorNames: []string{selectorName},
				CceManaged:    &cceManaged, // k8s及runtime所属存储空间。有且仅有一个group被设置为true，不填默认false
				VirtualSpaces: []model.VirtualSpace{
					{
						Name: "kubernetes",
						Size: "10%",
						LvmConfig: &model.LvmConfig{
							LvType: "linear",
						},
					},
					{
						Name: "runtime",
						Size: "90%",
						RuntimeConfig: &model.RuntimeConfig{
							LvType: "linear",
						},
					},
				},
			})
		} else {
			storageGroup := model.StorageGroups{
				Name:          fmt.Sprintf("group%d", k),
				SelectorNames: []string{selectorName},
				VirtualSpaces: []model.VirtualSpace{
					{
						Name: "user",
						Size: "100%",
						LvmConfig: &model.LvmConfig{
							LvType: "linear",
							Path:   &v.MountPath,
						},
					},
				},
			}
			if v.MountPath != "" {
				storageGroup.VirtualSpaces[0].LvmConfig.Path = &v.MountPath
			}
			storageGroups = append(storageGroups, storageGroup)
		}
	}

	return dataVolumes, storageSelectors, storageGroups
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

// CreateClusterRequest create cluster request
type CreateClusterRequest struct {
	// Name 集群名称
	Name string
	// Spec 集群配置
	Spec CreateClusterSpec
}

func (c *CreateClusterRequest) Trans2CreateClusterRequest() *model.CreateClusterRequest {
	category := model.GetClusterSpecCategoryEnum().TURBO
	clusterType := model.GetClusterSpecTypeEnum().VIRTUAL_MACHINE

	billingMode, periodType, periodNum, isAutoRenew, isAutoPay := GetChargeConfig(c.Spec.Charge)
	// 根据 单节点 Pod 数量上限 算出 容器网络固定IP池掩码位数
	alphaCceFixPoolMask := fmt.Sprintf("%d", 32-int(math.Log2(float64(c.Spec.AlphaCceFixPoolMask))))

	clusterTags := make([]model.ResourceTag, 0)
	for k, v := range c.Spec.ClusterTag {
		tmpK := k
		tmpV := v
		clusterTags = append(clusterTags, model.ResourceTag{Key: &tmpK, Value: &tmpV})
	}

	confOverName := "kube-apiserver"
	confName := "support-overload"
	var confValue interface{} = true

	req := &model.CreateClusterRequest{
		Body: &model.Cluster{
			Kind:       "Cluster",
			ApiVersion: "v3",
			Metadata: &model.ClusterMetadata{
				Name: c.Name,
				Annotations: map[string]string{
					ClusterInstallAddonsExternalInstall: ClusterInstallAddonsExternalInstallValue,
					ClusterInstallAddonsInstall:         ClusterInstallAddonsInstallValue,
				},
			},
			Spec: &model.ClusterSpec{
				Category:    &category,
				Type:        &clusterType,
				Flavor:      c.Spec.Flavor,
				Version:     &c.Spec.Version,
				Description: &c.Spec.Description,
				Ipv6enable:  &c.Spec.Ipv6Enable,
				HostNetwork: &model.HostNetwork{
					Vpc:           c.Spec.VpcID,
					Subnet:        c.Spec.SubnetID,
					SecurityGroup: &c.Spec.SecurityGroupID,
				},
				ServiceNetwork: &model.ServiceNetwork{IPv4CIDR: &c.Spec.ServiceCidr},
				BillingMode:    &billingMode,
				ClusterTags:    &clusterTags,
				KubeProxyMode: func() *model.ClusterSpecKubeProxyMode {
					proxyMode := model.GetClusterSpecKubeProxyModeEnum().IPTABLES
					if c.Spec.KubeProxyMode == model.GetClusterSpecKubeProxyModeEnum().IPVS.Value() {
						proxyMode = model.GetClusterSpecKubeProxyModeEnum().IPVS
					}
					return &proxyMode

				}(),
				ExtendParam: &model.ClusterExtendParam{
					PeriodType:  &periodType,
					PeriodNum:   &periodNum,
					IsAutoRenew: &isAutoRenew,
					IsAutoPay:   &isAutoPay,
				},
				ConfigurationsOverride: &[]model.PackageConfiguration{
					{
						Name: &confOverName,
						Configurations: &[]model.ConfigurationItem{
							{Name: &confName, Value: &confValue},
						},
					},
				},
			},
		},
	}

	if len(c.Spec.Az) == 1 {
		req.Body.Spec.ExtendParam.ClusterAZ = &c.Spec.Az[0]
	} else if len(c.Spec.Az) == 3 {
		clusterAz := "multi_az"
		req.Body.Spec.ExtendParam.ClusterAZ = &clusterAz
		masters := make([]model.MasterSpec, 0)
		for _, az := range c.Spec.Az {
			tmp := az
			masters = append(masters, model.MasterSpec{AvailabilityZone: &tmp})
		}
		req.Body.Spec.Masters = &masters
	}

	if len(c.Spec.PublicIP) > 0 {
		req.Body.Spec.ExtendParam.ClusterExternalIP = &c.Spec.PublicIP
	}

	if c.Spec.Category == model.GetClusterSpecCategoryEnum().CCE.Value() {
		category = model.GetClusterSpecCategoryEnum().CCE
		containerCidr := make([]model.ContainerCidr, 0)
		for _, cidr := range c.Spec.ContainerCidr {
			containerCidr = append(containerCidr, model.ContainerCidr{Cidr: cidr})
		}
		req.Body.Spec.Category = &category
		req.Body.Spec.ContainerNetwork = &model.ContainerNetwork{
			Mode: func() model.ContainerNetworkMode {
				mode := model.GetContainerNetworkModeEnum().OVERLAY_L2
				if c.Spec.ContainerMode == model.GetContainerNetworkModeEnum().VPC_ROUTER.Value() {
					mode = model.GetContainerNetworkModeEnum().VPC_ROUTER
				}
				return mode
			}(),
			Cidrs: &containerCidr,
		}
		req.Body.Spec.ExtendParam.AlphaCceFixPoolMask = &alphaCceFixPoolMask
	} else {
		networkSubnet := make([]model.NetworkSubnet, 0)
		for _, subnetId := range c.Spec.EniNetworkSubnet {
			networkSubnet = append(networkSubnet, model.NetworkSubnet{
				SubnetID: subnetId,
			})
		}
		req.Body.Spec.ContainerNetwork = &model.ContainerNetwork{
			Mode: model.GetContainerNetworkModeEnum().ENI,
		}
		req.Body.Spec.EniNetwork = &model.EniNetwork{
			Subnets: networkSubnet,
		}
	}

	return req
}

// CreateClusterSpec create cluster spec
type CreateClusterSpec struct {
	// Category 集群类别
	Category string
	// Az 可用区
	Az []string
	// Flavor 节点规格
	Flavor string
	// Version 集群版本
	Version string
	// Description 集群描述
	Description string
	// VpcID vpc ID
	VpcID string
	// SubnetID 子网ID
	SubnetID string
	// SecurityGroupID 安全组ID
	SecurityGroupID string
	// ContainerMode 容器网络类型
	ContainerMode string
	// ContainerCidr 容器网段
	ContainerCidr []string
	// ServiceCidr 服务网段
	ServiceCidr string
	// Charge 节点计费模式
	Charge ChargePrepaid
	// Ipv6Enable 是否开启ipv6
	Ipv6Enable bool
	// AlphaCceFixPoolMask 容器网络固定IP池掩码位数
	AlphaCceFixPoolMask uint32
	// KubeProxyMode 服务转发模式
	KubeProxyMode string
	// ClusterTag 集群标签
	ClusterTag map[string]string
	// EniNetworkSubnet IPv4子网ID列表
	EniNetworkSubnet []string
	// PublicIP 公网ip地址
	PublicIP string
}
