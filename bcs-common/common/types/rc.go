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

//EnvVar represents an environment variable present in a Container
type EnvVar struct {
	Name      string        `json:"name"`
	Value     string        `json:"value,omitempty"`
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty"`
}

// EnvVarSource represents a source for the value of an EnvVar.
type EnvVarSource struct {
	// Selects a resource of the container: only resources limits and requests
	// (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.
	// +optional
	ResourceFieldRef *ResourceFieldSelector `json:"resourceFieldRef,omitempty" protobuf:"bytes,2,opt,name=resourceFieldRef"`
}

// ResourceFieldSelector represents container resources (cpu, memory) and their output format
type ResourceFieldSelector struct {
	// Required: resource to select
	Resource string `json:"resource" protobuf:"bytes,2,opt,name=resource"`
}

//ContainerPort represents a network port in a single container
type ContainerPort struct {
	//each named port in a pod must have a unique name.
	Name          string `json:"name"`
	HostPort      int    `json:"hostPort,omitempty"`
	ContainerPort int    `json:"containerPort"`
	HostIP        string `json:"hostIP,omitempty"`
	Protocol      string `json:"protocol"`
}

type HostPathVolumeSource struct {
	Path string `json:"path"`
}

type SecretItem struct {
	Type      DataUsageType `json:"type,omitempty"`
	DataKey   string        `json:"dataKey,omitempty"`
	KeyOrPath string        `json:"keyOrPath,omitempty"`
	SubPath   string        `json:"subPath,omitempty"`
	ReadOnly  bool          `json:"readOnly,omitempty"`
	User      string        `json:"user,omitempty"`
}

type Secret struct {
	SecretName string       `json:"secretName,omitempty"`
	Items      []SecretItem `json:"items,omitempty"`
}

type SecretUnit struct {
	Name   string `json:"name"`
	Secret Secret `json:"secret"`
}

type ConfigMap struct {
	Name  string      `json:"name"`
	Items []KeyToPath `json:"items,omitempty"`
}

type DataUsageType string

const (
	DataUsageType_UNKNOWN DataUsageType = "unknown"
	DataUsageType_ENV     DataUsageType = "env"
	DataUsageType_FILE    DataUsageType = "file"
)

type ConfigMapType string

const (
	ConfigMap_Local   ConfigMapType = "local"
	ConfigMap_Remote  ConfigMapType = "remote"
	ConfigMap_Env     ConfigMapType = "env"
	ConfigMap_FileEnv ConfigMapType = "file_env"
)

type ConfigMapFileRight string

const (
	ConfigMapFileRight_R  ConfigMapFileRight = "r"
	ConfigMapFileRight_RW ConfigMapFileRight = "rw"
)

type KeyToPath struct {
	// the key to project
	Type DataUsageType `json:"type"`
	// in kubernetes, DataKey and DataKeyAlias must be used at the same time.
	// and you can also choose not to use them. Once you have used them both, you have to know that
	// only this "DataKey" will be mount into your container, the other "DataKey" will be ignored.
	// for example, you have an configmap named example-configmap, which have keys like as follows:
	// key1: "data for key1"
	// key2: "data for key2"
	// if you set Datakey = "key1", and DataKeyAlias = "alias/key1alias.txt", and KeyOrPath = /home/mountexample,
	// then in your container you will find only "one" mounted file in path
	// "/home/mountexample/alias/keyalias.txt" and the file content is "data for key1."
	DataKey      string `json:"dataKey"`             //configmap sub item name
	DataKeyAlias string `json:"dataKeyAlias"`        //k8s only
	KeyOrPath    string `json:"keyOrPath,omitempty"` //indexs means ENV key, path means containerpath
	SubPath      string `json:"subPath,omitempty"`
	ReadOnly     bool   `json:"readOnly,omitempty"`
	User         string `json:"user,omitempty"`
}

type ConfigMapUnit struct {
	Name      string    `json:"name"`
	ConfigMap ConfigMap `json:"configmap"`
}

