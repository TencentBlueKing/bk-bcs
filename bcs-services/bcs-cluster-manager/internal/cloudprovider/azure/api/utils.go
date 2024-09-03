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

package api

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
)

//	nodeGroupToPool agentPool 转换器
//
// bcs nodeGroup 转换为 aks agentPool
type nodeGroupToPool struct {
	group *proto.NodeGroup
	pool  *armcontainerservice.AgentPool
}

// newNodeGroupToAgentPoolConverter create nodeGroupToPool
func newNodeGroupToAgentPoolConverter(group *proto.NodeGroup, pool *armcontainerservice.AgentPool) *nodeGroupToPool {
	return &nodeGroupToPool{
		group: group,
		pool:  pool,
	}
}

// convert 转换
func (c *nodeGroupToPool) convert() {
	if c.pool.Properties == nil {
		c.pool.Properties = new(armcontainerservice.ManagedClusterAgentPoolProfileProperties)
	}
	properties := c.pool.Properties
	// 以下为必填参数
	// 设置名称
	c.pool.Name = to.Ptr(c.group.CloudNodeGroupID)
	// 设置机型
	properties.VMSize = to.Ptr(c.group.LaunchTemplate.InstanceType)
	// 当前节点池为手动扩缩容模式，不可以指定最大、最小节点数!!!
	// 设置最大节点数
	// properties.MaxCount = to.Ptr(int32(c.group.AutoScaling.MaxSize))
	// 设置最小节点数
	// properties.MinCount = to.Ptr(int32(c.group.AutoScaling.MinSize))
	// 设置节点池大小
	properties.Count = to.Ptr(int32(c.group.AutoScaling.DesiredSize))
	// 设置节点池为用户模式
	properties.Mode = to.Ptr(armcontainerservice.AgentPoolModeUser)
	// 设置Azure节点池的弹性伸缩
	properties.EnableAutoScaling = to.Ptr(false)
	// 是否分配公网IP
	if c.group.LaunchTemplate.InternetAccess != nil {
		properties.EnableNodePublicIP = to.Ptr(c.group.LaunchTemplate.InternetAccess.PublicIPAssigned)

	}

	// 以下为可选参数

	// 设置tags
	c.setTags()
	// 设置labels
	c.setLabels()
	// 设置taints
	c.setTaints()
	// 设置每一个节点的最大pod数量
	c.setMaxPods()
	// 设置系统和机型
	c.setOSAndInstanceType()
	// 设置节点系统盘类型
	c.setOSDiskType()
	// 设置节点系统盘大小
	c.setOSDiskSizeGB()
	// 设置区域
	c.setAvailabilityZones()
	// 设置扩容模式
	c.setScalingMode()
	// note:
	// 1.kubelet参数设置
	// 2."pool.NodeImageVersion"为只读字段，无法指定节点的镜像
	// 3.目前，对k8s版本 暂无变更需求 , b.setOrchestratorVersion()
	// 4.目前，不持支设置GPU , setGpuInstanceProfile(pool,ng.LaunchTemplate)
}

// setOSDiskType 设置扩容模式
func (c *nodeGroupToPool) setScalingMode() {
	asg := c.group.AutoScaling
	if len(asg.ScalingMode) == 0 {
		return
	}
	found := false
	for _, mode := range armcontainerservice.PossibleScaleDownModeValues() {
		if string(mode) == asg.ScalingMode {
			found = true
			break
		}
	}
	if !found {
		return
	}
	c.pool.Properties.ScaleDownMode = to.Ptr(armcontainerservice.ScaleDownMode(asg.ScalingMode))
}

// setOSDiskType 设置系统盘类型
func (c *nodeGroupToPool) setOSDiskType() {
	// 默认 使用托管类型(Managed)
	c.pool.Properties.OSDiskType = to.Ptr(armcontainerservice.OSDiskTypeManaged)
	lc := c.group.LaunchTemplate
	if lc.SystemDisk == nil || len(lc.SystemDisk.DiskType) == 0 {
		return
	}
	found := false
	for _, osDiskType := range armcontainerservice.PossibleOSDiskTypeValues() {
		if string(osDiskType) == lc.SystemDisk.DiskType {
			found = true
			break
		}
	}
	if !found {
		return
	}
	c.pool.Properties.OSDiskType = to.Ptr(armcontainerservice.OSDiskType(lc.SystemDisk.DiskType))
}

// setOSDiskSizeGB 设置系统盘大小
func (c *nodeGroupToPool) setOSDiskSizeGB() {
	// 创建节点池时，OS磁盘默认大小由 vCPU 数确定
	// 1至7核 - 128G
	// 8至15核 - 256G
	// 16至63核 - 512G
	// 64+核 - 1024G
	// Azure节点系统盘 默认为128GB
	sysDisk := c.group.LaunchTemplate.SystemDisk
	if sysDisk == nil || len(sysDisk.DiskSize) == 0 {
		return
	}
	// 自定义节点系统盘 大小
	size, err := strconv.ParseInt(sysDisk.DiskSize, 10, 32)
	if err != nil {
		return
	}
	c.pool.Properties.OSDiskSizeGB = to.Ptr(int32(size))
}

// setOSAndInstanceType  设置系统和机型
func (c *nodeGroupToPool) setOSAndInstanceType() {
	nodeOS := c.group.NodeOS
	// 默认系统类型为linux
	c.pool.Properties.OSType = to.Ptr(armcontainerservice.OSTypeLinux)
	if strings.Contains(nodeOS, winTypeOS) {
		c.pool.Properties.OSType = to.Ptr(armcontainerservice.OSTypeWindows)
	}
	// 默认OS为Ubuntu
	c.pool.Properties.OSSKU = to.Ptr(armcontainerservice.OSSKUUbuntu)
	if c.group.LaunchTemplate != nil && c.group.LaunchTemplate.ImageInfo != nil {
		c.pool.Properties.OSSKU = to.Ptr(armcontainerservice.OSSKU(c.group.LaunchTemplate.ImageInfo.ImageName))
	}
	if c.group.LaunchTemplate != nil {
		c.pool.Properties.VMSize = to.Ptr(c.group.LaunchTemplate.InstanceType)
	}
}

// setOrchestratorVersion 设置k8s版本
func (c *nodeGroupToPool) setOrchestratorVersion() { // nolint
	// BCS暂无设置K8S版本需求
	// runTimeInfo := b.group.NodeTemplate.Runtime
	// 默认为空
	// b.pool.Properties.OrchestratorVersion = to.Ptr("")
	// 若有运行时，则按Runtime版本
	// if runTimeInfo != nil {
	//	b.pool.Properties.OrchestratorVersion = to.Ptr(runTimeInfo.RuntimeVersion)
	// }
}

// setMaxPods 设置每一个节点的最大pod数量
func (c *nodeGroupToPool) setMaxPods() {
	// 根据解析参数获得
	c.pool.Properties.MaxPods = to.Ptr(int32(c.group.NodeTemplate.MaxPodsPerNode))
	// 默认为250个
	if *c.pool.Properties.MaxPods == 0 {
		c.pool.Properties.MaxPods = to.Ptr(int32(250))
	}
}

