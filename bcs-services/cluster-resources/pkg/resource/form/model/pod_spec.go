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

package model

// NodeSelect 节点选择
type NodeSelect struct {
	Type     string         `structs:"type"`
	NodeName string         `structs:"nodeName"`
	Selector []NodeSelector `structs:"selector"`
}

// NodeSelector 节点选择器
type NodeSelector struct {
	Key   string `structs:"key"`
	Value string `structs:"value"`
}

// Affinity 亲和性配置
type Affinity struct {
	NodeAffinity []NodeAffinity `structs:"nodeAffinity"`
	PodAffinity  []PodAffinity  `structs:"podAffinity"`
}

// NodeAffinity 节点亲和性
type NodeAffinity struct {
	Priority string               `structs:"priority"`
	Weight   int64                `structs:"weight"`
	Selector NodeAffinitySelector `structs:"selector"`
}

// NodeAffinitySelector 节点亲和性选择器
type NodeAffinitySelector struct {
	Expressions []ExpSelector   `structs:"expressions"`
	Fields      []FieldSelector `structs:"fields"`
}

// ExpSelector 表达式选择器
type ExpSelector struct {
	Key    string `structs:"key"`
	Op     string `structs:"op"`
	Values string `structs:"values"`
}

// FieldSelector 字段选择器
type FieldSelector struct {
	Key    string `structs:"key"`
	Op     string `structs:"op"`
	Values string `structs:"values"`
}

// PodAffinity Pod 亲和性
type PodAffinity struct {
	Type        string              `structs:"type"`
	Priority    string              `structs:"priority"`
	Namespaces  []string            `structs:"namespaces"`
	Weight      int64               `structs:"weight"`
	TopologyKey string              `structs:"topologyKey"`
	Selector    PodAffinitySelector `structs:"selector"`
}

// PodAffinitySelector Pod 亲和性选择器
type PodAffinitySelector struct {
	Expressions []ExpSelector   `structs:"expressions"`
	Labels      []LabelSelector `structs:"labels"`
}

// LabelSelector 标签选择器
type LabelSelector struct {
	Key   string `structs:"key"`
	Value string `structs:"value"`
}

// Toleration 容忍性
type Toleration struct {
	Rules []TolerationRule `structs:"rules"`
}

// TolerationRule 容忍规则
type TolerationRule struct {
	Key            string `structs:"key"`
	Op             string `structs:"op" mapstructure:"operator"`
	Value          string `structs:"value"`
	Effect         string `structs:"effect"`
	TolerationSecs int64  `structs:"tolerationSecs" mapstructure:"tolerationSeconds"`
}

// Networking 网络配置
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

// DNSResolverOpt DNS 解析配置
type DNSResolverOpt struct {
	Name  string `structs:"name"`
	Value string `structs:"value"`
}

// HostAlias 主机别名
type HostAlias struct {
	IP    string `structs:"ip"`
	Alias string `structs:"alias"`
}

// PodSecurityCtx Pod 安全性上下文
type PodSecurityCtx struct {
	RunAsUser    int64      `structs:"runAsUser"`
	RunAsNonRoot bool       `structs:"runAsNonRoot"`
	RunAsGroup   int64      `structs:"runAsGroup"`
	FSGroup      int64      `structs:"fsGroup"`
	SELinuxOpt   SELinuxOpt `structs:"seLinuxOpt" mapstructure:"seLinuxOptions"`
}

// SELinuxOpt Linux 安全配置
type SELinuxOpt struct {
	Level string `structs:"level"`
	Role  string `structs:"role"`
	Type  string `structs:"type"`
	User  string `structs:"user"`
}

// ReadinessGates 就绪探针
type ReadinessGates struct {
	ReadinessGates []string `structs:"readinessGates"`
}

// SpecOther 额外配置
type SpecOther struct {
	RestartPolicy              string   `structs:"restartPolicy"`              // 重启策略，其中 CJ，Job 没有 Always
	TerminationGracePeriodSecs int64    `structs:"terminationGracePeriodSecs"` // 终止容忍期
	ImagePullSecrets           []string `structs:"imagePullSecrets"`
	SAName                     string   `structs:"saName"`
}
