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

// Package types xxx
package types

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/cmanager"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

const (
	// DataJobQueue queue name
	DataJobQueue = "dataJob"
)

// extract type
const (
	ProjectType            = "project"
	ClusterType            = "cluster"
	NamespaceType          = "namespace"
	WorkloadType           = "workload"
	PublicType             = "public"
	HPAType                = "HorizontalPodAutoscaler"
	GPAType                = "GeneralPodAutoscaler"
	PodAutoscalerType      = "PodAutoscaler"
	GetWorkloadRequestType = "getWorkloadRequest"
)

// extract dimension
const (
	// DimensionMinute  minute dimension for job
	DimensionMinute = "minute"
	// DimensionHour hour dimension for job
	DimensionHour = "hour"
	// DimensionDay  day dimension for job
	DimensionDay = "day"
)

// db table name
const (
	// DataTableNamePrefix db table prefix
	DataTableNamePrefix      = "bcsdatamanager_"
	ClusterTableName         = "cluster"
	ProjectTableName         = "project"
	NamespaceTableName       = "namespace"
	WorkloadTableName        = "workload"
	PublicTableName          = "public"
	PodAutoscalerTableName   = "podAutoscaler"
	WorkloadRequestTableName = "request_workload"
	PredictTableNamePrefix   = "bcs_predict_"
	WorkloadInfoTableName    = "workload_info"
)

// cluster type
const (
	Kubernetes = "k8s"
	Mesos      = "mesos"
)

// workload type
const (
	DeploymentType       = "Deployment"
	StatefulSetType      = "StatefulSet"
	DaemonSetType        = "DaemonSet"
	GameDeploymentType   = "GameDeployment"
	GameStatefulSetType  = "GameStatefulSet"
	MesosApplicationType = "application"
	MesosDeployment      = "deployment"
)

// service
const (
	// ServiceDomain domain name for service
	ServiceDomain = "datamanager.bkbcs.tencent.com"
	// MicroMetaKeyHTTPPort http port in micro service meta
	MicroMetaKeyHTTPPort = "httpport"
)

// time format
const (
	// MonthTimeFormat month bucket time format
	MonthTimeFormat = "2006-01"
	// DayTimeFormat day bucket time format
	DayTimeFormat = "2006-01-02"
	// HourTimeFormat hour bucket time format
	HourTimeFormat = "2006-01-02 15:00:00"
	// MinuteTimeFormat minute bucket time format
	MinuteTimeFormat = "2006-01-02 15:04:00"
	// SecondTimeFormat second bucket time format
	SecondTimeFormat = "2006-01-02 15:04:05"
)

// ProjectMeta meta for project
type ProjectMeta struct {
	ProjectID   string            `json:"projectID"`
	ProjectCode string            `json:"projectCode"`
	BusinessID  string            `json:"businessID"`
	Label       map[string]string `json:"label"`
}

// ClusterMeta meta for cluster
type ClusterMeta struct {
	ProjectID   string            `json:"projectID"`
	ProjectCode string            `json:"projectCode"`
	BusinessID  string            `json:"businessID"`
	ClusterID   string            `json:"clusterID"`
	ClusterType string            `json:"clusterType"`
	Label       map[string]string `json:"label"`
	IsBKMonitor bool              `json:"isBKMonitor"`
}

// NamespaceMeta meta for namespace
type NamespaceMeta struct {
	ProjectID   string            `json:"projectID"`
	ProjectCode string            `json:"projectCode"`
	BusinessID  string            `json:"businessID"`
	ClusterID   string            `json:"clusterID"`
	ClusterType string            `json:"clusterType"`
	Name        string            `json:"name"`
	Label       map[string]string `json:"label"`
	IsBKMonitor bool              `json:"isBKMonitor"`
}

// WorkloadMeta meta for workload
type WorkloadMeta struct {
	ProjectID    string            `json:"projectID"`
	ProjectCode  string            `json:"projectCode"`
	ClusterID    string            `json:"clusterID"`
	BusinessID   string            `json:"businessID"`
	ClusterType  string            `json:"clusterType"`
	Namespace    string            `json:"namespace"`
	ResourceType string            `json:"resourceType"`
	Name         string            `json:"name"`
	Label        map[string]string `json:"label"`
	IsBKMonitor  bool              `json:"isBKMonitor"`
}