// setTags 设置tags
func (c *nodeGroupToPool) setTags() {
	tags := c.group.Tags
	if tags == nil {
		return
	}
	c.pool.Properties.Tags = make(map[string]*string)
	for key := range tags {
		c.pool.Properties.Tags[key] = to.Ptr(tags[key])
	}
}

// setLabels 设置labels
func (c *nodeGroupToPool) setLabels() {
	labels := c.group.NodeTemplate.Labels
	if labels == nil {
		return
	}
	c.pool.Properties.NodeLabels = make(map[string]*string)
	for key := range labels {
		c.pool.Properties.NodeLabels[key] = to.Ptr(labels[key])
	}
}

// setTaints 设置taints
func (c *nodeGroupToPool) setTaints() {
	taints := c.group.NodeTemplate.Taints
	if taints == nil {
		taints = make([]*proto.Taint, 0)
	}

	/*
		// attention: azure not support addNodes to set unScheduled nodes, thus realize this feature by taint
		taints = append(taints, &proto.Taint{
			Key:    cutils.BCSNodeGroupTaintKey,
			Value:  cutils.BCSNodeGroupTaintValue,
			Effect: cutils.BCSNodeGroupAzureTaintEffect,
		})
	*/

	// key=value:NoSchedule NoExecute PreferNoSchedule
	c.pool.Properties.NodeTaints = make([]*string, 0)
	target := &c.pool.Properties.NodeTaints
	for _, taint := range taints {
		*target = append(*target, to.Ptr(fmt.Sprintf("%s=%s:%s", taint.Key, taint.Value, taint.Effect)))
	}
}

// setAvailabilityZones 设置可用性区域
// 若当前地域，支持可用性，则该地域至少会有三个可用性区域；若不支持可用性，则没有可用性区域.
// 参考：https://docs.azure.cn/zh-cn/availability-zones/az-overview#availability-zones
func (c *nodeGroupToPool) setAvailabilityZones() {
	as := c.group.AutoScaling
	if len(as.Zones) == 0 {
		return
	}
	c.pool.Properties.AvailabilityZones = make([]*string, 0)
	target := &c.pool.Properties.AvailabilityZones
	for i := range as.Zones {
		*target = append(*target, to.Ptr(as.Zones[i]))
	}
}

// getMaxPods 获取当前节点配置模板的最大pod数量
func (c *nodeGroupToPool) getMaxPods(extra map[string]string) int32 {
	if extra == nil {
		return 0
	}
	// kubelet params
	kp, ok := extra[kubeletType]
	if !ok {
		return 0
	}
	return c.regexpMaxPods(kp)
}

// regexpMaxPods 通过正则解析 最大pod数量
// 在kp字符串中，无法保证max-pods后面一定有';',示例如下：
// "xxxx;max-pods=200"
// "xxxx;max-pods=200;yyyy"
func (c *nodeGroupToPool) regexpMaxPods(kp string) int32 {
	reg := regexp.MustCompile("max-pods=(.*)$")
	result := reg.FindStringSubmatch(kp)
	if len(result) <= 1 || len(result[1]) == 0 {
		return 0
	}

	idx := 0
	for _, r := range result[1] {
		if r < '0' || r > '9' {
			break
		}
		idx++
	}

	n, err := strconv.ParseInt(result[1][0:idx], 10, 64)
	if err != nil {
		return 0
	}
	return int32(n)
}

// poolToNodeGroup AKS agentPool 转换为 BCS	NodeGroup
type poolToNodeGroup struct {
	group      *proto.NodeGroup
	pool       *armcontainerservice.AgentPool
	properties *armcontainerservice.ManagedClusterAgentPoolProfileProperties
}

// newPoolToNodeGroupConverter create poolToNodeGroup
func newPoolToNodeGroupConverter(pool *armcontainerservice.AgentPool, group *proto.NodeGroup) *poolToNodeGroup {
	if pool.Properties == nil {
		pool.Properties = new(armcontainerservice.ManagedClusterAgentPoolProfileProperties)
	}
	return &poolToNodeGroup{
		group:      group,
		pool:       pool,
		properties: pool.Properties,
	}
}

// convert 转换
func (c *poolToNodeGroup) convert() {
	if c.group.AutoScaling == nil {
		c.group.AutoScaling = new(proto.AutoScalingGroup)
	}
	if c.group.LaunchTemplate == nil {
		c.group.LaunchTemplate = new(proto.LaunchConfiguration)
	}
	if c.group.NodeTemplate == nil {
		c.group.NodeTemplate = new(proto.NodeTemplate)
	}
	// 设置tags
	c.setTags()
	// 设置状态
	// c.setStatus()
	// 设置CloudNodeGroupID
	c.setCloudNodeGroupID()
	// 设置NodeGroupID
	// c.setNodeGroupID()

	// 设置期望节点数量
	c.setCount()
	// 设置最大节点数
	c.setMaxSize()
	// 设置最小节点数
	c.setMinSize()
	// 设置扩缩容模式
	c.setScaleDownMode()
	// 设置机型
	c.setInstanceType()

	// 设置磁盘类型
	c.setOsDiskType()
	// 设置磁盘大小
	c.setOsDiskSizeGB()
	// 设置节点的镜像版本
	c.setNodeImageVersion()

	// 设置labels
	c.setLabels()
	// 设置taints
	c.setTaints()
	// 设置k8s版本
	// c.setCurrentOrchestratorVersion()
}

// setTags 设置tags
func (c *poolToNodeGroup) setTags() {
	tags := c.properties.Tags
	if tags == nil {
		return
	}
	c.group.Tags = make(map[string]string)
	for key := range tags {
		c.group.Tags[key] = *tags[key]
	}
}

// setMaxSize 设置最大节点数
func (c *poolToNodeGroup) setMaxSize() {
	asg := c.group.AutoScaling
	size := c.properties.MaxCount
	if size == nil {
		return
	}
	asg.MaxSize = uint32(*size)
}

// setMinSize 设置最小节点数
func (c *poolToNodeGroup) setMinSize() {
	asg := c.group.AutoScaling
	size := c.properties.MinCount
	if size == nil {
		return
	}
	asg.MinSize = uint32(*size)
}

// setScaleDownMode 设置扩缩容模式
func (c *poolToNodeGroup) setScaleDownMode() {
	asg := c.group.AutoScaling
	mode := c.properties.ScaleDownMode
	if mode == nil {
		return
	}
	switch *mode {
	case armcontainerservice.ScaleDownModeDelete:
		asg.ScalingMode = "CLASSIC_SCALING"
	case armcontainerservice.ScaleDownModeDeallocate:
		asg.ScalingMode = "WAKE_UP_STOPPED_SCALING"
	}
}

// setOsDiskType 设置磁盘类型
func (c *poolToNodeGroup) setOsDiskType() {
	// lc := c.group.LaunchTemplate
	//  diskType := c.properties.OSDiskType
	// if lc.SystemDisk == nil {
	//	lc.SystemDisk = new(proto.DataDisk)
	// }
	// lc.SystemDisk.DiskType = string(*diskType)
}

// setOsDiskSizeGB 设置磁盘大小
func (c *poolToNodeGroup) setOsDiskSizeGB() {
	lc := c.group.LaunchTemplate
	size := strconv.FormatInt(int64(*c.properties.OSDiskSizeGB), 10)
	if lc.SystemDisk == nil {
		lc.SystemDisk = new(proto.DataDisk)
	}
	lc.SystemDisk.DiskSize = size
}

