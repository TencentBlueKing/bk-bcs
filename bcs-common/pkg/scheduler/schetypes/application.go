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

package types

import (
	"fmt"
	"strconv"
	"strings"

	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	mesos_master "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos/master"
)

// executor or task default resources limits
const (
	CPUS_PER_EXECUTOR = 0.01
	CPUS_PER_TASK     = 1
	MEM_PER_EXECUTOR  = 64
	MEM_PER_TASK      = 64
	DISK_PER_EXECUTOR = 64
	DISK_PER_TASK     = 64
)

// operation operate
const (
	OPERATION_LAUNCH     = "LAUNCH"
	OPERATION_DELETE     = "DELETE"
	OPERATION_SCALE      = "SCALE"
	OPERATION_INNERSCALE = "INNERSCALE"
	OPERATION_ROLLBACK   = "ROLLBACK"
	OPERATION_RESCHEDULE = "RESCHEDULE"
	OPERATION_UPDATE     = "UPDATE"
)

// operation status
const (
	OPERATION_STATUS_INIT    = "INIT"
	OPERATION_STATUS_FINISH  = "FINISH"
	OPERATION_STATUS_FAIL    = "FAIL"
	OPERATION_STATUS_TIMEOUT = "TIMEOUT"
)

// extension for TaskState_TASK_...
const (
	Ext_TaskState_TASK_RESTARTING int32 = 101
)

// daemonset status
const (
	Daemonset_Status_Starting = "Starting"
	Daemonset_Status_Running  = "Running"
	Daemonset_Status_Abnormal = "Abnormal"
	Daemonset_Status_Deleting = "Deleting"
)

// app status
const (
	APP_STATUS_STAGING       = "Staging"
	APP_STATUS_DEPLOYING     = "Deploying"
	APP_STATUS_RUNNING       = "Running"
	APP_STATUS_FINISH        = "Finish"
	APP_STATUS_ERROR         = "Error"
	APP_STATUS_OPERATING     = "Operating"
	APP_STATUS_ROLLINGUPDATE = "RollingUpdate"
	APP_STATUS_UNKNOWN       = "Unknown"
	APP_STATUS_ABNORMAL      = "Abnormal"
)

// app status
const (
	APP_STATUS_RUNNING_STR  = "application is running"
	APP_STATUS_ABNORMAL_STR = "application is abnormal"
)

// app substatus
const (
	APP_SUBSTATUS_UNKNOWN            = "Unknown"
	APP_SUBSTATUS_ROLLINGUPDATE_DOWN = "RollingUpdateDown"
	APP_SUBSTATUS_ROLLINGUPDATE_UP   = "RollingUpdateUp"
)

// task status
const (
	TASK_STATUS_STAGING  = "Staging"
	TASK_STATUS_STARTING = "Starting"
	TASK_STATUS_RUNNING  = "Running"
	TASK_STATUS_FINISH   = "Finish"
	TASK_STATUS_ERROR    = "Error"
	TASK_STATUS_KILLING  = "Killing"
	TASK_STATUS_KILLED   = "Killed"
	TASK_STATUS_FAIL     = "Failed"
	TASK_STATUS_LOST     = "Lost"

	TASK_STATUS_RESTARTING = "Restarting"

	TASK_STATUS_UNKNOWN = "Unknown"
)

// taskgroup status
const (
	TASKGROUP_STATUS_STAGING  = "Staging"
	TASKGROUP_STATUS_STARTING = "Starting"
	TASKGROUP_STATUS_RUNNING  = "Running"
	TASKGROUP_STATUS_FINISH   = "Finish"
	TASKGROUP_STATUS_ERROR    = "Error"
	TASKGROUP_STATUS_KILLING  = "Killing"
	TASKGROUP_STATUS_KILLED   = "Killed"
	TASKGROUP_STATUS_FAIL     = "Failed"
	TASKGROUP_STATUS_LOST     = "Lost"

	TASKGROUP_STATUS_RESTARTING = "Restarting"

	TASKGROUP_STATUS_UNKNOWN = "Unknown"
)

const (
	// TASK_TEMPLATE_KEY_FORMAT xxx
	TASK_TEMPLATE_KEY_FORMAT = "${%s}"
	// TASK_TEMPLATE_KEY_PORT_FORMAT xxx
	TASK_TEMPLATE_KEY_PORT_FORMAT = "ports.%s"
	// TASK_TEMPLATE_KEY_PROCESSNAME xxx
	TASK_TEMPLATE_KEY_PROCESSNAME = "processname"
	// TASK_TEMPLATE_KEY_INSTANCEID xxx
	TASK_TEMPLATE_KEY_INSTANCEID = "instanceid"
	// TASK_TEMPLATE_KEY_HOSTIP xxx
	TASK_TEMPLATE_KEY_HOSTIP = "hostip"
	// TASK_TEMPLATE_KEY_NAMESPACE xxx
	TASK_TEMPLATE_KEY_NAMESPACE = "namespace"
	// TASK_TEMPLATE_KEY_WORKPATH xxx
	TASK_TEMPLATE_KEY_WORKPATH = "workPath"
	// TASK_TEMPLATE_KEY_PIDFILE xxx
	TASK_TEMPLATE_KEY_PIDFILE = "pidFile"
)

