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

package workload

import (
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParsePo ...
func ParsePo(manifest map[string]interface{}) map[string]interface{} {
	// TODO Pod 解析逻辑
	return map[string]interface{}{}
}

// ParseNodeSelect ...
// 类型优先级：指定节点 > 调度规则 > 任意节点
func ParseNodeSelect(podSpec map[string]interface{}, nodeSelect *model.NodeSelect) {
	nodeSelect.Type = NodeSelectTypeAnyAvailable
	nodeSelector, _ := mapx.GetItems(podSpec, "nodeSelector")
	if nodeSelector != nil {
		nodeSelect.Type = NodeSelectTypeSchedulingRule
		for k, v := range nodeSelector.(map[string]interface{}) {
			nodeSelect.Selector = append(nodeSelect.Selector, model.NodeSelector{Key: k, Value: v.(string)})
		}
	}
	nodeName, _ := mapx.GetItems(podSpec, "nodeName")
	if nodeName != nil {
		nodeSelect.Type = NodeSelectTypeSpecificNode
		nodeSelect.NodeName = nodeName.(string)
	}
}

// ParseAffinity ...
func ParseAffinity(podSpec map[string]interface{}, affinity *model.Affinity) {
	ParseNodeAffinity(podSpec, &affinity.NodeAffinity)
	ParsePodAffinity(podSpec, &affinity.PodAffinity)
}

// ParseNodeAffinity ...
func ParseNodeAffinity(manifest map[string]interface{}, nodeAffinity *[]model.NodeAffinity) {
	if affinity, _ := mapx.GetItems(manifest, "affinity.nodeAffinity"); affinity != nil { // nolint:nestif
		if terms, _ := mapx.GetItems(
			affinity.(map[string]interface{}), "requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms",
		); terms != nil {
			for _, term := range terms.([]interface{}) {
				aff := model.NodeAffinity{Priority: AffinityPriorityRequired}
				if matchExps, ok := term.(map[string]interface{})["matchExpressions"]; ok {
					aff.Selector.Expressions = parseAffinityExpSelector(matchExps)
				}
				if matchFields, ok := term.(map[string]interface{})["matchFields"]; ok {
					aff.Selector.Fields = parseAffinityFieldSelector(matchFields)
				}
				*nodeAffinity = append(*nodeAffinity, aff)
			}
		}
		if execs, ok := affinity.(map[string]interface{})["preferredDuringSchedulingIgnoredDuringExecution"]; ok {
			for _, exec := range execs.([]interface{}) {
				e, _ := exec.(map[string]interface{})
				aff := model.NodeAffinity{Priority: AffinityPriorityPreferred}
				if weight, ok := e["weight"]; ok {
					aff.Weight = weight.(int64)
				}
				if matchExps, _ := mapx.GetItems(e, "preference.matchExpressions"); matchExps != nil {
					aff.Selector.Expressions = parseAffinityExpSelector(matchExps)
				}
				if matchFields, _ := mapx.GetItems(e, "preference.matchFields"); matchFields != nil {
					aff.Selector.Fields = parseAffinityFieldSelector(matchFields)
				}
				*nodeAffinity = append(*nodeAffinity, aff)
			}
		}
	}
}

func parseAffinityExpSelector(matchExps interface{}) []model.ExpSelector {
	selectors := []model.ExpSelector{}
	for _, exps := range matchExps.([]interface{}) {
		es, _ := exps.(map[string]interface{})
		values := []string{}
		for _, v := range es["values"].([]interface{}) {
			values = append(values, v.(string))
		}
		selectors = append(selectors, model.ExpSelector{
			Key: es["key"].(string), Op: es["operator"].(string), Values: strings.Join(values, ","),
		})
	}
	return selectors
}

func parseAffinityFieldSelector(matchFields interface{}) []model.FieldSelector {
	selectors := []model.FieldSelector{}
	for _, fields := range matchFields.([]interface{}) {
		fs, _ := fields.(map[string]interface{})
		values := []string{}
		for _, v := range fs["values"].([]interface{}) {
			values = append(values, v.(string))
		}
		selectors = append(selectors, model.FieldSelector{
			Key: fs["key"].(string), Op: fs["operator"].(string), Values: strings.Join(values, ","),
		})
	}
	return selectors
}

// ParsePodAffinity ...
func ParsePodAffinity(podSpec map[string]interface{}, podAffinity *[]model.PodAffinity) {
	typeArgsList := []affinityTypeArgs{
		{AffinityTypeAffinity, "affinity.podAffinity"},
		{AffinityTypeAntiAffinity, "affinity.podAntiAffinity"},
	}
	priorityArgsList := []affinityPriorityArgs{
		{
			AffinityPriorityPreferred,
			"preferredDuringSchedulingIgnoredDuringExecution",
			"podAffinityTerm.labelSelector.matchExpressions",
			"podAffinityTerm.labelSelector.matchLabels",
			"podAffinityTerm.namespaces",
			"podAffinityTerm.topologyKey",
		},
		{
			AffinityPriorityRequired,
			"requiredDuringSchedulingIgnoredDuringExecution",
			"labelSelector.matchExpressions",
			"labelSelector.matchLabels",
			"namespaces",
			"topologyKey",
		},
	}
	for _, typeArgs := range typeArgsList {
		for _, priorityArgs := range priorityArgsList {
			parsePodAffinity(podSpec, podAffinity, typeArgs, priorityArgs)
		}
	}
}

type affinityTypeArgs struct {
	Type  string
	Paths string
}

type affinityPriorityArgs struct {
	Priority     string
	ExecKey      string
	ExpPaths     string
	LabelPaths   string
	NSPaths      string
	TopoKeyPaths string
}

// 解析具体的某类 Pod 亲和性配置，如 反亲和性 + 必须
func parsePodAffinity(
	podSpec map[string]interface{},
	podAffinity *[]model.PodAffinity,
	typeArgs affinityTypeArgs,
	priorityArgs affinityPriorityArgs,
) {
	affinity, _ := mapx.GetItems(podSpec, typeArgs.Paths)
	if affinity == nil {
		return
	}
	execs, ok := affinity.(map[string]interface{})[priorityArgs.ExecKey]
	if !ok {
		return
	}
	for _, exec := range execs.([]interface{}) {
		e, _ := exec.(map[string]interface{})
		namespaces := []string{}
		for _, ns := range mapx.Get(e, priorityArgs.NSPaths, []interface{}{}).([]interface{}) {
			namespaces = append(namespaces, ns.(string))
		}
		aff := model.PodAffinity{
			Type:        typeArgs.Type,
			Priority:    priorityArgs.Priority,
			Namespaces:  namespaces,
			TopologyKey: mapx.Get(e, priorityArgs.TopoKeyPaths, "").(string),
		}
		if weight, ok := e["weight"]; ok {
			aff.Weight = weight.(int64)
		}
		if matchExps, _ := mapx.GetItems(e, priorityArgs.ExpPaths); matchExps != nil {
			aff.Selector.Expressions = parseAffinityExpSelector(matchExps)
		}
		if matchLabels, _ := mapx.GetItems(e, priorityArgs.LabelPaths); matchLabels != nil {
			for k, v := range matchLabels.(map[string]interface{}) {
				aff.Selector.Labels = append(aff.Selector.Labels, model.LabelSelector{
					Key: k, Value: v.(string),
				})
			}
		}
		*podAffinity = append(*podAffinity, aff)
	}
}

// ParseToleration ...
func ParseToleration(podSpec map[string]interface{}, toleration *model.Toleration) {
	if tolerations, _ := mapx.GetItems(podSpec, "tolerations"); tolerations != nil {
		_ = mapstructure.Decode(tolerations, &toleration.Rules)
	}
}

// ParseNetworking ...
func ParseNetworking(podSpec map[string]interface{}, networking *model.Networking) {
	networking.DNSPolicy = mapx.Get(podSpec, "dnsPolicy", "ClusterFirst").(string)
	networking.HostIPC = mapx.Get(podSpec, "hostIPC", false).(bool)
	networking.HostNetwork = mapx.Get(podSpec, "hostNetwork", false).(bool)
	networking.HostPID = mapx.Get(podSpec, "hostPID", false).(bool)
	networking.ShareProcessNamespace = mapx.Get(podSpec, "shareProcessNamespace", false).(bool)
	networking.HostName = mapx.Get(podSpec, "hostname", "").(string)
	networking.Subdomain = mapx.Get(podSpec, "subdomain", "").(string)
	for _, ns := range mapx.Get(podSpec, "dnsConfig.nameservers", []interface{}{}).([]interface{}) {
		networking.NameServers = append(networking.NameServers, ns.(string))
	}
	for _, s := range mapx.Get(podSpec, "dnsConfig.searches", []interface{}{}).([]interface{}) {
		networking.Searches = append(networking.Searches, s.(string))
	}
	if dnsOpts, _ := mapx.GetItems(podSpec, "dnsConfig.options"); dnsOpts != nil {
		for _, opt := range dnsOpts.([]interface{}) {
			networking.DNSResolverOpts = append(networking.DNSResolverOpts, model.DNSResolverOpt{
				Name:  opt.(map[string]interface{})["name"].(string),
				Value: opt.(map[string]interface{})["value"].(string),
			})
		}
	}
	if hostAliases, _ := mapx.GetItems(podSpec, "hostAliases"); hostAliases != nil {
		for _, hostAlias := range hostAliases.([]interface{}) {
			alias, _ := hostAlias.(map[string]interface{})
			hostnames := []string{}
			for _, hName := range alias["hostnames"].([]interface{}) {
				hostnames = append(hostnames, hName.(string))
			}
			networking.HostAliases = append(networking.HostAliases, model.HostAlias{
				IP: alias["ip"].(string), Alias: strings.Join(hostnames, ","),
			})
		}
	}
}

// ParsePodSecurityCtx ...
func ParsePodSecurityCtx(podSpec map[string]interface{}, security *model.PodSecurityCtx) {
	if secCtx, _ := mapx.GetItems(podSpec, "securityContext"); secCtx != nil {
		_ = mapstructure.Decode(secCtx, security)
	}
}

// ParseSpecOther ...
func ParseSpecOther(podSpec map[string]interface{}, other *model.SpecOther) {
	other.RestartPolicy = mapx.Get(podSpec, "restartPolicy", "Always").(string)
	other.TerminationGracePeriodSecs = mapx.Get(podSpec, "terminationGracePeriodSeconds", int64(0)).(int64)
	other.SAName = mapx.Get(podSpec, "serviceAccountName", "").(string)
	if imagePullSecrets, ok := podSpec["imagePullSecrets"]; ok {
		for _, secret := range imagePullSecrets.([]interface{}) {
			other.ImagePullSecrets = append(other.ImagePullSecrets, secret.(map[string]interface{})["name"].(string))
		}
	}
}