// setNodeImageVersion 设置节点的镜像版本
func (c *poolToNodeGroup) setNodeImageVersion() {
	// 需要通过镜像接口拉取
	// lc := ap.group.LaunchTemplate
	// imageVersion := ap.properties.NodeImageVersion
	// if lc.ImageInfo == nil {
	//	lc.ImageInfo = new(proto.ImageInfo)
	// }
	// lc.ImageInfo.ImageName = *imageVersion
}

// setLabels 设置labels
func (c *poolToNodeGroup) setLabels() {
	nt := c.group.NodeTemplate
	labels := c.properties.NodeLabels
	if labels == nil || nt == nil {
		return
	}
	nt.Labels = make(map[string]string)
	for key := range labels {
		nt.Labels[key] = *labels[key]
	}
}

// setTaints 设置taints
func (c *poolToNodeGroup) setTaints() {
	nt := c.group.NodeTemplate
	taints := c.properties.NodeTaints
	if len(taints) == 0 || nt == nil {
		return
	}
	nt.Taints = make([]*proto.Taint, 0)
	for _, taint := range taints {
		nt.Taints = append(nt.Taints, c.buildTaint(taint))
	}
}

// buildTaint 解析taint
func (c *poolToNodeGroup) buildTaint(taint *string) (res *proto.Taint) {
	res = &proto.Taint{}
	v := strings.Split(*taint, "=")
	if len(v) == 0 {
		return res
	}
	res.Key = v[0]
	if len(v) <= 1 {
		return res
	}
	v = strings.Split(v[1], ":")
	if len(v) == 0 {
		return res
	}
	res.Value = v[0]
	if len(v) <= 1 {
		return res
	}
	res.Effect = v[1]
	return res
}

// setCurrentOrchestratorVersion 设置k8s版本
func (c *poolToNodeGroup) setCurrentOrchestratorVersion() { // nolint
	// nt := ap.group.NodeTemplate
	// version := ap.properties.CurrentOrchestratorVersion
	// if nt == nil || nt.Runtime == nil {
	//	return
	// }
	// nt.Runtime.RuntimeVersion = *version
}

// setStatus 设置状态
func (c *poolToNodeGroup) setStatus() {
	switch *c.properties.ProvisioningState {
	case NormalState:
		c.group.Status = NodeGroupLifeStateNormal
	case CreatingState:
		c.group.Status = NodeGroupLifeStateCreating
	case UpdatingState:
		c.group.Status = NodeGroupLifeStateUpdating
	default:
		c.group.Status = strings.ToLower(*c.properties.ProvisioningState)
	}
}

// setCloudNodeGroupID 设置CloudNodeGroupID
func (c *poolToNodeGroup) setCloudNodeGroupID() {
	if len(*c.pool.Name) == 0 {
		return
	}
	c.group.CloudNodeGroupID = *c.pool.Name
}

// setNodeGroupID 设置NodeGroupID
func (c *poolToNodeGroup) setNodeGroupID() { // nolint
	// if len(*c.pool.Name) == 0 {
	//	return
	// }
	// c.group.NodeGroupID = *c.pool.Name
}

// setCount 设置数量
func (c *poolToNodeGroup) setCount() {
	c.group.AutoScaling.DesiredSize = uint32(*c.properties.Count)
}

// setInstanceType 设置机型
func (c *poolToNodeGroup) setInstanceType() {
	c.group.LaunchTemplate.InstanceType = *c.properties.VMSize
}

// setToNodeGroup AKS VirtualMachineScaleSet 转到 BCS NodeGroup
type setToNodeGroup struct {
	group *proto.NodeGroup
	set   *armcompute.VirtualMachineScaleSet
}

// newSetToNodeGroupConverter  create setToNodeGroup
func newSetToNodeGroupConverter(set *armcompute.VirtualMachineScaleSet, ng *proto.NodeGroup) *setToNodeGroup {
	return &setToNodeGroup{
		group: ng,
		set:   set,
	}
}

// convert 转换
func (c *setToNodeGroup) convert() {
	if c.group.AutoScaling == nil {
		c.group.AutoScaling = new(proto.AutoScalingGroup)
	}
	if c.group.LaunchTemplate == nil {
		c.group.LaunchTemplate = new(proto.LaunchConfiguration)
	}
	if c.group.NodeTemplate == nil {
		c.group.NodeTemplate = new(proto.NodeTemplate)
	}
	// 设置区域
	c.setRegion()
	// 设置asg id
	c.setAutoScalingID()
	// 设置asg name
	c.setAutoScalingName()
	// 系统盘
	c.setSystemDisk()
	// 数据盘
	c.setDataDisks()
	// 设置可用性区域
	c.setZones()
	// 设置用户名
	c.setUsername()
}

// setLaunchConfigureName 设置lc name
func (c *setToNodeGroup) setLaunchConfigureName() { // nolint
	// lc := c.group.LaunchTemplate
	// if c.set.Name == nil || lc == nil {
	//	return
	// }
	// lc.LaunchConfigureName = *c.set.Name
}

// setLaunchConfigurationID 设置lc ID
func (c *setToNodeGroup) setLaunchConfigurationID() { // nolint
	// lc := c.group.LaunchTemplate
	// if c.set.ID == nil || lc == nil {
	//	return
	// }
	// lc.LaunchConfigureName = *c.set.ID
}

// setAutoScalingID 设置asg id
func (c *setToNodeGroup) setAutoScalingID() {
	asg := c.group.AutoScaling
	asg.AutoScalingID = *c.set.Name
}

// setAutoScalingName 设置asg name
func (c *setToNodeGroup) setAutoScalingName() {
	asg := c.group.AutoScaling
	asg.AutoScalingName = RegexpSetNodeGroupResourcesName(c.set)
}

// setZones 设置可用性区域
func (c *setToNodeGroup) setZones() {
	zones := c.set.Zones
	as := c.group.AutoScaling
	if len(zones) == 0 {
		return
	}
	as.Zones = make([]string, 0)
	for i := range zones {
		as.Zones = append(as.Zones, *zones[i])
	}
}

// setRegion 设置地区
func (c *setToNodeGroup) setRegion() {
	if c.set.Location != nil {
		c.group.Region = *c.set.Location
	}
}

// setSystemDisk 系统盘
func (c *setToNodeGroup) setSystemDisk() {
	lc := c.group.LaunchTemplate
	if lc.SystemDisk == nil {
		lc.SystemDisk = new(proto.DataDisk)
	}
	if c.set.Properties == nil || c.set.Properties.VirtualMachineProfile == nil || c.set.Properties.VirtualMachineProfile.
		StorageProfile == nil || c.set.Properties.VirtualMachineProfile.StorageProfile.OSDisk == nil {
		return
	}

	osDisk := c.set.Properties.VirtualMachineProfile.StorageProfile.OSDisk
	if osDisk.ManagedDisk != nil && osDisk.ManagedDisk.StorageAccountType != nil {
		lc.SystemDisk.DiskType = string(*osDisk.ManagedDisk.StorageAccountType)
	}
	lc.SystemDisk.DiskSize = strconv.Itoa(int(*osDisk.DiskSizeGB))
}