const (
	// APP_TASK_TEMPLATE_KEY_FORMAT xxx
	APP_TASK_TEMPLATE_KEY_FORMAT = "${%s}"
	// APP_TASK_TEMPLATE_KEY_PORT_FORMAT xxx
	APP_TASK_TEMPLATE_KEY_PORT_FORMAT = "bcs.ports.%s"
	// APP_TASK_TEMPLATE_KEY_APPNAME xxx
	APP_TASK_TEMPLATE_KEY_APPNAME = "bcs.appname"
	// APP_TASK_TEMPLATE_KEY_INSTANCEID xxx
	APP_TASK_TEMPLATE_KEY_INSTANCEID = "bcs.instanceid"
	// APP_TASK_TEMPLATE_KEY_HOSTIP xxx
	APP_TASK_TEMPLATE_KEY_HOSTIP = "bcs.hostip"
	// APP_TASK_TEMPLATE_KEY_NAMESPACE xxx
	APP_TASK_TEMPLATE_KEY_NAMESPACE = "bcs.namespace"
	// APP_TASK_TEMPLATE_KEY_PODID xxx
	APP_TASK_TEMPLATE_KEY_PODID = "bcs.taskgroupid"
	// APP_TASK_TEMPLATE_KEY_PODNAME xxx
	APP_TASK_TEMPLATE_KEY_PODNAME = "bcs.taskgroupname"
)

const (
	// MesosAttributeNoSchedule xxx
	MesosAttributeNoSchedule = "NoSchedule"
)

// Version for api resources application or deployment
type Version struct {
	ID            string
	Name          string
	ObjectMeta    commtypes.ObjectMeta
	PodObjectMeta commtypes.ObjectMeta
	Instances     int32
	RunAs         string
	Container     []*Container
	// add  20180802
	Process       []*commtypes.Process
	Labels        map[string]string
	KillPolicy    *commtypes.KillPolicy
	RestartPolicy *commtypes.RestartPolicy
	Constraints   *commtypes.Constraint
	Uris          []string
	Ip            []string
	Mode          string
	// added  20181011, add for differentiate process/application
	Kind commtypes.BcsDataType
	// commtypes.ReplicaController json
	RawJson *commtypes.ReplicaController `json:"raw_json,omitempty"`
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Version) DeepCopyInto(out *Version) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TaskGroupSpec.
func (in *Version) DeepCopy() *Version {
	if in == nil {
		return nil
	}
	out := new(Version)
	in.DeepCopyInto(out)
	return out
}

// GetExtendedResources xxx
func (in *Version) GetExtendedResources() map[string]*commtypes.ExtendedResource {
	ers := make(map[string]*commtypes.ExtendedResource)
	for _, c := range in.Container {
		for _, ex := range c.DataClass.ExtendedResources {
			o := ers[ex.Name]
			// if extended resources already exist, then superposition
			if o != nil {
				o.Value += ex.Value
			} else {
				ers[ex.Name] = ex
			}
		}
	}
	return ers
}

// Resource describe resources needed by a task
type Resource struct {
	// cpu核数
	Cpus float64
	// MB
	Mem  float64
	Disk float64
	// IOTps  uint32 //default times per second
	// IOBps  uint32 //default MB/s
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Resource) DeepCopyInto(out *Resource) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdmissionWebhookConfigurationSpec.
func (in *Resource) DeepCopy() *Resource {
	if in == nil {
		return nil
	}
	out := new(Resource)
	in.DeepCopyInto(out)
	return out
}

// CheckAndDefaultResource check the resource of each container, if no resource, set default value
func (version *Version) CheckAndDefaultResource() error {
	switch version.Kind {
	case commtypes.BcsDataType_PROCESS:
		for _, process := range version.Process {
			if process.Resources.Limits.Cpu == "" {
				process.Resources.Limits.Cpu = strconv.Itoa(CPUS_PER_TASK)
			}
			if process.Resources.Limits.Mem == "" {
				process.Resources.Limits.Mem = strconv.Itoa(MEM_PER_TASK)
			}
			if process.Resources.Limits.Storage == "" {
				process.Resources.Limits.Storage = strconv.Itoa(DISK_PER_TASK)
			}
		}
		return nil
	case commtypes.BcsDataType_APP, "":
		for index, container := range version.Container {
			if nil == container.DataClass {
				version.Container[index].DataClass = &DataClass{}
			}
			if nil == container.DataClass.Resources {
				version.Container[index].DataClass.Resources = &Resource{
					Cpus: float64(CPUS_PER_TASK),
					Mem:  float64(MEM_PER_TASK),
					Disk: float64(DISK_PER_TASK),
				}
			}
		}
		return nil
	}

	return nil
}

