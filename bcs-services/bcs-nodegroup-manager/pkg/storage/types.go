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

package storage

import (
	"time"
)

const (
	// StableState stable, ResourcePool do not scale down
	// unexpected num or resource pool num is stable
	StableState = "Stable"
	// ScaleUpState elastic nodegroup is scaling up
	ScaleUpState = "Scaleup"
	// ScaleDownState elastic nodegroup is scaling down
	ScaleDownState = "Scaledown"
	// ScaleDownByTaskState nodegroup is scaling down by resource manager group
	ScaleDownByTaskState = "ScaleDownByTask"
	// UpdateNodeMeta nodegroup need to update node meta
	UpdateNodeMeta = "UpdateNodeMeta"
	// ErrState error happened
	ErrState = "ErrState"
	// TimeoutState timeout
	TimeoutState = "TimeoutState"
	// InitState information initialization
	InitState = "InitState"
	// NodeInitState resource is initializing for idle state
	NodeInitState = "INIT"
	// NodeIdleState idle state
	NodeIdleState = "IDLE"
	// NodeConsumedState resource is used for specified nodegroup
	NodeConsumedState = "CONSUMED"
	// NodeReturnState resource return to resource pool
	NodeReturnState = "RETURNED"
	// TaskFinishedState task finished
	TaskFinishedState = "FINISHED"
	// TaskRequestingState task preparing
	TaskRequestingState = "REQUESTED"
	// TaskUnknownState  task unknown
	TaskUnknownState = "UNKNOWN"
	// TaskFailedState task failed
	TaskFailedState = "FAILED"
	// ActionFinishedState action finished
	ActionFinishedState = "FINISHED"
	// ActionRunningState action running
	ActionRunningState = "RUNNING"
	// ActionTimeoutState action timeout
	ActionTimeoutState = "TIMEOUT"
	// ActionTerminatedState action terminated
	ActionTerminatedState = "TERMINATED"
)

const (
	// BufferStrategyType buffer type
	BufferStrategyType = "buffer"
	// HierarchicalStrategyType hierarchicalBuffer
	HierarchicalStrategyType = "hierarchicalBuffer"
)

const (
	// NodeDrainDelayLabel node drain delay
	NodeDrainDelayLabel = "node.bkbcs.tencent.com/drain-delay"
	// NodeDeadlineLabel deadline
	NodeDeadlineLabel = "node.bkbcs.tencent.com/deadline"
	// NodeDrainTaskLabel task id
	NodeDrainTaskLabel = "node.bkbcs.tencent.com/drain-task-id"
	// NodeGroupLabel nodegroup
	NodeGroupLabel = "bkbcs.tencent.com/nodegroupid"
)

// NodeGroupMgrStrategy 定义如何管理指定的NodeGroup策略
type NodeGroupMgrStrategy struct {
	Name              string            `json:"name" bson:"name"`
	Labels            map[string]string `json:"labels" bson:"labels"`
	ResourcePool      string            `json:"resourcePool" bson:"resource_pool"`
	ReservedNodeGroup *GroupInfo        `json:"reservedNodeGroup" bson:"reserved_node_group"`
	ElasticNodeGroups []*GroupInfo      `json:"elasticNodeGroups" bson:"elastic_node_groups"`
	Strategy          *Strategy         `json:"strategy" bson:"strategy"`
	Status            *State            `json:"status" bson:"status"`
	IsDeleted         bool              `json:"isDeleted" bson:"is_deleted"`
}

// GroupInfo 定义
type GroupInfo struct {
	NodeGroupID string          `json:"nodeGroupId" bson:"node_group_id"`
	ConsumerID  string          `json:"consumerID" bson:"consumer_id"`
	ClusterID   string          `json:"clusterId" bson:"cluster_id"`
	Weight      int             `json:"weight" bson:"weight"`
	Limit       *NodegroupLimit `json:"limit" bson:"limit"`
}

// NodegroupLimit 节点池/集群 节点数量限制
type NodegroupLimit struct {
	NodegroupLimit    bool  `json:"nodegroupLimit" bson:"nodegroup_limit"`
	NodegroupLimitNum int32 `json:"nodegroupLimitNum" bson:"nodegroup_limit_num"`
	ClusterLimit      bool  `json:"clusterLimit" bson:"cluster_limit"`
	ClusterLimitNum   int32 `json:"clusterLimitNum" bson:"cluster_limit_num"`
}