// PodAutoscalerMeta meta for hpa or gpa
type PodAutoscalerMeta struct {
	ProjectID          string            `json:"projectID"`
	ProjectCode        string            `json:"projectCode"`
	ClusterID          string            `json:"clusterID"`
	BusinessID         string            `json:"businessID"`
	ClusterType        string            `json:"clusterType"`
	Namespace          string            `json:"namespace"`
	TargetResourceType string            `json:"targetResourceType"`
	TargetWorkloadName string            `json:"targetWorkloadName"`
	PodAutoscaler      string            `json:"podAutoscaler"`
	Label              map[string]string `json:"label"`
	IsBKMonitor        bool              `json:"isBKMonitor"`
}

// PublicData for public table
type PublicData struct {
	CreateTime   primitive.DateTime `json:"create_time" bson:"create_time"`
	UpdateTime   primitive.DateTime `json:"update_time" bson:"update_time"`
	ObjectType   string             `json:"object_type" bson:"object_type"`
	ProjectID    string             `json:"projectID" bson:"project_id"`
	ProjectCode  string             `json:"projectCode" bson:"project_code"`
	BusinessID   string             `json:"businessID" bson:"business_id"`
	ClusterID    string             `json:"clusterID" bson:"cluster_id"`
	ClusterType  string             `json:"clusterType" bson:"cluster_type"`
	Namespace    string             `json:"namespace" bson:"namespace"`
	WorkloadType string             `json:"workloadType" bson:"workload_type"`
	WorkloadName string             `json:"workloadName" bson:"workload_name"`
	Metrics      interface{}        `json:"metrics" bson:"metrics"`
}

// WorkloadData for workload table
// metrics contains detail of every minute/hour/day
// ExtremumRecord records the max/min value
type WorkloadData struct {
	CreateTime         primitive.DateTime             `json:"createTime" bson:"create_time"`
	UpdateTime         primitive.DateTime             `json:"updateTime" bson:"update_time"`
	BucketTime         string                         `json:"bucketTime" bson:"bucket_time"`
	Dimension          string                         `json:"dimension" bson:"dimension"`
	ProjectID          string                         `json:"projectID" bson:"project_id"`
	ProjectCode        string                         `json:"projectCode" bson:"project_code"`
	BusinessID         string                         `json:"businessID" bson:"business_id"`
	ClusterID          string                         `json:"clusterID" bson:"cluster_id"`
	ClusterType        string                         `json:"clusterType" bson:"cluster_type"`
	Namespace          string                         `json:"namespace" bson:"namespace"`
	WorkloadType       string                         `json:"workloadType" bson:"workload_type"`
	Name               string                         `json:"workloadName" bson:"workload_name"`
	Label              map[string]string              `json:"label"  bson:"label"`
	MaxCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"maxCPUUsageTime"  bson:"max_cpu_usage_time"`
	MinCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"minCPUUsageTime" bson:"min_cpu_usage_time"`
	MaxMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"maxMemoryUsageTime" bson:"max_memory_usage_time"`
	MinMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"minMemoryUsageTime" bson:"min_memory_usage_time"`
	MaxCPUTime         *bcsdatamanager.ExtremumRecord `json:"maxCPUTime"  bson:"max_cpu_time"`
	MinCPUTime         *bcsdatamanager.ExtremumRecord `json:"minCPUTime" bson:"min_cpu_time"`
	MaxMemoryTime      *bcsdatamanager.ExtremumRecord `json:"maxMemoryTime" bson:"max_memory_time"`
	MinMemoryTime      *bcsdatamanager.ExtremumRecord `json:"minMemoryTime" bson:"min_memory_time"`
	MinInstanceTime    *bcsdatamanager.ExtremumRecord `json:"minInstanceTime" bson:"min_instance_time"`
	MaxInstanceTime    *bcsdatamanager.ExtremumRecord `json:"maxInstanceTime" bson:"max_instance_time"`
	Metrics            []*WorkloadMetrics             `json:"metrics" bson:"metrics"`
}