// CheckConstraints xxx
// check application constraints whether is valid
func (version *Version) CheckConstraints() bool {
	if version.Constraints == nil {
		return true
	}

	for _, constraint := range version.Constraints.IntersectionItem {
		if constraint == nil {
			continue
		}
		for _, oneData := range constraint.UnionData {
			if oneData == nil {
				continue
			}
			if oneData.Type == commtypes.ConstValueType_Scalar && oneData.Scalar == nil {
				return false
			}
			if oneData.Type == commtypes.ConstValueType_Text && oneData.Text == nil {
				return false
			}
			if oneData.Type == commtypes.ConstValueType_Set && oneData.Set == nil {
				return false
			}
			if oneData.Type == commtypes.ConstValueType_Range {
				for _, oneRange := range oneData.Ranges {
					if oneRange == nil {
						return false
					}
				}
			}
		}
	}

	return true
}

// AllCpus return taskgroup will use cpu resources
func (version *Version) AllCpus() float64 {
	var allCpus float64
	allCpus = 0

	// split process and containers
	switch version.Kind {
	case commtypes.BcsDataType_PROCESS:
		for _, process := range version.Process {
			cpu, _ := strconv.ParseFloat(process.Resources.Limits.Cpu, 64)
			allCpus = allCpus + cpu
		}
	case commtypes.BcsDataType_APP, "":
		for _, container := range version.Container {
			allCpus = allCpus + container.DataClass.Resources.Cpus
		}
	}
	return allCpus
}

// AllMems return taskgroup will use memory resource
func (version *Version) AllMems() float64 {
	var allMem float64
	allMem = 0

	switch version.Kind {
	case commtypes.BcsDataType_PROCESS:
		for _, process := range version.Process {
			mem, _ := strconv.ParseFloat(process.Resources.Limits.Mem, 64)
			allMem = allMem + mem
		}
	case commtypes.BcsDataType_APP, "":
		for _, container := range version.Container {
			allMem = allMem + container.DataClass.Resources.Mem
		}
	}
	return allMem + float64(MEM_PER_EXECUTOR)
}

// AllDisk return taskgroup will use disk resources
func (version *Version) AllDisk() float64 {
	var allDisk float64
	allDisk = 0

	switch version.Kind {
	case commtypes.BcsDataType_PROCESS:
		for _, process := range version.Process {
			disk, _ := strconv.ParseFloat(process.Resources.Limits.Storage, 64)
			allDisk = allDisk + disk
		}
	case commtypes.BcsDataType_APP, "":
		for _, container := range version.Container {
			allDisk = allDisk + container.DataClass.Resources.Disk
		}
	}
	return allDisk + float64(DISK_PER_EXECUTOR)
}

// AllResource return  taskgroup used cpu, memory, disk resources
func (version *Version) AllResource() *Resource {
	return &Resource{
		Cpus: version.AllCpus(),
		Mem:  version.AllMems(),
		Disk: version.AllDisk(),
	}
}

// Container for Version
type Container struct {
	Type          string
	Docker        *Docker
	Volumes       []*Volume
	Resources     *Resource
	LimitResoures *Resource
	// ExtendedResources []*commtypes.ExtendedResource
	DataClass *DataClass

	ConfigMaps []commtypes.ConfigMap
	Secrets    []commtypes.Secret

	HealthChecks []*commtypes.HealthCheck

	// network flow limit
	NetLimit *commtypes.NetLimit
}

// Docker for container
type Docker struct {
	Hostname        string
	ForcePullImage  bool
	Image           string
	ImagePullUser   string
	ImagePullPasswd string
	Network         string
	NetworkType     string
	Command         string
	Arguments       []string
	Parameters      []*Parameter
	PortMappings    []*PortMapping
	Env             map[string]string
	Privileged      bool
}

// Parameter for container
type Parameter struct {
	Key   string
	Value string
}

// PortMapping for container
type PortMapping struct {
	ContainerPort int32
	HostPort      int32
	Name          string
	Protocol      string
}

// Volume for container
type Volume struct {
	ContainerPath string
	HostPath      string
	Mode          string
}

// HealthCheck
// type HealthCheck struct {
//	ID                     string
//	Address                string
//	TaskID                 string
//	AppID                  string
//	Protocol               string
//	Port                   int32
//	PortIndex              int32
//	PortName               string
//	Command                *Command
//	Path                   string
//	MaxConsecutiveFailures uint32
//	GracePeriodSeconds     float64
//	IntervalSeconds        float64
//	TimeoutSeconds         float64
//	DelaySeconds           float64
//	ConsecutiveFailures    uint32
// }

// Command xxx
type Command struct {
	Value string
}