// Strategy define ResourcePool strategy of management
type Strategy struct {
	// Type strategy type, buffer effective now
	Type string `json:"type" bson:"type"`
	// ScaleUpCoolDown elasticNodegroup扩容冷却时间，单位分钟
	ScaleUpCoolDown int `json:"scaleUpCoolDown" bson:"scale_up_cool_down"`
	// ScaleUpDelay elasticNodeGroup扩容最大周期，单位分钟
	ScaleUpDelay int `json:"scaleUpDelay" bson:"scale_up_delay"`
	// MinScaleUpSize elasticNodeGroup最小扩容数量，用于防止抖动
	MinScaleUpSize int `json:"minScaleUpSize" bson:"min_scale_up_size"`
	// ScaleDownDelay elasticNodeGroup缩容任务最大周期，单位分钟
	ScaleDownDelay int `json:"scaleDownDelay" bson:"scale_down_delay"`
	// MaxIdleDelay resource pool空闲稳定最大周期，单位分钟
	MaxIdleDelay      int    `json:"maxIdleDelay" bson:"max_idle_delay"`
	ReservedTimeRange string `json:"reservedTimeRange" bson:"reserved_time_range"`
	// Buffer策略
	Buffer *BufferStrategy `json:"buffer" bson:"buffer"`
	// ScaleDownBeforeDDL 在ddl指定分钟前执行缩容
	ScaleDownBeforeDDL int `json:"scaleDownBeforeDDL" bson:"scale_down_before_ddl"`
	// TimeMode 定时模式
	TimeMode *BufferTimeMode `json:"timeMode" bson:"time_mode"`
	// 多个nodegroup分别设置buffer
	NodegroupBuffer map[string]*NodegroupBuffer `json:"nodegroupBuffer" bson:"nodegroup_buffer"`
}

// BufferStrategy 空闲资源水位策略
// It's hard to control resources to specified number in resource pool.
// So when resources are between Low and High, controller considers that pool is stable.
type BufferStrategy struct {
	// Low低水位，空闲资源比例小于该水位时，elasticNodeGroup必须缩容补充资源池
	Low int `json:"low" bson:"low"`
	// High高水位，空闲资源比例大于该水位时，elasticNodeGroup可以扩容消耗资源池资源
	High int `json:"high" bson:"high"`
}

// BufferTimeMode 时间模式配置
type BufferTimeMode struct {
	ScaleDownWhenTimeout bool          `json:"scaleDownWhenTimeout" bson:"scale_down_when_timeout"`
	TimePeriods          []*TimePeriod `json:"timePeriods" bson:"time_periods"`
	ReservedHours        int           `json:"reservedHours" bson:"reserved_hours"`
}

// NodegroupBuffer 单nodegroup buffer设置
type NodegroupBuffer struct {
	Percent int32 `json:"percent" bson:"percent"`
	Count   int32 `json:"count" bson:"count"`
}

// TimePeriod 扩缩容时间周期
type TimePeriod struct {
	ScaleOutCron string `json:"scaleOutCron" bson:"scale_out_cron"`
	ScaleInCron  string `json:"scaleInCron" bson:"scale_in_cron"`
	ScaleOutTime string `json:"scaleOutTime" bson:"scale_out_time"`
	ScaleInTime  string `json:"scaleInTime" bson:"scale_in_time"`
}

// State strategy status
type State struct {
	// Status current state
	Status     string `json:"status" bson:"status"`
	LastStatus string `json:"lastStatus" bson:"last_status"`
	// Error message if strategy got any failure
	Error string `json:"error" bson:"error"`
	// Message状态提示信息，非错误状态下主要用于信息跟踪
	Message string `json:"message" bson:"message"`
	// CreatedTime策略创建时间
	CreatedTime time.Time `json:"createdTime" bson:"created_time"`
	// UpdatedTime策略更新时间，主要用于跟进操作时间点
	UpdatedTime time.Time `json:"updatedTime" bson:"updated_time"`
	// LastScaleUpTime上次elasticNodeGroup扩容时间，用于恒定状态追踪
	LastScaleUpTime time.Time `json:"lastScaleUpTime" bson:"last_scale_up_time"`
	// LastScaleDownTime上次elasticNodeGroup缩容时间，用于状态追踪
	LastScaleDownTime time.Time `json:"lastScaleDownTime" bson:"last_scale_down_time"`
	// PoolPrevious & PoolExpected 用于进一步资源追踪
	PoolPrevious *PoolOverview `json:"poolPrevious" bson:"pool_previous"`
	PoolExpected *PoolOverview `json:"poolExpected" bson:"pool_expected"`
}