// setDataDisks 数据盘
func (c *setToNodeGroup) setDataDisks() {
	lc := c.group.LaunchTemplate
	nt := c.group.NodeTemplate

	lc.DataDisks = make([]*proto.DataDisk, 0)
	nt.DataDisks = make([]*proto.CloudDataDisk, 0)
	if c.set.Properties == nil || c.set.Properties.VirtualMachineProfile == nil || c.set.Properties.VirtualMachineProfile.
		StorageProfile == nil || c.set.Properties.VirtualMachineProfile.StorageProfile.OSDisk == nil {
		return
	}

	dataDisks := c.set.Properties.VirtualMachineProfile.StorageProfile.DataDisks
	if len(dataDisks) == 0 {
		return
	}
	for _, disk := range dataDisks {
		d := new(proto.DataDisk)
		if disk.ManagedDisk != nil {
			d.DiskType = string(*disk.ManagedDisk.StorageAccountType)
		}
		d.DiskSize = strconv.Itoa(int(*disk.DiskSizeGB))
		lc.DataDisks = append(lc.DataDisks, d)

		nt.DataDisks = append(nt.DataDisks, &proto.CloudDataDisk{
			DiskType: d.DiskType,
			DiskSize: d.DiskSize,
		})
	}
}

// setUsername 设置用户名
func (c *setToNodeGroup) setUsername() {
	if c.set.Properties == nil || c.set.Properties.VirtualMachineProfile == nil ||
		c.set.Properties.VirtualMachineProfile.OSProfile == nil {
		return
	}
	if c.set.Properties.VirtualMachineProfile.OSProfile.LinuxConfiguration != nil &&
		*c.set.Properties.VirtualMachineProfile.OSProfile.LinuxConfiguration.DisablePasswordAuthentication {
		// 当前该虚拟规模集未设置用户名与密码，使用的是默认密码
		return
	}
	if c.set.Properties.VirtualMachineProfile.OSProfile.AdminUsername != nil &&
		len(*c.set.Properties.VirtualMachineProfile.OSProfile.AdminUsername) != 0 {
		c.group.LaunchTemplate.InitLoginUsername = *c.set.Properties.VirtualMachineProfile.OSProfile.AdminUsername
	}
}

// nodeGroupToSet BCS NodeGroup 转换为 AKS VirtualMachineScaleSet
type nodeGroupToSet struct {
	group *proto.NodeGroup
	set   *armcompute.VirtualMachineScaleSet
}

// newNodeGroupToSetConverter new nodeGroupToSet
func newNodeGroupToSetConverter(ng *proto.NodeGroup, set *armcompute.VirtualMachineScaleSet) *nodeGroupToSet {
	return &nodeGroupToSet{
		group: ng,
		set:   set,
	}
}

func (c *nodeGroupToSet) convert() {
	set := c.set
	if set.Properties == nil {
		set.Properties = new(armcompute.VirtualMachineScaleSetProperties)
	}
	if set.Properties.VirtualMachineProfile == nil {
		set.Properties.VirtualMachineProfile = new(armcompute.VirtualMachineScaleSetVMProfile)
	}
	// 用户数据
	c.setUserData()
	// 设置区域
	c.setLocation()
	// 系统盘
	c.setSystemDisk()
	// 数据盘
	c.setDataDisks()
	// 设置可用性区域
	c.setZones()
}

// setUserData  用户数据
func (c *nodeGroupToSet) setUserData() {
	set := c.set
	lc := c.group.LaunchTemplate
	if lc == nil || len(lc.UserData) == 0 {
		return
	}
	if set.Properties == nil {
		set.Properties = new(armcompute.VirtualMachineScaleSetProperties)
	}
	if set.Properties.VirtualMachineProfile == nil {
		set.Properties.VirtualMachineProfile = new(armcompute.VirtualMachineScaleSetVMProfile)
	}
	set.Properties.VirtualMachineProfile.UserData = to.Ptr(lc.UserData)
}

// setLocation 设置区域
func (c *nodeGroupToSet) setLocation() {
	if len(c.group.Region) != 0 {
		c.set.Location = to.Ptr(c.group.Region)
	}
}

// setZones 设置可用性区域
func (c *nodeGroupToSet) setZones() {
	asg := c.group.AutoScaling
	if asg == nil || len(asg.Zones) == 0 {
		return
	}
	c.set.Zones = make([]*string, 0)
	for i := range asg.Zones {
		c.set.Zones = append(c.set.Zones, to.Ptr(asg.Zones[i]))
	}
}

// setSystemDisk 系统盘
func (c *nodeGroupToSet) setSystemDisk() {
	// 系统盘只购买一次！不允许修改
}

// setDataDisks 数据盘
func (c *nodeGroupToSet) setDataDisks() {
	// 数据盘只购买一次！不允许修改
}

// vmToNode vm 转 node
type vmToNode struct {
	node *proto.Node
	vm   *armcompute.VirtualMachineScaleSetVM
}

func newVmToNodeConverter(vm *armcompute.VirtualMachineScaleSetVM, node *proto.Node) *vmToNode {
	return &vmToNode{
		vm:   vm,
		node: node,
	}
}

func (c *vmToNode) convert() {
	/*
		节点私有IP
		Interface.properties.ipConfigurations[n].properties.privateIPAddress（IP地址）
		注：
			“IPv4”和“IPv６”共用当前字段，通过Interface.properties.ipConfigurations[n].properties.privateIPAddressVersion区分地址类型；
			Interface是网卡对象，一个主机可以有多个网卡；
			ipConfigurations是网卡的IP对象，因此，IP只能从网卡中拿到；
	*/
	c.setNodeID()
	c.setNodeName()
	c.setInstanceType()
	c.setStatus()
	c.setZoneID()
	c.setVpcID()
	c.setRegion()
	c.setPassword()
	c.node.NodeType = "CVM"
}

func (c *vmToNode) setNodeID() {
	c.node.NodeID = VmIDToNodeID(c.vm)
}

func (c *vmToNode) setNodeName() {
	//	使用VirtualMachineScaleSetVM.properties.osProfile.computerName
	if c.vm.Properties == nil {
		return
	}
	if c.vm.Properties.OSProfile == nil {
		return
	}
	name := *c.vm.Properties.OSProfile.ComputerName
	if len(name) == 0 {
		return
	}
	c.node.NodeName = name
}

func (c *vmToNode) setInstanceType() {
	// VirtualMachineScaleSetVM.properties.hardwareProfile.vmSize (机型)
	if c.vm.Properties == nil {
		return
	}
	if c.vm.Properties.HardwareProfile == nil {
		return
	}
	vmSize := string(*c.vm.Properties.HardwareProfile.VMSize)
	if len(vmSize) == 0 {
		return
	}
	c.node.InstanceType = vmSize
}

func (c *vmToNode) setStatus() {
	// 实例的状态（running 运行中，initializing 初始化中，failed 异常）
	switch *c.vm.Properties.ProvisioningState {
	case NormalState:
		c.node.Status = "running"
	case CreatingState:
		c.node.Status = "initializing"
	default:
		c.node.Status = strings.ToLower(*c.vm.Properties.ProvisioningState)
	}
}

func (c *vmToNode) setZoneID() {
	if len(c.vm.Zones) >= 1 {
		c.node.ZoneID = *c.vm.Zones[0]
	}
}

