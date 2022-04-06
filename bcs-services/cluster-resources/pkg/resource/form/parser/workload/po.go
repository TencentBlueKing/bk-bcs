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
func ParseNodeSelect(manifest map[string]interface{}, nodeSelect *model.NodeSelect) {
	nodeSelect.Type = NodeSelectTypeAnyAvailable
	nodeSelector, _ := mapx.GetItems(manifest, "spec.template.spec.nodeSelector")
	if nodeSelector != nil {
		nodeSelect.Type = NodeSelectTypeSchedulingRule
		for k, v := range nodeSelector.(map[string]interface{}) {
			nodeSelect.Selector = append(nodeSelect.Selector, model.NodeSelector{Key: k, Value: v.(string)})
		}
	}
	nodeName, _ := mapx.GetItems(manifest, "spec.template.spec.nodeName")
	if nodeName != nil {
		nodeSelect.Type = NodeSelectTypeSpecificNode
		nodeSelect.NodeName = nodeName.(string)
	}
}

// ParseAffinity ...
func ParseAffinity(manifest map[string]interface{}, affinity *model.Affinity) {
	ParseNodeAffinity(manifest, &affinity.NodeAffinity)
	ParsePodAffinity(manifest, &affinity.PodAffinity)
}

// ParseNodeAffinity ...
func ParseNodeAffinity(manifest map[string]interface{}, nodeAffinity *[]model.NodeAffinity) {
	if affinity, _ := mapx.GetItems(manifest, "spec.template.spec.affinity.nodeAffinity"); affinity != nil { // nolint:nestif
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
func ParsePodAffinity(manifest map[string]interface{}, podAffinity *[]model.PodAffinity) {
	for _, typeArgs := range [][]string{
		{AffinityTypeAffinity, "spec.template.spec.affinity.podAffinity"},
		{AffinityTypeAntiAffinity, "spec.template.spec.affinity.podAntiAffinity"},
	} {
		for _, priorityArgs := range [][]string{
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
		} {
			affinityType, affinityPaths := typeArgs[0], typeArgs[1]
			priority, execKey, expPaths := priorityArgs[0], priorityArgs[1], priorityArgs[2]
			labelPaths, nsPaths, topoKeyPaths := priorityArgs[3], priorityArgs[4], priorityArgs[5]
			if affinity, _ := mapx.GetItems(manifest, affinityPaths); affinity != nil { // nolint:nestif
				if execs, ok := affinity.(map[string]interface{})[execKey]; ok {
					for _, exec := range execs.([]interface{}) {
						e, _ := exec.(map[string]interface{})
						namespaces := []string{}
						for _, ns := range mapx.Get(e, nsPaths, []interface{}{}).([]interface{}) {
							namespaces = append(namespaces, ns.(string))
						}
						aff := model.PodAffinity{
							Type:        affinityType,
							Priority:    priority,
							Namespaces:  namespaces,
							TopologyKey: mapx.Get(e, topoKeyPaths, "").(string),
						}
						if weight, ok := e["weight"]; ok {
							aff.Weight = weight.(int64)
						}
						if matchExps, _ := mapx.GetItems(e, expPaths); matchExps != nil {
							aff.Selector.Expressions = parseAffinityExpSelector(matchExps)
						}
						if matchLabels, _ := mapx.GetItems(e, labelPaths); matchLabels != nil {
							for k, v := range matchLabels.(map[string]interface{}) {
								aff.Selector.Labels = append(aff.Selector.Labels, model.LabelSelector{
									Key: k, Value: v.(string),
								})
							}
						}
						*podAffinity = append(*podAffinity, aff)
					}
				}
			}
		}
	}
}

// ParseToleration ...
func ParseToleration(manifest map[string]interface{}, toleration *model.Toleration) {
	if tolerations, _ := mapx.GetItems(manifest, "spec.template.spec.tolerations"); tolerations != nil {
		_ = mapstructure.Decode(tolerations, &toleration.Rules)
	}
}

// ParseNetworking ...
func ParseNetworking(manifest map[string]interface{}, networking *model.Networking) {
	if templateSpec, _ := mapx.GetItems(manifest, "spec.template.spec"); templateSpec != nil {
		tmplSpec, _ := templateSpec.(map[string]interface{})
		networking.DNSPolicy = mapx.Get(tmplSpec, "dnsPolicy", "ClusterFirst").(string)
		networking.HostIPC = mapx.Get(tmplSpec, "hostIPC", false).(bool)
		networking.HostNetwork = mapx.Get(tmplSpec, "hostNetwork", false).(bool)
		networking.HostPID = mapx.Get(tmplSpec, "hostPID", false).(bool)
		networking.ShareProcessNamespace = mapx.Get(tmplSpec, "shareProcessNamespace", false).(bool)
		networking.HostName = mapx.Get(tmplSpec, "hostname", "").(string)
		networking.Subdomain = mapx.Get(tmplSpec, "subdomain", "").(string)
		for _, ns := range mapx.Get(tmplSpec, "dnsConfig.nameservers", []interface{}{}).([]interface{}) {
			networking.NameServers = append(networking.NameServers, ns.(string))
		}
		for _, s := range mapx.Get(tmplSpec, "dnsConfig.searches", []interface{}{}).([]interface{}) {
			networking.Searches = append(networking.Searches, s.(string))
		}
		if dnsOpts, _ := mapx.GetItems(tmplSpec, "dnsConfig.options"); dnsOpts != nil {
			for _, opt := range dnsOpts.([]interface{}) {
				networking.DNSResolverOpts = append(networking.DNSResolverOpts, model.DNSResolverOpt{
					Name:  opt.(map[string]interface{})["name"].(string),
					Value: opt.(map[string]interface{})["value"].(string),
				})
			}
		}
		if hostAliases, _ := mapx.GetItems(tmplSpec, "hostAliases"); hostAliases != nil {
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
}

// ParsePodSecurityCtx ...
func ParsePodSecurityCtx(manifest map[string]interface{}, security *model.PodSecurityCtx) {
	if secCtx, _ := mapx.GetItems(manifest, "spec.template.spec.securityContext"); secCtx != nil {
		_ = mapstructure.Decode(secCtx, security)
	}
}

// ParseSpecOther ...
func ParseSpecOther(manifest map[string]interface{}, other *model.SpecOther) {
	if templateSpec, _ := mapx.GetItems(manifest, "spec.template.spec"); templateSpec != nil {
		tmplSpec, _ := templateSpec.(map[string]interface{})
		other.RestartPolicy = mapx.Get(tmplSpec, "restartPolicy", "Always").(string)
		other.TerminationGracePeriodSecs = mapx.Get(tmplSpec, "terminationGracePeriodSeconds", int64(0)).(int64)
		other.SAName = mapx.Get(tmplSpec, "serviceAccountName", "").(string)
		if imagePullSecrets, ok := tmplSpec["imagePullSecrets"]; ok {
			for _, secret := range imagePullSecrets.([]interface{}) {
				other.ImagePullSecrets = append(other.ImagePullSecrets, secret.(map[string]interface{})["name"].(string))
			}
		}
	}
}