// NodeGroup 定义伸缩的具体信息与关联的状态
type NodeGroup struct {
	// NodeGroup unique id
	NodeGroupID string `json:"nodeGroupID" bson:"node_group_id"`
	// ClusterID nodegroup所属集群
	ClusterID string `json:"clusterID" bson:"cluster_id"`
	// MaxSize is the upper limit of the node group
	MaxSize int `json:"maxSize" bson:"max_size"`
	// MinSize is the lower limit of the node group
	MinSize int `json:"minSize" bson:"min_size"`
	// CmDesiredSize is the desire size of node group from cluster manager
	CmDesiredSize int `json:"cmDesiredSize" bson:"cm_desired_size"`
	// DesiredSize is the current size of the node group.
	DesiredSize int `json:"desiredSize" bson:"desired_size"`
	// UpcomingSize is the number that indicates how many nodes have not registered in
	// Kubernetes or have not been ready to be used.
	UpcomingSize int `json:"upcomingSize" bson:"upcoming_size"`
	// NodeIPs are the IP of nodes which belongs to the node group
	NodeIPs []string `json:"nodeIPs" bson:"node_ips"`
	// Status of nodegroup
	Status string `json:"status" bson:"status"`
	// LastStatus用于问题排查
	LastStatus string `json:"lastStatus" bson:"last_status"`
	// HookConfirm用于标注NodeGroup新状态被cluster-autoscaler提取
	HookConfirm bool `json:"hookConfirm" bson:"hook_confirm"`
	// Message running state information
	Message string `json:"message" bson:"message"`
	// UpdatedTime信息更新时间，用于跟进webhook的状态，controller不能更新该字段
	UpdatedTime time.Time `json:"updatedTime" bson:"updated_time"`
	// IsDeleted soft delete
	IsDeleted bool `json:"isDeleted" bson:"is_deleted"`
	// LastScaleUpTime上次扩容时间，用于恒定状态追踪
	LastScaleUpTime time.Time `json:"lastScaleUpTime" bson:"last_scale_up_time"`
	// LastScaleDownTime上次缩容时间，用于状态追踪
	LastScaleDownTime time.Time `json:"lastScaleDownTime" bson:"last_scale_down_time"`
}

// PoolOverview simple resource information
type PoolOverview struct {
	// InitNum处于init状态的resource数量
	InitNum int `json:"initNum" bson:"init_num"`
	// IdleNum处于空闲状态的数量
	IdleNum int `json:"idleNum" bson:"idle_num"`
	// ConsumedNum被NodeGroup消费掉的数量
	ConsumedNum int `json:"consumedNum" bson:"consumed_num"`
	// ReturnedNum在退还中的数量，状态变化会比较快
	ReturnedNum int `json:"returnedNum" bson:"returned_num"`
	// TotalNum 资源总量，用于预防可能出现的资源下架操作
	TotalNum int
}

// NodeGroupAction action for nodegroup scaleup or scaledown
type NodeGroupAction struct {
	NodeGroupID string    `json:"nodeGroupId" bson:"node_group_id"`
	ClusterID   string    `json:"clusterId" bson:"cluster_id"`
	TaskID      string    `json:"taskID" bson:"task_id"`
	CreatedTime time.Time `json:"createdTime" bson:"created_time"`
	// Event scaleUp or scaleDown
	Event string `json:"event" bson:"event"`
	// DeltaNum node number that will scaleUp or scaleDown in nodegroup
	DeltaNum int `json:"deltaNum" bson:"delta_num"`
	// NewDesiredNum new desiredSize for NodeGroup
	NewDesiredNum int `json:"newDesiredNum" bson:"new_desired_num"`
	// OriginalNum last nodegroup desiredNum
	OriginalDesiredNum int `json:"originalDesiredNum" bson:"original_desired_num"`
	// OriginalNodeNum original real node number
	OriginalNodeNum int `json:"originalNodeNum" bson:"original_node_num"`
	// NodeIPs comes from NodeGroup
	NodeIPs []string `json:"nodeIps" bson:"node_ips"`
	// Process simple process
	Process int `json:"process" bson:"process"`
	// Status操作状态是否正常
	Status string `json:"status" bson:"status"`
	// UpdatedTime update time when processing or nodegroup changed
	UpdatedTime time.Time `json:"updatedTime" bson:"updated_time"`
	// IsDeleted 软删除
	IsDeleted bool `json:"isDeleted" bson:"is_deleted"`
	// Strategy 关联的strategy
	Strategy string `json:"strategy" bson:"strategy"`
}