type Volume struct {
	HostPath  string `json:"hostPath,omitempty"`
	MountPath string `json:"mountPath,omitempty"`
	SubPath   string `json:"subPath,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}

type VolumeUnit struct {
	//Each volume in a pod must have a unique name
	Name   string `json:"name"`
	Volume Volume `json:"volume,omitempty"`
}

// ResourceRequirements describes the compute resource requirement
type ResourceRequirements struct {
	Limits            ResourceList        `json:"limits,omitempty"`
	Requests          ResourceList        `json:"requests,omitempty"`
	ExtendedResources []*ExtendedResource `json:"extendedResources,omitempty"`
}

type ResourceList struct {
	Cpu     string `json:"cpu,omitempty"`
	Mem     string `json:"memory,omitempty"`
	Storage string `json:"storage,omitempty"`
}

type ImagePullPolicyType string

const (
	ImagePullPolicy_ALWAYS       ImagePullPolicyType = "Always"
	ImagePullPolicy_IFNOTPRESENT ImagePullPolicyType = "IfNotPresent"
	ImagePullPolicy_NEVER        ImagePullPolicyType = "Never"
)

type BcsHealthCheckType string

const (
	BcsHealthCheckType_COMMAND    BcsHealthCheckType = "COMMAND"
	BcsHealthCheckType_HTTP       BcsHealthCheckType = "HTTP"
	BcsHealthCheckType_TCP        BcsHealthCheckType = "TCP"
	BcsHealthCheckType_REMOTEHTTP BcsHealthCheckType = "REMOTE_HTTP"
	BcsHealthCheckType_REMOTETCP  BcsHealthCheckType = "REMOTE_TCP"
)

// a single container that is expected to be run on the host
type Container struct {
	Name            string               `json:"name,omitempty"`
	Hostname        string               `json:"hostname,omitempty"`
	Command         string               `json:"command,omitempty"`
	Args            []string             `json:"args,omitempty"`
	Parameters      []*KeyToValue        `json:"parameters,omitempty"`
	Type            string               `json:"type,omitempty"`
	Image           string               `json:"image"`
	ImagePullUser   string               `json:"imagePullUser"`
	ImagePullPasswd string               `json:"imagePullPasswd"`
	ImagePullPolicy ImagePullPolicyType  `json:"imagePullPolicy"`
	Privileged      bool                 `json:"privileged"`
	Env             []EnvVar             `json:"env,omitempty"`
	WorkingDir      string               `json:"workingDir,omitempty"`
	Ports           []ContainerPort      `json:"ports,omitempty"`
	HealthChecks    []*HealthCheck       `json:"healthChecks,omitempty"`
	Resources       ResourceRequirements `json:"resources,omitempty"`
	Volumes         []VolumeUnit         `json:"volumes,omitempty"`
	ConfigMaps      []ConfigMap          `json:"configmaps,omitempty"`
	Secrets         []Secret             `json:"secrets,omitempty"`
}

// a single process that is expected to be run on the host
type Process struct {
	ProcName   string `json:"procName"`
	WorkPath   string `json:"workPath"`
	Uris       []*Uri `json:"uris"`
	PidFile    string `json:"pidFile"`
	StartCmd   string `json:"startCmd"`
	CheckCmd   string `json:"checkCmd"`
	StopCmd    string `json:"stopCmd"`
	RestartCmd string `json:"restartCmd"`
	ReloadCmd  string `json:"reloadCmd"`
	KillCmd    string `json:"killCmd"`
	LogPath    string `json:"logPath"`
	CfgPath    string `json:"cfgPath"`
	User       string `json:"user"`
	ProcGroup  string `json:"procGroup"`

	Env          []EnvVar             `json:"env,omitempty"`
	HealthChecks []*HealthCheck       `json:"healthChecks,omitempty"`
	Resources    ResourceRequirements `json:"resources,omitempty"`
	ConfigMaps   []ConfigMap          `json:"configmaps,omitempty"`
	Secrets      []Secret             `json:"secrets,omitempty"`
	Ports        []ContainerPort      `json:"ports,omitempty"`

	StartGracePeriod int `json:"startGracePeriod"`
}

type Uri struct {
	Value     string //process package registry uri, example for "http://xxx.registry.xxx.com/xxx/v1/pack.tar.gz"
	User      string //package registry user
	Pwd       string //package registry password, example for curl -u 'user:pwd' -X GET "http://xxx.registry.xxx.com/xxx/v1/pack.tar.gz"
	OutputDir string
}

type PodSpec struct {
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Containers   []Container       `json:"containers,omitempty"`
	Processes    []Process         `json:"processes,omitempty"`
	NetworkMode  string            `json:"networkMode,omitempty"`
	NetworkType  string            `json:"networktype,omitempty"`
	NetLimit     *NetLimit         `json:"netLimit,omitempty"`
}

//PodTemplateSpec specification for pod
type PodTemplateSpec struct {
	// Metadata of the pods created from this template
	ObjectMeta `json:"metadata,omitempty"`
	// Spec defines the behavior of a pod
	PodSpec PodSpec `json:"spec,omitempty"`
}

type ReplicaControllerSpec struct {
	Instance int               `json:"instance"`
	Selector map[string]string `json:"selector,omitempty"`
	Template *PodTemplateSpec  `json:"template"`
}

type ReplicaController struct {
	TypeMeta `json:",inline"`
	//AppMeta               `json:",inline"`
	ObjectMeta            `json:"metadata"`
	ReplicaControllerSpec ReplicaControllerSpec `json:"spec"`
	UpPolicy              UpdatePolicy          `json:"updatePolicy,omitempty"`
	RestartPolicy         RestartPolicy         `json:"restartPolicy,omitempty"`
	KillPolicy            KillPolicy            `json:"killPolicy,omitempty"`
	Constraints           *Constraint           `json:"constraint,omitempty"`
}

type HealthCheck struct {
	Type                BcsHealthCheckType  `json:"type,omitempty"`
	DelaySeconds        int                 `json:"delaySeconds,omitempty"`
	GracePeriodSeconds  int                 `json:"gracePeriodSeconds,omitempty"`
	IntervalSeconds     int                 `json:"intervalSeconds,omitempty"`
	TimeoutSeconds      int                 `json:"timeoutSeconds,omitempty"`
	ConsecutiveFailures uint32              `json:"consecutiveFailures,omitempty"`
	Command             *CommandHealthCheck `json:"command,omitempty"`
	Http                *HttpHealthCheck    `json:"http,omitempty"`
	Tcp                 *TcpHealthCheck     `json:"tcp,omitempty"`
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HealthCheck) DeepCopyInto(out *HealthCheck) {
	*out = *in
	if in.Command != nil {
		in, out := &in.Command, &out.Command
		*out = new(CommandHealthCheck)
		**out = **in
	}
	if in.Http != nil {
		in, out := &in.Http, &out.Http
		*out = new(HttpHealthCheck)
		(*in).DeepCopyInto(*out)
	}
	if in.Tcp != nil {
		in, out := &in.Tcp, &out.Tcp
		*out = new(TcpHealthCheck)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HealthCheck.
func (in *HealthCheck) DeepCopy() *HealthCheck {
	if in == nil {
		return nil
	}
	out := new(HealthCheck)
	in.DeepCopyInto(out)
	return out
}

type CommandHealthCheck struct {
	Value string `json:"value,omitempty"`
}

type HttpHealthCheck struct {
	Port     int32             `json:"port"`
	PortName string            `json:"portName"`
	Scheme   string            `json:"scheme"`
	Path     string            `json:"path"`
	Headers  map[string]string `json:"headers"`
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HttpHealthCheck) DeepCopyInto(out *HttpHealthCheck) {
	*out = *in
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HttpHealthCheck.
func (in *HttpHealthCheck) DeepCopy() *HttpHealthCheck {
	if in == nil {
		return nil
	}
	out := new(HttpHealthCheck)
	in.DeepCopyInto(out)
	return out
}

type TcpHealthCheck struct {
	Port     int32  `json:"port"`
	PortName string `json:"portName"`
}

type HealthCheckResult struct {
	ID   string             `json:"id,omitempty"`
	Type BcsHealthCheckType `json:"type,omitempty"`

	Http *HttpHealthCheck `json:"http,omitempty"`
	Tcp  *TcpHealthCheck  `json:"tcp,omitempty"`

	Status  bool   `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

type UpdatePolicy struct {
	UpdateDelay  int    `json:"updateDelay,omitempty"`
	MaxRetries   int    `json:"maxRetries,omitempty"`
	MaxFailovers int    `json:"maxFailovers,omitempty"`
	Action       string `json:"action,omitempty"`
}

//RestartPolicyType type for restart strategy
type RestartPolicyType string

const (
	RestartPolicy_NEVER     RestartPolicyType = "Never"
	RestartPolicy_ALWAYS    RestartPolicyType = "Always"
	RestartPolicy_ONFAILURE RestartPolicyType = "OnFailure"
)

//RestartPolicy for pod
type RestartPolicy struct {
	Policy         RestartPolicyType `json:"policy"`                   //value: Nerver | Always | OnFailure
	Interval       int               `json:"interval,omitempty"`       //only for mesos
	Backoff        int               `json:"backoff,omitempty"`        //only for mesos
	MaxTimes       int               `json:"maxtimes,omitempty"`       //only for mesos
	HostRetainTime int64             `json:"hostRetainTime,omitempty"` //only for mesos
}

//KillPolicy for container
type KillPolicy struct {
	GracePeriod int64 `json:"gracePeriod"` //seconds
}

//network flow limit for container
type NetLimit struct {
	EgressLimit int `json:"egressLimit"`
}