func (c *vmToNode) setVpcID() {
	vmSubnetNames := ParseVmReturnSubnetNames(c.vm)
	if len(vmSubnetNames) >= 1 {
		c.node.VPC = vmSubnetNames[0]
	}
}

func (c *vmToNode) setRegion() {
	if len(*c.vm.Location) != 0 {
		c.node.Region = *c.vm.Location
	}
}

// setPassword 设置密码
func (c *vmToNode) setPassword() {
	if c.vm.Properties == nil || c.vm.Properties.OSProfile == nil {
		return
	}
	if c.vm.Properties.OSProfile.LinuxConfiguration != nil &&
		*c.vm.Properties.OSProfile.LinuxConfiguration.DisablePasswordAuthentication {
		return
	}

	// username 从 vm 对象获取
	// password 从 nodeGroup.launchTemplate.initLoginPassword 对象获取
	// if c.vm.Properties.OSProfile.AdminUsername != nil && len(*c.vm.Properties.OSProfile.AdminUsername) != 0 {
	//	c.node.Username = *c.vm.Properties.OSProfile.AdminUsername
	//}
	// if c.vm.Properties.OSProfile.AdminPassword != nil && len(*c.vm.Properties.OSProfile.AdminPassword) != 0 {
	//	c.node.Passwd = *c.vm.Properties.OSProfile.AdminPassword
	// }
}

// nodeToVm node 转 vm
type nodeToVm struct {
	node *proto.Node
	vm   *armcompute.VirtualMachineScaleSetVM
}

func newNodeToVmConverter(node *proto.Node, vm *armcompute.VirtualMachineScaleSetVM) *nodeToVm {
	return &nodeToVm{
		node: node,
		vm:   vm,
	}
}

func (c *nodeToVm) convert() {
	if c.vm.Properties == nil {
		c.vm.Properties = new(armcompute.VirtualMachineScaleSetVMProperties)
	}
	c.setVmSize()
	c.setLocation()
}

func (c *nodeToVm) setVmSize() {
	// VirtualMachineScaleSetVM.properties.hardwareProfile.vmSize (机型)
	if len(c.node.InstanceType) == 0 {
		return
	}
	if c.vm.Properties.HardwareProfile == nil {
		c.vm.Properties.HardwareProfile = new(armcompute.HardwareProfile)
	}
	c.vm.Properties.HardwareProfile.VMSize = to.Ptr(armcompute.VirtualMachineSizeTypes(c.node.InstanceType))
}

func (c *nodeToVm) setLocation() {
	if len(c.node.Region) != 0 {
		c.vm.Location = to.Ptr(c.node.Region)
	}
}

// SetVmSetNetWork 设置虚拟规模集网络
func SetVmSetNetWork(ctx context.Context, client AksService, group *proto.NodeGroup, rg, nrg string,
	set *armcompute.VirtualMachineScaleSet) error {
	vpcID := group.AutoScaling.VpcID
	subnetIDs := group.AutoScaling.SubnetIDs
	if len(vpcID) == 0 || len(subnetIDs) == 0 || len(group.LaunchTemplate.SecurityGroupIDs) == 0 {
		return fmt.Errorf("SetVmSetNetWork vpcID, subnetID or SecurityGroupIDs can not be empty")
	}

	defaultVpcIDs := ParseSetReturnSubnetNames(set)
	defaultSubnets := ParseSetReturnSubnetIDs(set)
	if len(defaultVpcIDs) == 0 || len(defaultSubnets) == 0 {
		return fmt.Errorf("SetVmSetNetWork vpcID or subnetID for scaleset %s is empty", *set.Name)
	}

	subnetDetail, err := client.GetSubnet(ctx, nrg, vpcID, subnetIDs[0])
	if err != nil {
		blog.Errorf("SetVmSetNetWork GetSubnet %s failed, %v", subnetIDs[0], err)
		return err
	}
	set.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations[0].Properties.
		IPConfigurations[0].Properties.Subnet.ID = subnetDetail.ID
	//	仍然使用本集群默认的安全组,
	sg, err := client.GetNetworkSecurityGroups(ctx, rg, group.LaunchTemplate.SecurityGroupIDs[0])
	if err != nil {
		blog.Errorf("SetVmSetNetWork GetNetworkSecurityGroups %s failed, %v",
			group.LaunchTemplate.SecurityGroupIDs[0], err)
		return err
	}
	set.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations[0].Properties.
		NetworkSecurityGroup.ID = sg.ID

	return nil
}

// SetVmSetPasswd 创建节点池时，调用本方法，用于设置虚拟规模集的密码
func SetVmSetPasswd(group *proto.NodeGroup, set *armcompute.VirtualMachineScaleSet) {
	if group == nil || set == nil || group.LaunchTemplate == nil {
		return
	}
	lc := group.LaunchTemplate
	if lc == nil || len(lc.InitLoginPassword) == 0 || len(lc.InitLoginUsername) == 0 {
		return
	}
	if set.Properties == nil {
		set.Properties = new(armcompute.VirtualMachineScaleSetProperties)
	}
	if set.Properties.VirtualMachineProfile == nil {
		set.Properties.VirtualMachineProfile = new(armcompute.VirtualMachineScaleSetVMProfile)
	}
	if set.Properties.VirtualMachineProfile.OSProfile == nil {
		set.Properties.VirtualMachineProfile.OSProfile = new(armcompute.VirtualMachineScaleSetOSProfile)
	}
	set.Properties.VirtualMachineProfile.OSProfile.AdminUsername = to.Ptr(lc.InitLoginUsername)

	pwd, _ := encrypt.Decrypt(nil, lc.GetInitLoginPassword())
	set.Properties.VirtualMachineProfile.OSProfile.AdminPassword = to.Ptr(pwd)
	// 重置配置
	set.Properties.VirtualMachineProfile.OSProfile.LinuxConfiguration = new(armcompute.LinuxConfiguration)
	// 启用密码验证
	set.Properties.VirtualMachineProfile.OSProfile.LinuxConfiguration.DisablePasswordAuthentication = to.Ptr(false)
}

// SetVmSetSSHPublicKey 创建节点池时，调用本方法，用于设置虚拟规模集SSH免密登录公钥
func SetVmSetSSHPublicKey(group *proto.NodeGroup, set *armcompute.VirtualMachineScaleSet) {
	if group == nil || set == nil || group.LaunchTemplate == nil {
		return
	}
	lc := group.LaunchTemplate
	if lc == nil || lc.KeyPair == nil || len(lc.KeyPair.KeyPublic) == 0 {
		return
	}
	if set.Properties == nil {
		set.Properties = new(armcompute.VirtualMachineScaleSetProperties)
	}
	if set.Properties.VirtualMachineProfile == nil {
		set.Properties.VirtualMachineProfile = new(armcompute.VirtualMachineScaleSetVMProfile)
	}
	if set.Properties.VirtualMachineProfile.OSProfile == nil {
		set.Properties.VirtualMachineProfile.OSProfile = new(armcompute.VirtualMachineScaleSetOSProfile)
	}
	osProfile := set.Properties.VirtualMachineProfile.OSProfile

	// adminName := osProfile.AdminUsername
	osProfile.AdminUsername = to.Ptr(lc.InitLoginUsername)

	if osProfile.LinuxConfiguration == nil {
		osProfile.LinuxConfiguration = new(armcompute.LinuxConfiguration)
	}
	if osProfile.LinuxConfiguration.SSH == nil {
		osProfile.LinuxConfiguration.SSH = new(armcompute.SSHConfiguration)
	}

	osProfile.LinuxConfiguration.SSH.PublicKeys = make([]*armcompute.SSHPublicKey, 0)
	// 设置SSH免密登录公钥
	dePublicKey, _ := encrypt.Decrypt(nil, lc.GetKeyPair().GetKeyPublic())
	osProfile.LinuxConfiguration.SSH.PublicKeys = append(osProfile.LinuxConfiguration.SSH.PublicKeys,
		&armcompute.SSHPublicKey{
			KeyData: to.Ptr(dePublicKey),
			Path:    to.Ptr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", lc.InitLoginUsername)),
		})
}