// PodAutoscalerData for podAutoscaler table
// metrics contains detail of every minute/hour/day
// ExtremumRecord records the max/min value
type PodAutoscalerData struct {
	CreateTime        primitive.DateTime      `json:"createTime" bson:"create_time"`
	UpdateTime        primitive.DateTime      `json:"updateTime" bson:"update_time"`
	BucketTime        string                  `json:"bucketTime" bson:"bucket_time"`
	Dimension         string                  `json:"dimension" bson:"dimension"`
	ProjectID         string                  `json:"projectID" bson:"project_id"`
	ProjectCode       string                  `json:"projectCode" bson:"project_code"`
	BusinessID        string                  `json:"businessID" bson:"business_id"`
	ClusterID         string                  `json:"clusterID" bson:"cluster_id"`
	ClusterType       string                  `json:"clusterType" bson:"cluster_type"`
	Namespace         string                  `json:"namespace" bson:"namespace"`
	WorkloadType      string                  `json:"workloadType" bson:"workload_type"`
	WorkloadName      string                  `json:"workloadName" bson:"workload_name"`
	PodAutoscalerType string                  `json:"podAutoscalerType" bson:"pod_autoscaler_type"`
	PodAutoscalerName string                  `json:"podAutoscalerName" bson:"pod_autoscaler_name"`
	Total             int64                   `json:"total" bson:"total"`
	Label             map[string]string       `json:"label"  bson:"label"`
	Metrics           []*PodAutoscalerMetrics `json:"metrics" bson:"metrics"`
}

// ProjectData for project table
// metrics contains detail of every minute/hour/day
// ExtremumRecord records the max/min value
type ProjectData struct {
	CreateTime  primitive.DateTime             `json:"createTime" bson:"create_time"`
	UpdateTime  primitive.DateTime             `json:"updateTime" bson:"update_time"`
	BucketTime  string                         `json:"bucketTime" bson:"bucket_time"`
	Dimension   string                         `json:"dimension" bson:"dimension"`
	ProjectID   string                         `json:"projectID" bson:"project_id"`
	ProjectCode string                         `json:"projectCode" bson:"project_code"`
	BusinessID  string                         `json:"businessID" bson:"business_id"`
	Label       map[string]string              `json:"label"  bson:"label"`
	MinNode     *bcsdatamanager.ExtremumRecord `json:"minNode,omitempty" bson:"min_node"`
	MaxNode     *bcsdatamanager.ExtremumRecord `json:"maxNode,omitempty" bson:"max_node"`
	Metrics     []*ProjectMetrics              `json:"metrics" bson:"metrics"`
}

// NamespaceData for namespace table
// metrics contains detail of every minute/hour/day
// ExtremumRecord records the max/min value
type NamespaceData struct {
	CreateTime         primitive.DateTime             `json:"createTime" bson:"create_time"`
	UpdateTime         primitive.DateTime             `json:"updateTime" bson:"update_time"`
	BucketTime         string                         `json:"bucketTime" bson:"bucket_time"`
	Dimension          string                         `json:"dimension" bson:"dimension"`
	ProjectID          string                         `json:"projectID" bson:"project_id"`
	ProjectCode        string                         `json:"projectCode" bson:"project_code"`
	BusinessID         string                         `json:"businessID" bson:"business_id"`
	ClusterID          string                         `json:"clusterID" bson:"cluster_id"`
	ClusterType        string                         `json:"clusterType" bson:"cluster_type"`
	Namespace          string                         `json:"namespace" bson:"namespace"`
	Label              map[string]string              `json:"label"  bson:"label"`
	MaxCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"maxCPUUsageTime" bson:"max_cpu_usage_time"`
	MinCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"minCPUUsageTime" bson:"min_cpu_usage_time"`
	MaxMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"maxMemoryUsageTime" bson:"max_memory_usage_time"`
	MinMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"minMemoryUsageTime" bson:"min_memory_usage_time"`
	MinInstanceTime    *bcsdatamanager.ExtremumRecord `json:"minInstanceTime" bson:"min_instance_time"`
	MaxInstanceTime    *bcsdatamanager.ExtremumRecord `json:"maxInstanceTime" bson:"max_instance_time"`
	MinWorkloadUsage   *bcsdatamanager.ExtremumRecord `json:"minWorkloadUsage" bson:"min_workload_usage"`
	MaxWorkloadUsage   *bcsdatamanager.ExtremumRecord `json:"maxWorkloadUsage" bson:"max_workload_usage"`
	Metrics            []*NamespaceMetrics            `json:"metrics" bson:"metrics"`
}