// IsTerminated check nodegroup action progress is in final state
func (action *NodeGroupAction) IsTerminated() bool {
	return action.Process == 100
}

// IsTimeout check nodegroup action progress is timeout
func (action *NodeGroupAction) IsTimeout(delay int) bool {
	gap := time.Since(action.UpdatedTime)
	return gap.Seconds() >= float64(delay*60)
}

// NodeGroupEvent all event tracing for nodegroup
type NodeGroupEvent struct {
	NodeGroupID string    `json:"nodeGroupId" bson:"node_group_id"`
	ClusterID   string    `json:"clusterId" bson:"cluster_id"`
	EventTime   time.Time `json:"eventTime" bson:"event_time"`
	Event       string    `json:"event" bson:"event"`
	MaxNum      int       `json:"maxNum" bson:"max_num"`
	MinNum      int       `json:"minNum" bson:"min_num"`
	// DesiredNum new desired number for nodegroup
	DesiredNum int `json:"desiredNum" bson:"desired_num"`
	// OriginalDesiredNum last desired number
	OriginalDesiredNum int `json:"originalNum" bson:"original_num"`
	// OriginalNodeNum node number when event happened
	OriginalNodeNum int `json:"originalNodeNum" bson:"original_node_num"`
	// event trigger reason
	Reason string `json:"reason" bson:"reason"`
	// detail message that for debug
	Message   string `json:"message" bson:"message"`
	IsDeleted bool   `json:"isDeleted" bson:"is_deleted"`
}

// Resource 节点资源定义，为不依赖resource-manager进行重定义
type Resource struct {
	// ID信息
	ID string `json:"id" bson:"id"`
	// InnerIP ipv4地址
	InnerIP string `json:"innerIp" bson:"inner_ip"`
	// InnerIPv6 ipv6地址
	InnerIPv6 string `json:"inneriPv6" bson:"inneri_pv6"`
	// ResourceType 物理机/CVM，暂时保留
	ResourceType string `json:"resourceType" bson:"resource_type"`
	// ResourceProvider 资源供给者，暂时保留
	ResourceProvider string `json:"resourceProvider" bson:"resource_provider"`
	// Labels资源标签，用于标注VPC相关等信息
	Labels map[string]string `json:"labels" bson:"labels"`
	// UpdatedTime 状态变化更新时间
	UpdatedTime time.Time `json:"updatedTime" bson:"updated_time"`
	// Phase 参照NodeXXXXState
	Phase string `json:"phase" bson:"phase"`
	// DevicePool关联的设备池
	DevicePool string `json:"devicePool" bson:"device_pool"`
	// Cluster关联的集群ID
	Cluster string `json:"cluster" bson:"cluster"`
}

// ClusterNode 划分到集群中统计节点信息
type ClusterNode struct {
	ClusterID string      `json:"clusterId" bson:"cluster_id"`
	Used      int32       `json:"used" bson:"used"`
	Resources []*Resource `json:"resources" bson:"resources"`
}