// SetVmSetCustomScript 创建节点池时，调用本方法，用于设置虚拟规模集用户脚本
func SetVmSetCustomScript(group *proto.NodeGroup, set *armcompute.VirtualMachineScaleSet) {

	if set.Properties == nil {
		set.Properties = new(armcompute.VirtualMachineScaleSetProperties)
	}
	if set.Properties.VirtualMachineProfile == nil {
		set.Properties.VirtualMachineProfile = new(armcompute.VirtualMachineScaleSetVMProfile)
	}
	if set.Properties.VirtualMachineProfile.ExtensionProfile == nil {
		set.Properties.VirtualMachineProfile.ExtensionProfile = new(armcompute.VirtualMachineScaleSetExtensionProfile)
	}
	if set.Properties.VirtualMachineProfile.ExtensionProfile.Extensions == nil {
		set.Properties.VirtualMachineProfile.ExtensionProfile.Extensions =
			make([]*armcompute.VirtualMachineScaleSetExtension, 0)
	}

	exist := false
	for _, e := range set.Properties.VirtualMachineProfile.ExtensionProfile.Extensions {
		if *e.Properties.Type == "CustomScript" {
			e.Properties.Settings = map[string]interface{}{
				"script": group.NodeTemplate.PreStartUserScript,
			}
			exist = true
			break
		}
	}
	if !exist {
		set.Properties.VirtualMachineProfile.ExtensionProfile.Extensions =
			append(set.Properties.VirtualMachineProfile.ExtensionProfile.Extensions,
				&armcompute.VirtualMachineScaleSetExtension{
					Name: to.Ptr("vmssCSEBCS"),
					Properties: &armcompute.VirtualMachineScaleSetExtensionProperties{
						AutoUpgradeMinorVersion: to.Ptr(true),
						Publisher:               to.Ptr("Microsoft.Azure.Extensions"),
						Settings: map[string]interface{}{
							"script": group.NodeTemplate.PreStartUserScript,
						},
						Type:               to.Ptr("CustomScript"),
						TypeHandlerVersion: to.Ptr("2.0"),
					},
				})
	}
}

// BuySystemDisk 购买系统盘
func BuySystemDisk(group *proto.NodeGroup, set *armcompute.VirtualMachineScaleSet) {
	lc := group.LaunchTemplate
	if lc == nil || lc.SystemDisk == nil {
		return
	}
	if set.Properties == nil {
		set.Properties = new(armcompute.VirtualMachineScaleSetProperties)
	}
	if set.Properties.VirtualMachineProfile == nil {
		set.Properties.VirtualMachineProfile = new(armcompute.VirtualMachineScaleSetVMProfile)
	}
	if set.Properties.VirtualMachineProfile.StorageProfile == nil {
		set.Properties.VirtualMachineProfile.StorageProfile = new(armcompute.VirtualMachineScaleSetStorageProfile)
	}
	if set.Properties.VirtualMachineProfile.StorageProfile.OSDisk == nil {
		set.Properties.VirtualMachineProfile.StorageProfile.OSDisk = new(armcompute.VirtualMachineScaleSetOSDisk)
	}
	if set.Properties.VirtualMachineProfile.StorageProfile.OSDisk.ManagedDisk == nil {
		set.Properties.VirtualMachineProfile.StorageProfile.OSDisk.ManagedDisk =
			new(armcompute.VirtualMachineScaleSetManagedDiskParameters)
	}
	if len(lc.SystemDisk.DiskSize) != 0 { // size < 10224
		size, _ := strconv.ParseInt(lc.SystemDisk.DiskSize, 10, 32)
		set.Properties.VirtualMachineProfile.StorageProfile.OSDisk.DiskSizeGB = to.Ptr(int32(size))
	}
}

// BuyDataDisk 购买数据盘
func BuyDataDisk(group *proto.NodeGroup, set *armcompute.VirtualMachineScaleSet) {
	lc := group.LaunchTemplate
	if lc == nil || len(lc.DataDisks) == 0 {
		return
	}
	if set.Properties == nil {
		set.Properties = new(armcompute.VirtualMachineScaleSetProperties)
	}
	if set.Properties.VirtualMachineProfile == nil {
		set.Properties.VirtualMachineProfile = new(armcompute.VirtualMachineScaleSetVMProfile)
	}
	if set.Properties.VirtualMachineProfile.StorageProfile == nil {
		set.Properties.VirtualMachineProfile.StorageProfile = new(armcompute.VirtualMachineScaleSetStorageProfile)
	}

	set.Properties.VirtualMachineProfile.StorageProfile.DataDisks = make([]*armcompute.VirtualMachineScaleSetDataDisk, 0)
	disks := &set.Properties.VirtualMachineProfile.StorageProfile.DataDisks

	for i, d := range lc.DataDisks {
		disk := new(armcompute.VirtualMachineScaleSetDataDisk)
		if len(d.DiskType) != 0 && checkDataDiskType(d.DiskType) { // 设置磁盘类型
			disk.ManagedDisk = new(armcompute.VirtualMachineScaleSetManagedDiskParameters)
			disk.ManagedDisk.StorageAccountType = to.Ptr(armcompute.StorageAccountTypes(d.DiskType))
		}
		if d.DiskType == string(armcompute.StorageAccountTypesUltraSSDLRS) { // 超级磁盘
			setUltraSSD(set)
		}
		if len(d.DiskSize) != 0 { // 设置磁盘大小
			x, _ := strconv.ParseInt(d.DiskSize, 10, 32)
			disk.DiskSizeGB = to.Ptr(int32(x))
		}
		disk.Lun = to.Ptr(int32(i))                                       // 启动顺序
		disk.CreateOption = to.Ptr(armcompute.DiskCreateOptionTypesEmpty) // 空盘
		*disks = append(*disks, disk)
	}
}

// checkDataDiskType 检查数据盘类型是否存在
func checkDataDiskType(diskType string) bool {
	dt := armcompute.StorageAccountTypes(diskType)
	for _, t := range armcompute.PossibleStorageAccountTypesValues() {
		if dt == t {
			return true
		}
	}
	return false
}

// setUltraSSD 购买超级磁盘需要满足两个条件: 1.开启了可用区，2.设置启用超级磁盘
func setUltraSSD(set *armcompute.VirtualMachineScaleSet) {
	if len(set.Zones) == 0 {
		return
	}
	if set.Properties.AdditionalCapabilities == nil {
		set.Properties.AdditionalCapabilities = new(armcompute.AdditionalCapabilities)
	}
	set.Properties.AdditionalCapabilities.UltraSSDEnabled = to.Ptr(true)
}