// Task xxx
type Task struct {
	Kind            commtypes.BcsDataType
	ID              string
	Name            string
	Hostame         string
	Command         string
	Arguments       []string
	Image           string
	ImagePullUser   string
	ImagePullPasswd string
	Network         string
	NetworkType     string
	PortMappings    []*PortMapping
	Privileged      bool
	Parameters      []*Parameter
	ForcePullImage  bool
	Volumes         []*Volume
	Env             map[string]string
	Labels          map[string]string
	DataClass       *DataClass
	// whether cpuset
	Cpuset       bool
	HealthChecks []*commtypes.HealthCheck
	// health check status
	HealthCheckStatus           []*commtypes.BcsHealthCheckStatus
	Healthy                     bool
	IsChecked                   bool
	ConsecutiveFailureTimes     uint32
	LocalMaxConsecutiveFailures uint32

	OfferId        string
	AgentId        string
	AgentHostname  string
	AgentIPAddress string
	Status         string
	LastStatus     string
	UpdateTime     int64
	StatusData     string
	AppId          string
	RunAs          string
	KillPolicy     *commtypes.KillPolicy
	Uris           []string
	LastUpdateTime int64
	Message        string
	// network flow limit
	NetLimit *commtypes.NetLimit
	// Populated by the system.
	// Read-only.
	// Value must be treated as opaque by clients and .
	ResourceVersion string `json:"-"`
}

// GetAgentIp xxx
// get taskgroup allocated node ip
func (t *TaskGroup) GetAgentIp() string {
	if len(t.Taskgroup) == 0 {
		return ""
	}

	return t.Taskgroup[0].AgentIPAddress
}

// GetRunAsAndAppIDbyTaskID xxx
// return namespace, appid
func GetRunAsAndAppIDbyTaskID(taskId string) (string, string) {
	appID := ""
	runAs := ""

	szSplit := strings.Split(taskId, ".")
	// RunAs
	if len(szSplit) >= 6 {
		runAs = szSplit[4]
	}

	// appID
	if len(szSplit) >= 6 {
		appID = szSplit[3]
	}

	return runAs, appID
}