// ClusterData for cluster table
// metrics contains detail of every minute/hour/day
// ExtremumRecord records the max/min value
type ClusterData struct {
	CreateTime   primitive.DateTime             `json:"createTime" bson:"create_time"`
	UpdateTime   primitive.DateTime             `json:"updateTime" bson:"update_time"`
	BucketTime   string                         `json:"bucketTime" bson:"bucket_time"`
	Dimension    string                         `json:"dimension" bson:"dimension"`
	ProjectID    string                         `json:"projectID" bson:"project_id"`
	ProjectCode  string                         `json:"projectCode" bson:"project_code"`
	BusinessID   string                         `json:"businessID" bson:"business_id"`
	ClusterID    string                         `json:"clusterID" bson:"cluster_id"`
	ClusterType  string                         `json:"clusterType" bson:"cluster_type"`
	TotalCACount int64                          `json:"totalCACount" bson:"total_ca_count"`
	Label        map[string]string              `json:"label"  bson:"label"`
	MinNode      *bcsdatamanager.ExtremumRecord `json:"minNode,omitempty" bson:"min_node"`
	MaxNode      *bcsdatamanager.ExtremumRecord `json:"maxNode,omitempty" bson:"max_node"`
	MinInstance  *bcsdatamanager.ExtremumRecord `json:"minInstance,omitempty" bson:"min_instance"`
	MaxInstance  *bcsdatamanager.ExtremumRecord `json:"maxInstance,omitempty" bson:"max_instance"`
	MaxCPU       *bcsdatamanager.ExtremumRecord `json:"maxCPU" bson:"max_cpu"`
	MinCPU       *bcsdatamanager.ExtremumRecord `json:"minCPU" bson:"min_cpu"`
	MaxMemory    *bcsdatamanager.ExtremumRecord `json:"maxMemory" bson:"max_memory"`
	MinMemory    *bcsdatamanager.ExtremumRecord `json:"minMemory" bson:"min_memory"`
	Metrics      []*ClusterMetrics              `json:"metrics" bson:"metrics"`
}

// ProjectPublicMetrics public for project
type ProjectPublicMetrics struct {
	ClusterCount int64 `json:"cluster_count"`
}

// ClusterPublicMetrics public for cluster
type ClusterPublicMetrics struct {
}

// NamespacePublicMetrics public for namespace
type NamespacePublicMetrics struct {
	ResourceLimit *bcsdatamanager.ResourceLimit `json:"resource_limit"`
	SuggestCPU    float64                       `json:"suggest_cpu"`
	SuggestMemory float64                       `json:"suggest_memory"`
}

// WorkloadPublicMetrics public for workload
type WorkloadPublicMetrics struct {
	SuggestCPU    float64 `json:"suggest_cpu"`
	SuggestMemory float64 `json:"suggest_memory"`
}

// ProjectMetrics project metric
type ProjectMetrics struct {
	Index              int                            `json:"index"`
	Time               primitive.DateTime             `json:"time,omitempty"`
	ClustersCount      int64                          `json:"clustersCount,omitempty"`
	TotalCPU           float64                        `json:"totalCPU,omitempty"`
	TotalMemory        int64                          `json:"totalMemory,omitempty"`
	TotalLoadCPU       float64                        `json:"totalLoadCPU,omitempty"`
	TotalLoadMemory    int64                          `json:"totalLoadMemory,omitempty"`
	AvgLoadCPU         float64                        `json:"avgLoadCPU,omitempty"`
	AvgLoadMemory      int64                          `json:"avgLoadMemory,omitempty"`
	CPUUsage           float64                        `json:"CPUUsage,omitempty"`
	MemoryUsage        float64                        `json:"MemoryUsage,omitempty"`
	NodeCount          int64                          `json:"nodeCount,omitempty"`
	AvailableNodeCount int64                          `json:"availableNodeCount,omitempty"`
	MinNode            *bcsdatamanager.ExtremumRecord `json:"minNode,omitempty"`
	MaxNode            *bcsdatamanager.ExtremumRecord `json:"maxNode,omitempty"`
}

