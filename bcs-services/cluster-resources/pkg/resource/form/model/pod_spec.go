/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package model

// NodeSelect ...
type NodeSelect struct {
	Type     string         `structs:"type"`
	NodeName string         `structs:"nodeName"`
	Selector []NodeSelector `structs:"selector"`
}

// NodeSelector ...
type NodeSelector struct {
	Key   string `structs:"key"`
	Value string `structs:"value"`
}

// Affinity ...
type Affinity struct {
	NodeAffinity []NodeAffinity `structs:"nodeAffinity"`
	PodAffinity  []PodAffinity  `structs:"podAffinity"`
}

// NodeAffinity ...
type NodeAffinity struct {
	Priority string               `structs:"priority"`
	Weight   int64                `structs:"weight"`
	Selector NodeAffinitySelector `structs:"selector"`
}

// NodeAffinitySelector ...
type NodeAffinitySelector struct {
	Expressions []ExpSelector   `structs:"expressions"`
	Fields      []FieldSelector `structs:"fields"`
}

// ExpSelector ...
type ExpSelector struct {
	Key    string `structs:"key"`
	Op     string `structs:"op"`
	Values string `structs:"values"`
}

// FieldSelector ...
type FieldSelector struct {
	Key    string `structs:"key"`
	Op     string `structs:"op"`
	Values string `structs:"values"`
}

// PodAffinity ...
type PodAffinity struct {
	Type        string              `structs:"type"`
	Priority    string              `structs:"priority"`
	Namespaces  []string            `structs:"namespaces"`
	Weight      int64               `structs:"weight"`
	TopologyKey string              `structs:"topologyKey"`
	Selector    PodAffinitySelector `structs:"selector"`
}

// PodAffinitySelector ...
type PodAffinitySelector struct {
	Expressions []ExpSelector   `structs:"expressions"`
	Labels      []LabelSelector `structs:"labels"`
}

// LabelSelector ...
type LabelSelector struct {
	Key   string `structs:"key"`
	Value string `structs:"value"`
}

// Toleration ...
type Toleration struct {
	Rules []TolerationRule `structs:"rules"`
}

// TolerationRule ...
type TolerationRule struct {
	Key            string `structs:"key"`
	Op             string `structs:"op" mapstructure:"operator"`
	Value          string `structs:"value"`
	Effect         string `structs:"effect"`
	TolerationSecs int64  `structs:"tolerationSecs" mapstructure:"tolerationSeconds"`
}

// Networking ...
type Networking struct {
	DNSPolicy             string           `structs:"dnsPolicy"`
	HostIPC               bool             `structs:"hostIPC"`
	HostNetwork           bool             `structs:"hostNetwork"`
	HostPID               bool             `structs:"hostPID"`
	ShareProcessNamespace bool             `structs:"shareProcessNamespace"`
	HostName              string           `structs:"hostName"`
	Subdomain             string           `structs:"subdomain"`
	NameServers           []string         `structs:"nameServers"`
	Searches              []string         `structs:"searches"`
	DNSResolverOpts       []DNSResolverOpt `structs:"dnsResolverOpts"`
	HostAliases           []HostAlias      `structs:"hostAliases"`
}

// DNSResolverOpt ...
type DNSResolverOpt struct {
	Name  string `structs:"name"`
	Value string `structs:"value"`
}

// HostAlias ...
type HostAlias struct {
	IP    string `structs:"ip"`
	Alias string `structs:"alias"`
}

// PodSecurityCtx ...
type PodSecurityCtx struct {
	RunAsUser    int64      `structs:"runAsUser"`
	RunAsNonRoot bool       `structs:"runAsNonRoot"`
	RunAsGroup   int64      `structs:"runAsGroup"`
	FSGroup      int64      `structs:"fsGroup"`
	SELinuxOpt   SELinuxOpt `structs:"seLinuxOpt" mapstructure:"seLinuxOptions"`
}

// SELinuxOpt ...
type SELinuxOpt struct {
	Level string `structs:"level"`
	Role  string `structs:"role"`
	Type  string `structs:"type"`
	User  string `structs:"user"`
}

// SpecOther ...
type SpecOther struct {
	RestartPolicy              string   `structs:"restartPolicy"`
	TerminationGracePeriodSecs int64    `structs:"terminationGracePeriodSecs"`
	ImagePullSecrets           []string `structs:"imagePullSecrets"`
	SAName                     string   `structs:"saName"`
}

// WorkloadVolume ...
type WorkloadVolume struct {
	PVC       []PVCVolume      `structs:"pvc"`
	HostPath  []HostPathVolume `structs:"hostPath"`
	ConfigMap []CMVolume       `structs:"configMap"`
	Secret    []SecretVolume   `structs:"secret"`
	EmptyDir  []EmptyDirVolume `structs:"emptyDir"`
	NFS       []NFSVolume      `structs:"nfs"`
}

// PVCVolume ...
type PVCVolume struct {
	Name     string `structs:"name"`
	PVCName  string `structs:"pvcName"`
	ReadOnly bool   `structs:"readOnly"`
}

// HostPathVolume ...
type HostPathVolume struct {
	Name string `structs:"name"`
	Path string `structs:"path"`
	Type string `structs:"type"`
}

// CMVolume ...
type CMVolume struct {
	Name        string      `structs:"name"`
	DefaultMode int64       `structs:"defaultMode"`
	CMName      string      `structs:"cmName"`
	Items       []KeyToPath `structs:"items"`
}

// SecretVolume ...
type SecretVolume struct {
	Name        string      `structs:"name"`
	DefaultMode int64       `structs:"defaultMode"`
	SecretName  string      `structs:"secretName"`
	Items       []KeyToPath `structs:"items"`
}

// KeyToPath ...
type KeyToPath struct {
	Key  string `structs:"key"`
	Path string `structs:"path"`
}

// EmptyDirVolume ...
type EmptyDirVolume struct {
	Name string `structs:"name"`
}

// NFSVolume ...
type NFSVolume struct {
	Name     string `structs:"name"`
	Path     string `structs:"path"`
	Server   string `structs:"server"`
	ReadOnly bool   `structs:"readOnly"`
}

// ContainerGroup ...
type ContainerGroup struct {
	InitContainers []Container `structs:"initContainers"`
	Containers     []Container `structs:"containers"`
}