// GetTaskGroupID xxx
// return taskgroupId by taskid
func GetTaskGroupID(taskID string) string {

	splitID := strings.Split(taskID, ".")
	if len(splitID) < 6 {
		return ""
	}
	// appInstances, appID, appRunAs, appClusterId, idTime
	taskGroupID := fmt.Sprintf("%s.%s.%s.%s.%s", splitID[2], splitID[3], splitID[4], splitID[5], splitID[0])

	return taskGroupID
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Task) DeepCopyInto(out *Task) {
	*out = *in
	if in.Arguments != nil {
		in, out := &in.Arguments, &out.Arguments
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.PortMappings != nil {
		in, out := &in.PortMappings, &out.PortMappings
		*out = make([]*PortMapping, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(PortMapping)
				**out = **in
			}
		}
	}
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = make([]*Parameter, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Parameter)
				**out = **in
			}
		}
	}
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make([]*Volume, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Volume)
				**out = **in
			}
		}
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.DataClass != nil {
		in, out := &in.DataClass, &out.DataClass
		*out = new(DataClass)
		(*in).DeepCopyInto(*out)
	}
	if in.HealthChecks != nil {
		in, out := &in.HealthChecks, &out.HealthChecks
		*out = make([]*commtypes.HealthCheck, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(commtypes.HealthCheck)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.HealthCheckStatus != nil {
		in, out := &in.HealthCheckStatus, &out.HealthCheckStatus
		*out = make([]*commtypes.BcsHealthCheckStatus, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(commtypes.BcsHealthCheckStatus)
				**out = **in
			}
		}
	}
	if in.KillPolicy != nil {
		in, out := &in.KillPolicy, &out.KillPolicy
		*out = new(commtypes.KillPolicy)
		**out = **in
	}
	if in.Uris != nil {
		in, out := &in.Uris, &out.Uris
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.NetLimit != nil {
		in, out := &in.NetLimit, &out.NetLimit
		*out = new(commtypes.NetLimit)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TaskSpec.
func (in *Task) DeepCopy() *Task {
	if in == nil {
		return nil
	}
	out := new(Task)
	in.DeepCopyInto(out)
	return out
}

// TaskGroup describes the implements of multiple tasks
type TaskGroup struct {
	Kind            commtypes.BcsDataType
	ID              string
	Name            string
	AppID           string
	RunAs           string
	ObjectMeta      commtypes.ObjectMeta
	AgentID         string
	ExecutorID      string
	Status          string
	LastStatus      string
	InstanceID      uint64
	Taskgroup       []*Task
	KillPolicy      *commtypes.KillPolicy
	RestartPolicy   *commtypes.RestartPolicy
	VersionName     string
	LastUpdateTime  int64
	Attributes      []*mesos.Attribute
	StartTime       int64
	UpdateTime      int64
	ReschededTimes  int
	LastReschedTime int64
	// we should replace the next three BcsXXX, using ObjectMeta.Labels directly
	// BcsAppID       string
	// BcsSetID       string
	// BcsModuleID    string
	HostName       string
	Message        string
	LaunchResource *Resource
	CurrResource   *Resource
	// BcsMessages map[int64]*BcsMessage
	BcsEventMsg *BcsMessage
	// Populated by the system.
	// Read-only.
	// Value must be treated as opaque by clients and .
	ResourceVersion string `json:"-"`
}

// GetExtendedResources xxx
func (in *TaskGroup) GetExtendedResources() map[string]*commtypes.ExtendedResource {
	ers := make(map[string]*commtypes.ExtendedResource)
	for _, task := range in.Taskgroup {
		for _, ex := range task.DataClass.ExtendedResources {
			o := ers[ex.Name]
			// if extended resources already exist, then superposition
			if o != nil {
				o.Value += ex.Value
			} else {
				ers[ex.Name] = ex
			}
		}
	}
	return ers
}

// GetRunAsAndAppIDbyTaskGroupID xxx
// return namespace, appid
func GetRunAsAndAppIDbyTaskGroupID(taskGroupId string) (string, string) {
	appID := ""
	runAs := ""

	szSplit := strings.Split(taskGroupId, ".")
	// RunAs
	if len(szSplit) >= 3 {
		runAs = szSplit[2]
	}

	// appID
	if len(szSplit) >= 2 {
		appID = szSplit[1]
	}

	return runAs, appID
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TaskGroup) DeepCopyInto(out *TaskGroup) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Taskgroup != nil {
		in, out := &in.Taskgroup, &out.Taskgroup
		*out = make([]*Task, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Task)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.KillPolicy != nil {
		in, out := &in.KillPolicy, &out.KillPolicy
		*out = new(commtypes.KillPolicy)
		**out = **in
	}
	if in.RestartPolicy != nil {
		in, out := &in.RestartPolicy, &out.RestartPolicy
		*out = new(commtypes.RestartPolicy)
		**out = **in
	}
	if in.LaunchResource != nil {
		in, out := &in.LaunchResource, &out.LaunchResource
		*out = (*in).DeepCopy()
	}
	if in.CurrResource != nil {
		in, out := &in.CurrResource, &out.CurrResource
		*out = (*in).DeepCopy()
	}
	/*if in.BcsEventMsg != nil {
		in, out := &in.BcsEventMsg, &out.BcsEventMsg
		*out = (*in).DeepCopy()
	}*/
	// there are no externally modified fields, so deepCopy is not required
	/*if in.Attributes != nil {
		in, out := &in.Attributes, &out.Attributes
		*out = make([]*mesos.Attribute, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(mesos.Attribute)
				err := deepcopy.DeepCopy(out, in)
				if err != nil {
					fmt.Println("DeepCopy TaskGroup.Attributes", "failed", err.Error())
				}
			}
		}
	}*/
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TaskGroupSpec.
func (in *TaskGroup) DeepCopy() *TaskGroup {
	if in == nil {
		return nil
	}
	out := new(TaskGroup)
	in.DeepCopyInto(out)
	return out
}

// Application for container
type Application struct {
	Kind             commtypes.BcsDataType
	ID               string
	Name             string
	ObjectMeta       commtypes.ObjectMeta
	DefineInstances  uint64
	Instances        uint64
	RunningInstances uint64
	RunAs            string
	ClusterId        string
	Status           string
	SubStatus        string
	LastStatus       string
	Created          int64
	UpdateTime       int64
	Mode             string
	LastUpdateTime   int64
	// we should replace the next three BcsXXX, using ObjectMeta.Labels directly
	// BcsAppID    string
	// BcsSetID    string
	// BcsModuleID string
	Message string
	Pods    []*commtypes.BcsPodIndex
	// Populated by the system.
	// Read-only.
	// Value must be treated as opaque by clients and .
	ResourceVersion string `json:"-"`
	// RC current original definition
	// RawJson []byte `json:"raw_json,omitempty"`
}

// GetUuid xxx
func (in *Application) GetUuid() string {
	uuid := fmt.Sprintf("%s.%s", in.RunAs, in.ID)
	return uuid
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Application) DeepCopyInto(out *Application) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Pods != nil {
		in, out := &in.Pods, &out.Pods
		*out = make([]*commtypes.BcsPodIndex, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(commtypes.BcsPodIndex)
				**out = **in
			}
		}
	}

	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AgentSchedInfoSpec.
func (in *Application) DeepCopy() *Application {
	if in == nil {
		return nil
	}
	out := new(Application)
	in.DeepCopyInto(out)
	return out
}

// Operation for application
type Operation struct {
	ID             string
	RunAs          string
	AppID          string
	OperationType  string
	Status         string
	CreateTime     int64
	LastUpdateTime int64
	ErrorStr       string
}

// OperationIndex xxx
type OperationIndex struct {
	Operation string
}

// Agent xxx
// mesos slave info
type Agent struct {
	Key          string
	LastSyncTime int64
	AgentInfo    *mesos_master.Response_GetAgents_Agent
	// Populated by the system.
	// Read-only.
	// Value must be treated as opaque by clients and .
	ResourceVersion string `json:"-"`
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Agent) DeepCopyInto(out *Agent) {
	*out = *in
	/*if in.AgentInfo != nil {
		in, out := &in.AgentInfo, &out.AgentInfo
		*out = new(mesos_master.Response_GetAgents_Agent)
		err := deepcopy.DeepCopy(out, in)
		if err != nil {
			fmt.Println("deepcopy Agent", "failed", err.Error())
		}
	}*/

	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdmissionWebhookConfigurationSpec.
func (in *Agent) DeepCopy() *Agent {
	if in == nil {
		return nil
	}
	out := new(Agent)
	in.DeepCopyInto(out)
	return out
}

// GetAgentInfo xxx
func (om *Agent) GetAgentInfo() *commtypes.BcsClusterAgentInfo {
	agent := new(commtypes.BcsClusterAgentInfo)
	// blog.V(3).Infof("get agents: ===>agent[%d]: %+v", index, om.AgentInfo)
	agent.HostName = om.AgentInfo.GetAgentInfo().GetHostname()

	szSplit := strings.Split(om.AgentInfo.GetPid(), "@")
	if len(szSplit) == 2 {
		agent.IP = szSplit[1]
	} else {
		agent.IP = om.AgentInfo.GetPid()
	}
	if strings.Contains(agent.IP, ":") {
		agent.IP = strings.Split(agent.IP, ":")[0]
	}

	totalRes := om.AgentInfo.GetTotalResources()
	for _, resource := range totalRes {
		if resource.GetName() == "cpus" {
			agent.CpuTotal = resource.GetScalar().GetValue()
		}
		if resource.GetName() == "mem" {
			agent.MemTotal = resource.GetScalar().GetValue()
		}
		if resource.GetName() == "disk" {
			agent.DiskTotal = resource.GetScalar().GetValue()
		}
	}

	usedRes := om.AgentInfo.GetAllocatedResources()
	for _, resource := range usedRes {
		if resource.GetName() == "cpus" {
			agent.CpuUsed = resource.GetScalar().GetValue()
		}
		if resource.GetName() == "mem" {
			agent.MemUsed = resource.GetScalar().GetValue()
		}
		if resource.GetName() == "disk" {
			agent.DiskUsed = resource.GetScalar().GetValue()
		}
	}

	agent.HostAttributes = mesosAttribute2commonAttribute(om.AgentInfo.AgentInfo.Attributes)
	agent.Attributes = agent.HostAttributes

	if om.AgentInfo.RegisteredTime != nil && om.AgentInfo.RegisteredTime.Nanoseconds != nil {
		agent.RegisteredTime = *om.AgentInfo.RegisteredTime.Nanoseconds
	}
	if om.AgentInfo.ReregisteredTime != nil && om.AgentInfo.ReregisteredTime.Nanoseconds != nil {
		agent.ReRegisteredTime = *om.AgentInfo.ReregisteredTime.Nanoseconds
	}

	return agent
}

// GetAgentIP xxx
func (om *Agent) GetAgentIP() string {
	if om.AgentInfo == nil || om.AgentInfo.AgentInfo == nil {
		return ""
	}
	if len(om.AgentInfo.AgentInfo.Attributes) == 0 {
		return ""
	}
	for _, attr := range om.AgentInfo.AgentInfo.Attributes {
		if attr.GetName() == "InnerIP" {
			return attr.GetText().GetValue()
		}
	}

	return ""
}

func mesosAttribute2commonAttribute(oldAttributeList []*mesos.Attribute) []*commtypes.BcsAgentAttribute {
	if oldAttributeList == nil {
		return nil
	}

	attributeList := make([]*commtypes.BcsAgentAttribute, 0)

	for _, oldAttribute := range oldAttributeList {
		if oldAttribute == nil {
			continue
		}

		attribute := new(commtypes.BcsAgentAttribute)
		if oldAttribute.Name != nil {
			attribute.Name = *oldAttribute.Name
		}
		if oldAttribute.Type != nil {
			switch *oldAttribute.Type {
			case mesos.Value_SCALAR:
				attribute.Type = commtypes.MesosValueType_Scalar
				if oldAttribute.Scalar != nil && oldAttribute.Scalar.Value != nil {
					attribute.Scalar = &commtypes.MesosValue_Scalar{
						Value: *oldAttribute.Scalar.Value,
					}
				}
			case mesos.Value_RANGES:
				attribute.Type = commtypes.MesosValueType_Ranges
				if oldAttribute.Ranges != nil {
					rangeList := make([]*commtypes.MesosValue_Ranges, 0)
					for _, oldRange := range oldAttribute.Ranges.Range {
						newRange := &commtypes.MesosValue_Ranges{}
						if oldRange.Begin != nil {
							newRange.Begin = *oldRange.Begin
						}
						if oldRange.End != nil {
							newRange.End = *oldRange.End
						}
						rangeList = append(rangeList, newRange)
					}
				}
			case mesos.Value_SET:
				attribute.Type = commtypes.MesosValueType_Set
				if oldAttribute.Set != nil {
					attribute.Set = &commtypes.MesosValue_Set{
						Item: oldAttribute.Set.Item,
					}
				}
			case mesos.Value_TEXT:
				attribute.Type = commtypes.MesosValueType_Text
				if oldAttribute.Text != nil && oldAttribute.Text.Value != nil {
					attribute.Text = &commtypes.MesosValue_Text{
						Value: *oldAttribute.Text.Value,
					}
				}
			}
		}
		attributeList = append(attributeList, attribute)
	}
	return attributeList
}

// Check xxx
type Check struct {
	ID          string   `json:"id"`
	Protocol    string   `json:"protocol"`
	Address     string   `json:"address"`
	Port        int      `json:"port"`
	Command     *Command `json:"command"`
	Path        string   `json:"path"`
	MaxFailures int      `json:"max_failures"`
	Interval    int      `json:"interval"`
	Timeout     int      `json:"timeout"`
	TaskID      string   `json:"task_id"`
	TaskGroupID string   `json:"taskgroup_id"`
	AppID       string   `json:"app_id"`
}

// ProcDef xxx
type ProcDef struct {
	ProcName   string           `json:"procName"`
	WorkPath   string           `json:"workPath"`
	PidFile    string           `json:"pidFile"`
	StartCmd   string           `json:"startCmd"`
	CheckCmd   string           `json:"checkCmd"`
	StopCmd    string           `json:"stopCmd"`
	RestartCmd string           `json:"restartCmd"`
	ReloadCmd  string           `json:"reloadCmd"`
	KillCmd    string           `json:"killCmd"`
	LogPath    string           `json:"logPath"`
	CfgPath    string           `json:"cfgPath"`
	Uris       []*commtypes.Uri `json:"uris"`
	// seconds
	StartGracePeriod int `json:"startGracePeriod"`
}

// DataClass xxx
type DataClass struct {
	// resources request cpu\memory
	Resources *Resource
	// resources limit cpu\memory
	LimitResources *Resource
	// extended resources, key=ExtendedResource.Name
	ExtendedResources []*commtypes.ExtendedResource
	Msgs              []*BcsMessage
	NetLimit          *commtypes.NetLimit
	// add for proc 20180730
	ProcInfo *ProcDef
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataClass) DeepCopyInto(out *DataClass) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(Resource)
		**out = **in
	}
	if in.LimitResources != nil {
		in, out := &in.LimitResources, &out.LimitResources
		*out = new(Resource)
		**out = **in
	}
	if in.NetLimit != nil {
		in, out := &in.NetLimit, &out.NetLimit
		*out = new(commtypes.NetLimit)
		**out = **in
	}
	if in.ProcInfo != nil {
		in, out := &in.ProcInfo, &out.ProcInfo
		*out = new(ProcDef)
		(*in).DeepCopyInto(*out)
	}
	/*if in.Msgs != nil {
		in, out := &in.Msgs, &out.Msgs
		*out = make([]*BcsMessage, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(BcsMessage)
				err := deepcopy.DeepCopy(out, in)
				if err != nil {
					fmt.Println("DeepCopy DataClass.BcsMessage", "failed", err.Error())
				}
			}
		}
	}*/
	if in.ExtendedResources != nil {
		in, out := &in.ExtendedResources, &out.ExtendedResources
		*out = make([]*commtypes.ExtendedResource, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(commtypes.ExtendedResource)
				**out = **in
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataClass.
func (in *DataClass) DeepCopy() *DataClass {
	if in == nil {
		return nil
	}
	out := new(DataClass)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProcDef) DeepCopyInto(out *ProcDef) {
	*out = *in
	if in.Uris != nil {
		in, out := &in.Uris, &out.Uris
		*out = make([]*commtypes.Uri, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(commtypes.Uri)
				**out = **in
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProcDef.
func (in *ProcDef) DeepCopy() *ProcDef {
	if in == nil {
		return nil
	}
	out := new(ProcDef)
	in.DeepCopyInto(out)
	return out
}

// DeploymentDef xxx
type DeploymentDef struct {
	ObjectMeta commtypes.ObjectMeta      `json:"metadata"`
	Selector   map[string]string         `json:"selector,omitempty"`
	Version    *Version                  `json:"version"`
	Strategy   commtypes.UpgradeStrategy `json:"strategy"`
	// BcsDeployment original definition
	RawJson *commtypes.BcsDeployment `json:"raw_json,omitempty"`
}

const (
	// DEPLOYMENT_STATUS_DEPLOYING xxx
	DEPLOYMENT_STATUS_DEPLOYING = "Deploying"
	// DEPLOYMENT_STATUS_RUNNING xxx
	DEPLOYMENT_STATUS_RUNNING = "Running"
	// DEPLOYMENT_STATUS_ROLLINGUPDATE xxx
	DEPLOYMENT_STATUS_ROLLINGUPDATE = "Update"
	// DEPLOYMENT_STATUS_ROLLINGUPDATE_PAUSED xxx
	DEPLOYMENT_STATUS_ROLLINGUPDATE_PAUSED = "UpdatePaused"
	// DEPLOYMENT_STATUS_ROLLINGUPDATE_SUSPEND xxx
	DEPLOYMENT_STATUS_ROLLINGUPDATE_SUSPEND = "UpdateSuspend"
	// DEPLOYMENT_STATUS_DELETING xxx
	DEPLOYMENT_STATUS_DELETING = "Deleting"
	// DEPLOYMENT_STATUS_UPDATERESOURCE xxx
	DEPLOYMENT_STATUS_UPDATERESOURCE = "UpdateResource"
)

const (
	// DEPLOYMENT_OPERATION_NIL xxx
	DEPLOYMENT_OPERATION_NIL = ""
	// DEPLOYMENT_OPERATION_DELETE xxx
	DEPLOYMENT_OPERATION_DELETE = "DELETE"
	// DEPLOYMENT_OPERATION_START xxx
	DEPLOYMENT_OPERATION_START = "START"
)

// Deployment xxx
type Deployment struct {
	ObjectMeta      commtypes.ObjectMeta        `json:"metadata"`
	Selector        map[string]string           `json:"selector,omitempty"`
	Strategy        commtypes.UpgradeStrategy   `json:"strategy"`
	Status          string                      `json:"status"`
	Application     *DeploymentReferApplication `json:"application"`
	ApplicationExt  *DeploymentReferApplication `json:"application_ext"`
	LastRollingTime int64                       `json:"last_rolling_time"`
	CurrRollingOp   string                      `json:"curr_rolling_operation"`
	IsInRolling     bool                        `json:"is_in_rolling"`
	CheckTime       int64                       `json:"check_time"`
	Message         string                      `json:"message"`
	// BcsDeployment current original definition
	RawJson *commtypes.BcsDeployment `json:"raw_json,omitempty"`
	// BcsDeployment old version original definition
	RawJsonBackup *commtypes.BcsDeployment `json:"raw_json_backup,omitempty"`
}

// DeploymentReferApplication xxx
type DeploymentReferApplication struct {
	ApplicationName         string `json:"name"`
	CurrentTargetInstances  int    `json:"curr_target_instances"`
	CurrentRollingInstances int    `josn:"curr_rolling_instances"`
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Deployment) DeepCopyInto(out *Deployment) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.Strategy.DeepCopyInto(&out.Strategy)
	if in.Application != nil {
		in, out := &in.Application, &out.Application
		*out = new(DeploymentReferApplication)
		**out = **in
	}
	if in.ApplicationExt != nil {
		in, out := &in.ApplicationExt, &out.ApplicationExt
		*out = new(DeploymentReferApplication)
		**out = **in
	}
	/*if in.RawJson != nil {
		deepcopy.DeepCopy(out.RawJson, in.RawJson)
	}
	if in.RawJsonBackup != nil {
		deepcopy.DeepCopy(out.RawJsonBackup, in.RawJsonBackup)
	}*/

	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeploymentSpec.
func (in *Deployment) DeepCopy() *Deployment {
	if in == nil {
		return nil
	}
	out := new(Deployment)
	in.DeepCopyInto(out)
	return out
}

// AgentSchedInfo xxx
type AgentSchedInfo struct {
	HostName   string  `json:"host_name"`
	DeltaCPU   float64 `json:"delta_cpu"`
	DeltaMem   float64 `json:"delta_mem"`
	DeltaDisk  float64 `json:"delta_disk"`
	Taskgroups map[string]*Resource
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AgentSchedInfo) DeepCopyInto(out *AgentSchedInfo) {
	*out = *in
	if in.Taskgroups != nil {
		in, out := &in.Taskgroups, &out.Taskgroups
		*out = make(map[string]*Resource, len(*in))
		for key, val := range *in {
			var outVal *Resource
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(Resource)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AgentSchedInfoSpec.
func (in *AgentSchedInfo) DeepCopy() *AgentSchedInfo {
	if in == nil {
		return nil
	}
	out := new(AgentSchedInfo)
	in.DeepCopyInto(out)
	return out
}

// TaskGroupOpResult xxx
type TaskGroupOpResult struct {
	ID     string
	Status string
	Err    string
}