// ClusterMetrics cluster metric
type ClusterMetrics struct {
	Index              int                            `json:"index"`
	Time               primitive.DateTime             `json:"time,omitempty"`
	NodeCount          int64                          `json:"nodeCount,omitempty"`
	AvailableNodeCount int64                          `json:"availableNodeCount,omitempty"`
	MinNode            *bcsdatamanager.ExtremumRecord `json:"minNode,omitempty"`
	MaxNode            *bcsdatamanager.ExtremumRecord `json:"maxNode,omitempty"`
	NodeQuantile       []*bcsdatamanager.NodeQuantile `json:"nodeQuantile,omitempty"`
	MinUsageNode       string                         `json:"minUsageNode,omitempty"`
	TotalCPU           float64                        `json:"totalCPU,omitempty"`
	TotalMemory        int64                          `json:"totalMemory,omitempty"`
	TotalLoadCPU       float64                        `json:"totalLoadCPU,omitempty"`
	TotalLoadMemory    int64                          `json:"totalLoadMemory,omitempty"`
	AvgLoadCPU         float64                        `json:"avgLoadCPU,omitempty"`
	AvgLoadMemory      int64                          `json:"avgLoadMemory,omitempty"`
	CPUUsage           float64                        `json:"CPUUsage,omitempty"`
	MemoryUsage        float64                        `json:"MemoryUsage,omitempty"`
	WorkloadCount      int64                          `json:"workloadCount,omitempty"`
	InstanceCount      int64                          `json:"instanceCount,omitempty"`
	MinInstance        *bcsdatamanager.ExtremumRecord `json:"minInstance,omitempty"`
	MaxInstance        *bcsdatamanager.ExtremumRecord `json:"maxInstance,omitempty"`
	MinCPU             *bcsdatamanager.ExtremumRecord `json:"minCPU,omitempty"`
	MaxCPU             *bcsdatamanager.ExtremumRecord `json:"maxCPU,omitempty"`
	MinMemory          *bcsdatamanager.ExtremumRecord `json:"minMemory,omitempty"`
	MaxMemory          *bcsdatamanager.ExtremumRecord `json:"maxMemory,omitempty"`
	CpuRequest         float64                        `json:"cpuRequest,omitempty"`
	CPULimit           float64                        `json:"CPULimit,omitempty"`
	MemoryRequest      int64                          `json:"memoryRequest,omitempty"`
	MemoryLimit        int64                          `json:"memoryLimit,omitempty"`
	CACount            int64                          `json:"CACount,omitempty"`
}

// NamespaceMetrics namespace metric
type NamespaceMetrics struct {
	Index              int                            `json:"index"`
	Time               primitive.DateTime             `json:"time"`
	CPURequest         float64                        `json:"CPURequest"`
	CPULimit           float64                        `json:"CPULimit"`
	MemoryRequest      int64                          `json:"memoryRequest"`
	MemoryLimit        int64                          `json:"memoryLimit"`
	CPUUsageAmount     float64                        `json:"CPUUsageAmount"`
	MemoryUsageAmount  int64                          `json:"memoryUsageAmount"`
	CPUUsage           float64                        `json:"CPUUsage"`
	MemoryUsage        float64                        `json:"MemoryUsage"`
	WorkloadCount      int64                          `json:"workloadCount"`
	InstanceCount      int64                          `json:"instanceCount"`
	MaxCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"maxCPUUsageTime"`
	MinCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"minCPUUsageTime"`
	MaxMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"maxMemoryUsageTime"`
	MinMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"minMemoryUsageTime"`
	MinInstanceTime    *bcsdatamanager.ExtremumRecord `json:"minInstanceTime"`
	MaxInstanceTime    *bcsdatamanager.ExtremumRecord `json:"maxInstanceTime"`
	MinWorkloadUsage   *bcsdatamanager.ExtremumRecord `json:"minWorkloadUsage"`
	MaxWorkloadUsage   *bcsdatamanager.ExtremumRecord `json:"maxWorkloadUsage"`
}

// WorkloadMetrics workload metric
type WorkloadMetrics struct {
	Index              int                            `json:"index"`
	Time               primitive.DateTime             `json:"time"`
	CPURequest         float64                        `json:"CPURequest"`
	CPULimit           float64                        `json:"CPULimit"`
	MemoryRequest      int64                          `json:"memoryRequest"`
	MemoryLimit        int64                          `json:"memoryLimit"`
	CPUUsageAmount     float64                        `json:"CPUUsageAmount"`
	MemoryUsageAmount  int64                          `json:"memoryUsageAmount"`
	CPUUsage           float64                        `json:"CPUUsage"`
	MemoryUsage        float64                        `json:"MemoryUsage"`
	InstanceCount      int64                          `json:"instanceCount"`
	MaxCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"maxCPUUsageTime"`
	MinCPUUsageTime    *bcsdatamanager.ExtremumRecord `json:"minCPUUsageTime"`
	MaxMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"maxMemoryUsageTime"`
	MinMemoryUsageTime *bcsdatamanager.ExtremumRecord `json:"minMemoryUsageTime"`
	MinInstanceTime    *bcsdatamanager.ExtremumRecord `json:"minInstanceTime"`
	MaxInstanceTime    *bcsdatamanager.ExtremumRecord `json:"maxInstanceTime"`
	MaxCPUTime         *bcsdatamanager.ExtremumRecord `json:"maxCPUTime" `
	MinCPUTime         *bcsdatamanager.ExtremumRecord `json:"minCPUTime" `
	MaxMemoryTime      *bcsdatamanager.ExtremumRecord `json:"maxMemoryTime"`
	MinMemoryTime      *bcsdatamanager.ExtremumRecord `json:"minMemoryTime" `
}