// ParseSetReturnNgResourcesName 解析 VMSSs 资源名称(ParseSetReturnNodeGroupResourcesName)
func ParseSetReturnNgResourcesName(set *armcompute.VirtualMachineScaleSet) string {
	return RegexpSetNodeGroupResourcesName(set)
}

// ParseSetReturnSubnetNames 解析 VMSSs subnet name(私有网络名称)
func ParseSetReturnSubnetNames(set *armcompute.VirtualMachineScaleSet) []string {
	return regexpSetSubnetName(set)
}

// ParseVmReturnSubnetNames 解析 vm subnet name(私有网络名称)
func ParseVmReturnSubnetNames(vm *armcompute.VirtualMachineScaleSetVM) []string {
	return regexpVmSubnetName(vm)
}

// ParseSetReturnSubnetIDs 解析 VMSSs subnet id(子网id)
func ParseSetReturnSubnetIDs(set *armcompute.VirtualMachineScaleSet) []string {
	return matchSubnetID(set)
}

// ParseSecurityGroupsInVpc 解析 安全组(ParseSecurityGroupsInVpc)
func ParseSecurityGroupsInVpc(vnet *armnetwork.VirtualNetwork) []string {
	return regexpVNetSecurityGroups(vnet)
}

// checkInterfaceConfigNameEmpty interfaceConfig name 不为空检查
func checkInterfaceConfigNameEmpty(vm *armcompute.VirtualMachineScaleSetVM) (string, bool) { // nolint
	if vm == nil || vm.InstanceID == nil {
		return "", false
	}
	properties := vm.Properties
	if properties == nil || properties.NetworkProfileConfiguration == nil {
		return "", false
	}
	interfaceConfig := properties.NetworkProfileConfiguration.NetworkInterfaceConfigurations
	if len(interfaceConfig) == 0 {
		return "", false
	}
	return *interfaceConfig[0].Name, true
}

// RegexpSetNodeGroupResourcesName 通过正则解析 VMSSs resources name
func RegexpSetNodeGroupResourcesName(set *armcompute.VirtualMachineScaleSet) string {
	if set == nil || set.Identity == nil {
		return ""
	}
	reg := regexp.MustCompile("/resourceGroups/(.+?)/")
	for key := range set.Identity.UserAssignedIdentities {
		res := reg.FindStringSubmatch(key)
		if len(res) <= 1 {
			continue
		}
		return res[1]
	}
	return ""
}

// regexpVmNodeGroupResourcesName 通过正则解析 vm resources name
func regexpVmNodeGroupResourcesName(vm *armcompute.VirtualMachineScaleSetVM) string {
	if vm == nil || len(*vm.ID) == 0 {
		return ""
	}
	reg := regexp.MustCompile("/resourceGroups/(.+?)/")
	res := reg.FindStringSubmatch(*vm.ID)
	if len(res) <= 1 {

		return ""
	}
	return res[1]
}

// traversalNetworkConfigurations 遍历 armcompute.VirtualMachineScaleSetNetworkConfiguration
func traversalNetworkConfigurations(configs []*armcompute.VirtualMachineScaleSetNetworkConfiguration) []string {
	res := make([]string, 0)
	for _, c := range configs {
		if c.Properties == nil || len(c.Properties.IPConfigurations) == 0 {
			continue
		}
		for _, ipConfig := range c.Properties.IPConfigurations {
			if ipConfig.Properties == nil || ipConfig.Properties.Subnet == nil || ipConfig.Properties.Subnet.ID == nil {
				continue
			}
			res = append(res, *ipConfig.Properties.Subnet.ID)
		}
	}
	return res
}

// checkSetSubnetID VirtualMachineScaleSet subnetID字段不为空检查
func checkSetSubnetID(set *armcompute.VirtualMachineScaleSet) []string {
	if set == nil {
		return nil
	}
	if set.Properties == nil {
		return nil
	}
	vmProfile := set.Properties.VirtualMachineProfile
	if vmProfile == nil {
		return nil
	}
	if vmProfile.NetworkProfile == nil || len(vmProfile.NetworkProfile.NetworkInterfaceConfigurations) == 0 {
		return nil
	}
	return traversalNetworkConfigurations(vmProfile.NetworkProfile.NetworkInterfaceConfigurations)
}

// regexpSetSubnetName 通过正则解析 VMSSs subnet name
func regexpSetSubnetName(set *armcompute.VirtualMachineScaleSet) []string {
	result := make([]string, 0)
	subnetIDs := checkSetSubnetID(set)
	reg := regexp.MustCompile("/virtualNetworks/(.+?)/")
	for _, id := range subnetIDs {
		res := reg.FindStringSubmatch(id)
		if len(res) <= 1 {
			continue
		}
		result = append(result, res[1])
	}
	return result
}

// RegexpSetSubnetResourceGroup 通过正则解析 VMSSs subnet resourceGroup
func RegexpSetSubnetResourceGroup(set *armcompute.VirtualMachineScaleSet) string {
	subnetIDs := checkSetSubnetID(set)
	reg := regexp.MustCompile("/resourceGroups/(.+?)/")
	for _, id := range subnetIDs {
		res := reg.FindStringSubmatch(id)
		if len(res) <= 1 {
			continue
		}
		return res[1]
	}
	return ""
}

// checkVmSubnetID VirtualMachineScaleSetVM subnetID字段不为空检查
func checkVmSubnetID(vm *armcompute.VirtualMachineScaleSetVM) []string {
	if vm == nil {
		return nil
	}
	if vm.Properties == nil {
		return nil
	}
	vmProfile := vm.Properties.NetworkProfileConfiguration
	if vmProfile == nil {
		return nil
	}
	return traversalNetworkConfigurations(vmProfile.NetworkInterfaceConfigurations)
}

func checkVNetSecurityGroups(vnet *armnetwork.VirtualNetwork) []string {
	if vnet == nil {
		return nil
	}
	if vnet.Properties == nil {
		return nil
	}
	resp := make([]string, 0)
	// VirtualNetwork.properties.subnets[n].properties.networkSecurityGroup.id（安全组ID）
	for _, subnet := range vnet.Properties.Subnets {
		if subnet.Properties != nil && subnet.Properties.NetworkSecurityGroup != nil {
			resp = append(resp, *subnet.Properties.NetworkSecurityGroup.ID)
		}
	}
	return resp
}

// regexpVmSubnetName 通过正则解析 vm subnet name
func regexpVmSubnetName(vm *armcompute.VirtualMachineScaleSetVM) []string {
	result := make([]string, 0)
	subnetIDs := checkVmSubnetID(vm)
	reg := regexp.MustCompile("/virtualNetworks/(.+?)/")
	for _, id := range subnetIDs {
		res := reg.FindStringSubmatch(id)
		if len(res) <= 1 {
			continue
		}
		result = append(result, res[1])
	}
	return result
}

func regexpVNetSecurityGroups(vnet *armnetwork.VirtualNetwork) []string {
	result := make([]string, 0)
	subnetIDs := checkVNetSecurityGroups(vnet)
	for _, id := range subnetIDs {
		idx := strings.LastIndexByte(id, '/')
		result = append(result, id[idx+1:])
	}
	return result
}