// ResourcePool 资源池定义，用于本地状态确认
type ResourcePool struct {
	// ID信息
	ID string `json:"id" bson:"id"`
	// Name池子名称，仅存储用途
	Name string `json:"name" bson:"name"`
	// CreatedTime 池子构建时间，仅用于本地缓存
	CreatedTime time.Time `json:"createdTime" bson:"created_time"`
	// UpdatedTime 资源池状态变化更新时间
	UpdatedTime time.Time `json:"updateTime" bson:"update_time"`
	// InitNum处于init状态的resource数量
	InitNum int `json:"initNum" bson:"init_num"`
	// IdleNum处于空闲状态的数量
	IdleNum int `json:"idleNum" bson:"idle_num"`
	// ConsumedNum被NodeGroup消费掉的数量
	ConsumedNum int `json:"consumedNum" bson:"consumed_num"`
	// ReturnedNum在退还中的数量，状态变化会比较快
	ReturnedNum int `json:"returnedNum" bson:"returned_num"`
	// ClusterNodes记录节点的集群归属
	ClusterNodes []*ClusterNode `json:"clusterNodes" bson:"cluster_nodes"`
	// Resources池子下详细资源信息
	Resources []*Resource `json:"resources" bson:"resources"`
}

// ScaleDownTask 记录从 resource manager获取到的任务信息，并记录筛选出的nodegroup及具体缩容ip
type ScaleDownTask struct {
	TaskID            string             `json:"taskID" bson:"task_id"`
	TotalNum          int                `json:"totalNum" bson:"total_num"`
	NodeGroupStrategy string             `json:"nodeGroupStrategy" bson:"node_group_strategy"`
	DevicePoolID      string             `json:"devicePoolID" bson:"device_pool_id"`
	ScaleDownGroups   []*ScaleDownDetail `json:"scaleDownGroups" bson:"scale_down_groups"`
	DrainDelay        string             `json:"drainDelay" bson:"drain_delay"`
	Deadline          time.Time          `json:"deadline" bson:"deadline"`
	// CreatedTime task创建时间
	CreatedTime time.Time `json:"createdTime" bson:"created_time"`
	// UpdatedTime 任务更新时间
	UpdatedTime      time.Time `json:"updateTime" bson:"update_time"`
	BeginExecuteTime time.Time `json:"beginExecuteTime" bson:"begin_execute_time"`
	IsDeleted        bool      `json:"isDeleted" bson:"is_deleted"`
	IsExecuted       bool      `json:"isExecuted" bson:"is_executed"`
	Status           string    `json:"status" bson:"status"`
	DeviceList       []string  `json:"deviceList" bson:"device_list"`
	SpecifyScaleDown bool      `json:"specifyScaleDown" bson:"specify_scale_down"`
	AllocatedNum     int       `json:"allocatedNum" bson:"allocated_num"`
}

// IsTerminated check task status
func (t *ScaleDownTask) IsTerminated() bool {
	return t.Status != TaskRequestingState
}

// IsExecuting check task status
func (t *ScaleDownTask) IsExecuting() bool {
	if t.IsExecuted && !t.IsDeleted {
		return true
	}
	return false
}

// ScaleDownDetail 记录根据 ScaleDownTask 筛选出的节点ip
type ScaleDownDetail struct {
	ConsumerID  string   `json:"consumerID" bson:"consumer_id"`
	NodeGroupID string   `json:"nodeGroupID" bson:"node_group_id"`
	ClusterID   string   `json:"clusterID" bson:"cluster_id"`
	NodeIPs     []string `json:"nodeIPs" bson:"node_ips"`
	NodeNum     int      `json:"nodeNum" bson:"node_num"`
}

// DeviceGroup 根据consumer id查询得到的资源组
type DeviceGroup struct {
	// ConsumerID nodegroup和resource manager交互的consumer id
	ConsumerID string `json:"consumerID"`
	// InitNum处于init状态的resource数量
	InitNum int `json:"initNum"`
	// IdleNum处于空闲状态的数量
	IdleNum int `json:"idleNum"`
	// ConsumedNum被NodeGroup消费掉的数量
	ConsumedNum int `json:"consumedNum"`
	// ReturnedNum在退还中的数量，状态变化会比较快
	ReturnedNum int `json:"returnedNum"`
	// Resources池子下详细资源信息
	Resources []*Resource `json:"resources" bson:"resources"`
	// UpdatedTime 资源池状态变化更新时间
	UpdatedTime time.Time `json:"updateTime" bson:"update_time"`
}

// ScaleDownNodegroup 可以用于下架备份的nodegroup信息
type ScaleDownNodegroup struct {
	DrainDelayHour int
	Total          int
	GroupInfos     []*GroupInfo
	NodeGroups     map[string]*NodeGroup
}