// PodAutoscalerMetrics podAutoscaler metric
type PodAutoscalerMetrics struct {
	Index                  int                `json:"index"`
	Time                   primitive.DateTime `json:"time"`
	TotalSuccessfulRescale int64              `json:"totalSuccessfulRescale"`
}

// JobCommonOpts data job common opts
type JobCommonOpts struct {
	ObjectType        string
	ProjectID         string
	ProjectCode       string
	BusinessID        string
	ClusterID         string
	ClusterType       string
	Namespace         string
	WorkloadType      string
	WorkloadName      string
	PodAutoscalerName string
	PodAutoscalerType string
	Dimension         string
	CurrentTime       time.Time
	Timestamp         int64
	Label             map[string]string
	IsBKMonitor       bool
}

// Clients clients for dataJob
type Clients struct {
	MonitorClient   bcsmonitor.ClientInterface
	K8sStorageCli   bcsapi.Storage
	MesosStorageCli bcsapi.Storage
	CmCli           *cmanager.ClusterManagerClientWithHeader
}

// NewClients init dataJob clients
func NewClients(monitorClient bcsmonitor.ClientInterface, k8sStorageCli, mesosStorageCli bcsapi.Storage,
	cmCli *cmanager.ClusterManagerClientWithHeader) *Clients {
	return &Clients{MonitorClient: monitorClient, CmCli: cmCli,
		K8sStorageCli: k8sStorageCli, MesosStorageCli: mesosStorageCli}
}

// CPUMetrics metrics of cpu
type CPUMetrics struct {
	TotalCPU   float64
	CPURequest float64
	CPULimit   float64
	CPUUsage   float64
	CPUUsed    float64
}

// MemoryMetrics metrics of memory
type MemoryMetrics struct {
	TotalMemory   int64
	MemoryRequest int64
	MemoryLimit   int64
	MemoryUsage   float64
	MemoryUsed    int64
}

// BKBaseRequestRecommendResult define workload cpu/memory from bkbase
type BKBaseRequestRecommendResult struct {
	Metric           string    `json:"metric" bson:"metric"`
	BCSClusterID     string    `json:"bcs_cluster_id" bson:"bcs_cluster_id"`
	Namespace        string    `json:"namespace" bson:"namespace"`
	WorkloadKind     string    `json:"workload_kind" bson:"workload_kind"`
	WorkloadName     string    `json:"workload_name" bson:"workload_name"`
	ContainerName    string    `json:"container_name" bson:"container_name"`
	P90              float64   `json:"p90" bson:"p90"`
	P99              float64   `json:"p99" bson:"p99"`
	MaxVal           float64   `json:"max_val" bson:"max_val"`
	TSCnt            int64     `json:"ts_cnt" bson:"ts_cnt"`
	DTEventTime      string    `json:"dtEventTime" bson:"dt_event_time"`
	DTEventTimeStamp int64     `json:"dtEventTimeStamp" bson:"dt_event_time_stamp"`
	TheDate          int64     `json:"thedate" bson:"the_date"`
	LocalTime        string    `json:"localTime" bson:"local_time"`
	CreateAt         time.Time `json:"createAt" json:"create_at"`
}

// WorkloadOriginRequestContainer container request limit
type WorkloadOriginRequestContainer struct {
	Container string `json:"container" bson:"container"`
	Request   string `json:"request" bson:"request"`
	Limit     string `json:"limit" bson:"limit"`
}

// WorkloadOriginRequestResult struct
type WorkloadOriginRequestResult struct {
	CreateTime   time.Time                                        `json:"createTime" bson:"create_time"`
	ProjectID    string                                           `json:"projectID" bson:"project_id"`
	ClusterID    string                                           `json:"clusterID" bson:"cluster_id"`
	Namespace    string                                           `json:"namespace" bson:"namespace"`
	WorkloadType string                                           `json:"workloadType" bson:"workload_type"`
	WorkloadName string                                           `json:"workloadName" bson:"workload_name"`
	Cpu          []*bcsdatamanager.WorkloadOriginRequestContainer `json:"cpu" bson:"cpu"`
	Memory       []*bcsdatamanager.WorkloadOriginRequestContainer `json:"memory" bson:"memory"`
}