// matchSubnetID 获得子网id
func matchSubnetID(set *armcompute.VirtualMachineScaleSet) []string {
	result := make([]string, 0)
	subnetIDs := checkSetSubnetID(set)
	for _, id := range subnetIDs {
		idx := strings.LastIndexByte(id, '/')
		result = append(result, id[idx+1:])
	}
	return result
}

// getInterfaceMap vm interface map
// key = Interface ID
// value = IP数组
func getInterfaceMap(interfaceList []*armnetwork.Interface) map[string][]string {
	interfaceMap := make(map[string][]string)
	for _, v := range interfaceList {
		key := *v.ID
		ipList, ok := interfaceMap[key]
		if !ok || ipList == nil { // 不能存在 or list为空
			ipList = make([]string, 0)
		}
		// 存在，并加入到list中
		interfaceMap[key] = append(ipList, handlerNetworkInterface(v)...)
	}
	return interfaceMap
}

// handlerNetworkInterface 处理每个 Interface
func handlerNetworkInterface(netInterface *armnetwork.Interface) []string {
	ipList := make([]string, 0)
	if netInterface == nil || netInterface.Properties == nil {
		return ipList
	}
	for _, config := range netInterface.Properties.IPConfigurations {
		ip, ok := handlerIPConfig(config)
		if !ok || len(ip) == 0 {
			continue
		}
		ipList = append(ipList, ip)
	}
	return ipList
}

// handlerIPConfig 处理每个 Interface 下的 IP Configuration
func handlerIPConfig(config *armnetwork.InterfaceIPConfiguration) (string, bool) {
	if config.Properties == nil || config.Properties.PrivateIPAddress == nil {
		return "", false
	}
	// VNet地址范围支持问题
	// (https://learn.microsoft.com/zh-cn/azure/virtual-network
	// /virtual-networks-faq#what-address-ranges-can-i-use-in-my-vnets)
	ip := net.ParseIP(*config.Properties.PrivateIPAddress)
	if !(ip != nil && ip.IsPrivate()) { // 取反
		return "", false
	}
	return ip.String(), true
}

// vmMatchIP vm 匹配 ip
// 注意：一个vm可能会存在多个网卡，一个网卡可以有多个ip配置
func vmMatchIP(vm *armcompute.VirtualMachineScaleSetVM, interfaceMap map[string][]string) []string {
	ipList := make([]string, 0)
	if vm == nil || vm.Properties == nil || vm.Properties.NetworkProfile == nil || interfaceMap == nil {
		return ipList
	}
	for _, reference := range vm.Properties.NetworkProfile.NetworkInterfaces {
		//  使用 interface id 获取 ip数组
		if list, ok := interfaceMap[*reference.ID]; ok && len(list) != 0 {
			ipList = append(ipList, list...)
		}
	}
	return ipList
}

// VmMatchInterface vm list 匹配 interface list
func VmMatchInterface(
	vmList []*armcompute.VirtualMachineScaleSetVM, interfaceList []*armnetwork.Interface) map[string][]string {
	vmIPMap := make(map[string][]string)
	if len(vmList) == 0 || len(interfaceList) == 0 {
		return vmIPMap
	}

	interfaceMap := getInterfaceMap(interfaceList)
	for _, vm := range vmList {
		vmIPMap[*vm.Name] = vmMatchIP(vm, interfaceMap)
	}
	return vmIPMap
}

// VmIDToNodeID vm id 转换为 node id
func VmIDToNodeID(vm *armcompute.VirtualMachineScaleSetVM) string {
	if vm == nil {
		return ""
	}
	// 使用：“name/instanceId/nodeResourceGroup”作为节点ID
	nodeGroupResource := regexpVmNodeGroupResourcesName(vm)
	return fmt.Sprintf("%s/%s/%s", *vm.Name, *vm.InstanceID, nodeGroupResource)
}

// SetUserData 设置用户数据
func SetUserData(group *proto.NodeGroup, set *armcompute.VirtualMachineScaleSet) {
	if group.GetNodeTemplate() == nil || len(group.GetNodeTemplate().GetPreStartUserScript()) == 0 {
		return
	}
	if set == nil || set.Properties == nil || set.Properties.VirtualMachineProfile == nil {
		return
	}
	set.Properties.VirtualMachineProfile.UserData = to.Ptr(group.GetNodeTemplate().GetPreStartUserScript())
}

// SetImageReferenceNull 镜像引用-暂时置空处理，若不置空会导致无法更新set
func SetImageReferenceNull(set *armcompute.VirtualMachineScaleSet) {
	if set == nil || set.Properties == nil || set.Properties.VirtualMachineProfile == nil {
		return
	}
	profile := set.Properties.VirtualMachineProfile
	if profile.StorageProfile == nil {
		return
	}
	profile.StorageProfile.ImageReference = nil
}

// SetAgentPoolFromNodeGroup 更新pool，只允许设置tag、vm、label、taint
func SetAgentPoolFromNodeGroup(group *proto.NodeGroup, pool *armcontainerservice.AgentPool) {
	if group == nil || pool == nil {
		return
	}
	if pool.Properties == nil {
		pool.Properties = new(armcontainerservice.ManagedClusterAgentPoolProfileProperties)
	}
	pool.Properties.Tags = make(map[string]*string)
	for k := range group.Tags {
		pool.Properties.Tags[k] = to.Ptr(group.Tags[k])
	}
	if group.NodeTemplate != nil {
		pool.Properties.NodeLabels = make(map[string]*string)
		for k := range group.NodeTemplate.Labels {
			pool.Properties.NodeLabels[k] = to.Ptr(group.NodeTemplate.Labels[k])
		}
	}

	taints := group.NodeTemplate.GetTaints()
	if taints == nil || len(taints) == 0 {
		taints = make([]*proto.Taint, 0)
	}

	/*
		// attention: azure not support addNodes to set unScheduled nodes, thus realize this feature by taint
		taints = append(taints, &proto.Taint{
			Key:    cutils.BCSNodeGroupTaintKey,
			Value:  cutils.BCSNodeGroupTaintValue,
			Effect: cutils.BCSNodeGroupAzureTaintEffect,
		})
	*/
	pool.Properties.NodeTaints = make([]*string, 0)
	for _, taint := range taints {
		pool.Properties.NodeTaints = append(pool.Properties.NodeTaints,
			to.Ptr(fmt.Sprintf("%s=%s:%s", taint.Key, taint.Value, taint.Effect)))
	}

	if group.LaunchTemplate != nil && len(group.LaunchTemplate.InstanceType) != 0 &&
		checkInstanceType(group.LaunchTemplate.InstanceType) {
		// 检查机型是否在里面
		pool.Properties.VMSize = to.Ptr(group.LaunchTemplate.InstanceType)
	}
	pool.Properties.EnableAutoScaling = to.Ptr(false)
	// pool.Properties.OSDiskType = original.Properties.OSDiskType
	// pool.Properties.OSDiskSizeGB = original.Properties.OSDiskSizeGB
}

func checkInstanceType(intce string) bool {
	instance := armcontainerservice.ContainerServiceVMSizeTypes(intce)
	for _, vm := range armcontainerservice.PossibleContainerServiceVMSizeTypesValues() {
		if instance == vm {
			return true
		}
	}
	return false
}
