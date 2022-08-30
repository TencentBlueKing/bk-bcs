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

// NodeSelect xxx
type NodeSelect struct {
	Type     string         `structs:"type"`
	NodeName string         `structs:"nodeName"`
	Selector []NodeSelector `structs:"selector"`
}

// NodeSelector xxx
type NodeSelector struct {
	Key   string `structs:"key"`
	Value string `structs:"value"`
}

// Affinity xxx
type Affinity struct {
	NodeAffinity []NodeAffinity `structs:"nodeAffinity"`
	PodAffinity  []PodAffinity  `structs:"podAffinity"`
}

// NodeAffinity xxx
type NodeAffinity struct {
	Priority string               `structs:"priority"`
	Weight   int64                `structs:"weight"`
	Selector NodeAffinitySelector `structs:"selector"`
}

// NodeAffinitySelector xxx
type NodeAffinitySelector struct {
	Expressions []ExpSelector   `structs:"expressions"`
	Fields      []FieldSelector `structs:"fields"`
}

// ExpSelector xxx
type ExpSelector struct {
	Key    string `structs:"key"`
	Op     string `structs:"op"`
	Values string `structs:"values"`
}

// FieldSelector xxx
type FieldSelector struct {
	Key    string `structs:"key"`
	Op     string `structs:"op"`
	Values string `structs:"values"`
}

// PodAffinity xxx
type PodAffinity struct {
	Type        string              `structs:"type"`
	Priority    string              `structs:"priority"`
	Namespaces  []string            `structs:"namespaces"`
	Weight      int64               `structs:"weight"`
	TopologyKey string              `structs:"topologyKey"`
	Selector    PodAffinitySelector `structs:"selector"`
}

// PodAffinitySelector xxx
type PodAffinitySelector struct {
	Expressions []ExpSelector   `structs:"expressions"`
	Labels      []LabelSelector `structs:"labels"`
}

// LabelSelector xxx
type LabelSelector struct {
	Key   string `structs:"key"`
	Value string `structs:"value"`
}

// Toleration xxx
type Toleration struct {
	Rules []TolerationRule `structs:"rules"`
}

// TolerationRule xxx
type TolerationRule struct {
	Key            string `structs:"key"`
	Op             string `structs:"op" mapstructure:"operator"`
	Value          string `structs:"value"`
	Effect         string `structs:"effect"`
	TolerationSecs int64  `structs:"tolerationSecs" mapstructure:"tolerationSeconds"`
}

// Networking xxx
type Networking struct {
	DNSPolicy             string           `structs:"dnsPolicy"`
	HostIPC               bool             `structs:"hostIPC"`
	HostNetwork           bool             `structs:"hostNetwork"`
	HostPID               bool             `structs:"hostPID"`
	ShareProcessNamespace bool             `structs:"shareProcessNamespace"`
	Hostname              string           `structs:"hostname"`
	Subdomain             string           `structs:"subdomain"`
	NameServers           []string         `structs:"nameServers"`
	Searches              []string         `structs:"searches"`
	DNSResolverOpts       []DNSResolverOpt `structs:"dnsResolverOpts"`
	HostAliases           []HostAlias      `structs:"hostAliases"`
}

// DNSResolverOpt xxx
type DNSResolverOpt struct {
	Name  string `structs:"name"`
	Value string `structs:"value"`
}

// HostAlias xxx
type HostAlias struct {
	IP    string `structs:"ip"`
	Alias string `structs:"alias"`
}

// PodSecurityCtx xxx
type PodSecurityCtx struct {
	RunAsUser    int64      `structs:"runAsUser"`
	RunAsNonRoot bool       `structs:"runAsNonRoot"`
	RunAsGroup   int64      `structs:"runAsGroup"`
	FSGroup      int64      `structs:"fsGroup"`
	SELinuxOpt   SELinuxOpt `structs:"seLinuxOpt" mapstructure:"seLinuxOptions"`
}

// SELinuxOpt xxx
type SELinuxOpt struct {
	Level string `structs:"level"`
	Role  string `structs:"role"`
	Type  string `structs:"type"`
	User  string `structs:"user"`
}

// SpecOther xxx
type SpecOther struct {
	RestartPolicy              string   `structs:"restartPolicy"`              // 重启策略，其中 CJ，Job 没有 Always
	TerminationGracePeriodSecs int64    `structs:"terminationGracePeriodSecs"` // 终止容忍期
	ImagePullSecrets           []string `structs:"imagePullSecrets"`
	SAName                     string   `structs:"saName"`
}
