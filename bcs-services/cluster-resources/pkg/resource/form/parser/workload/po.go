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

package workload

import (
	"strings"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParsePo xxx
func ParsePo(manifest map[string]interface{}) map[string]interface{} {
	po := model.Po{}
	common.ParseMetadata(manifest, &po.Metadata)
	ParsePoSpec(manifest, &po.Spec)
	ParseWorkloadVolume(manifest, &po.Volume)
	ParseContainerGroup(manifest, &po.ContainerGroup)
	return structs.Map(po)
}

// ParsePoSpec xxx
func ParsePoSpec(manifest map[string]interface{}, spec *model.PoSpec) {
	tmplSpec, _ := mapx.GetItems(manifest, "spec")
	if podSpec, ok := tmplSpec.(map[string]interface{}); ok {
		ParseNodeSelect(podSpec, &spec.NodeSelect)
		ParseAffinity(podSpec, &spec.Affinity)
		ParseToleration(podSpec, &spec.Toleration)
		ParseNetworking(podSpec, &spec.Networking)
		ParsePodSecurityCtx(podSpec, &spec.Security)
		ParseSpecReadinessGates(podSpec, &spec.ReadinessGates)
		ParseSpecOther(podSpec, &spec.Other)
	}
}

// ParseNodeSelect xxx
// 类型优先级：指定节点 > 调度规则 > 任意节点
func ParseNodeSelect(podSpec map[string]interface{}, nodeSelect *model.NodeSelect) {
	nodeSelect.Type = resCsts.NodeSelectTypeAnyAvailable
	nodeSelector := mapx.GetMap(podSpec, "nodeSelector")
	if nodeSelector != nil {
		nodeSelect.Type = resCsts.NodeSelectTypeSchedulingRule
		for k := range nodeSelector {
			nodeSelect.Selector = append(nodeSelect.Selector,
				model.NodeSelector{Key: k, Value: mapx.GetStr(nodeSelector, k)})
		}
	}
	nodeName, _ := mapx.GetItems(podSpec, "nodeName")
	if nodeName != nil {
		nodeSelect.Type = resCsts.NodeSelectTypeSpecificNode
		if nodeName, ok := nodeName.(string); ok {
			nodeSelect.NodeName = nodeName
		}
	}
}

// ParseAffinity xxx
func ParseAffinity(podSpec map[string]interface{}, affinity *model.Affinity) {
	ParseNodeAffinity(podSpec, &affinity.NodeAffinity)
	ParsePodAffinity(podSpec, &affinity.PodAffinity)
}

// ParseNodeAffinity xxx
func ParseNodeAffinity(manifest map[string]interface{}, nodeAffinity *[]model.NodeAffinity) {
	affinity := mapx.GetMap(manifest, "affinity.nodeAffinity")
	for _, term := range mapx.GetList(
		affinity, "requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms",
	) {
		if t, ok := term.(map[string]interface{}); ok {
			aff := model.NodeAffinity{Priority: resCsts.AffinityPriorityRequired}
			aff.Selector.Expressions = parseAffinityExpSelector(mapx.GetList(t, "matchExpressions"))
			aff.Selector.Fields = parseAffinityFieldSelector(mapx.GetList(t, "matchFields"))
			*nodeAffinity = append(*nodeAffinity, aff)
		}
	}
	for _, exec := range mapx.GetList(affinity, "preferredDuringSchedulingIgnoredDuringExecution") {
		if e, ok := exec.(map[string]interface{}); ok {
			aff := model.NodeAffinity{Priority: resCsts.AffinityPriorityPreferred}
			aff.Weight = mapx.GetInt64(e, "weight")
			aff.Selector.Expressions = parseAffinityExpSelector(mapx.GetList(e, "preference.matchExpressions"))
			aff.Selector.Fields = parseAffinityFieldSelector(mapx.GetList(e, "preference.matchFields"))
			*nodeAffinity = append(*nodeAffinity, aff)
		}
	}
}

func parseAffinityExpSelector(matchExps []interface{}) []model.ExpSelector {
	selectors := []model.ExpSelector{}
	for _, exps := range matchExps {
		if es, ok := exps.(map[string]interface{}); ok {
			values := []string{}
			for _, v := range mapx.GetList(es, "values") {
				if v, ok := v.(string); ok {
					values = append(values, v)
				}
			}
			selectors = append(selectors, model.ExpSelector{
				Key: mapx.GetStr(es, "key"), Op: mapx.GetStr(es, "operator"), Values: strings.Join(values, ","),
			})
		}
	}
	return selectors
}

func parseAffinityFieldSelector(matchFields []interface{}) []model.FieldSelector {
	selectors := []model.FieldSelector{}
	for _, fields := range matchFields {
		if fs, ok := fields.(map[string]interface{}); ok {
			values := []string{}
			for _, v := range mapx.GetList(fs, "values") {
				if v, ok := v.(string); ok {
					values = append(values, v)
				}
			}
			selectors = append(selectors, model.FieldSelector{
				Key: mapx.GetStr(fs, "key"), Op: mapx.GetStr(fs, "operator"), Values: strings.Join(values, ","),
			})
		}
	}
	return selectors
}

// ParsePodAffinity xxx
func ParsePodAffinity(podSpec map[string]interface{}, podAffinity *[]model.PodAffinity) {
	typeArgsList := []affinityTypeArgs{
		{resCsts.AffinityTypeAffinity, "affinity.podAffinity"},
		{resCsts.AffinityTypeAntiAffinity, "affinity.podAntiAffinity"},
	}
	priorityArgsList := []affinityPriorityArgs{
		{
			resCsts.AffinityPriorityPreferred,
			"preferredDuringSchedulingIgnoredDuringExecution",
			"podAffinityTerm.labelSelector.matchExpressions",
			"podAffinityTerm.labelSelector.matchLabels",
			"podAffinityTerm.namespaces",
			"podAffinityTerm.topologyKey",
		},
		{
			resCsts.AffinityPriorityRequired,
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

// parsePodAffinity 解析具体的某类 Pod 亲和性配置，如 反亲和性 + 必须
func parsePodAffinity(
	podSpec map[string]interface{},
	podAffinity *[]model.PodAffinity,
	typeArgs affinityTypeArgs,
	priorityArgs affinityPriorityArgs,
) {
	for _, exec := range mapx.GetList(podSpec, typeArgs.Paths+"."+priorityArgs.ExecKey) {
		if e, ok := exec.(map[string]interface{}); ok {

			namespaces := []string{}
			for _, ns := range mapx.GetList(e, priorityArgs.NSPaths) {
				if ns, ok := ns.(string); ok {
					namespaces = append(namespaces, ns)
				}
			}
			aff := model.PodAffinity{
				Type:        typeArgs.Type,
				Priority:    priorityArgs.Priority,
				Namespaces:  namespaces,
				TopologyKey: mapx.GetStr(e, priorityArgs.TopoKeyPaths),
			}
			aff.Weight = mapx.GetInt64(e, "weight")
			aff.Selector.Expressions = parseAffinityExpSelector(mapx.GetList(e, priorityArgs.ExpPaths))
			for k, v := range mapx.GetMap(e, priorityArgs.LabelPaths) {
				if v, ok := v.(string); ok {
					aff.Selector.Labels = append(aff.Selector.Labels, model.LabelSelector{
						Key: k, Value: v,
					})
				}
			}
			*podAffinity = append(*podAffinity, aff)
		}
	}
}

// ParseToleration xxx
func ParseToleration(podSpec map[string]interface{}, toleration *model.Toleration) {
	if tolerations, _ := mapx.GetItems(podSpec, "tolerations"); tolerations != nil {
		_ = mapstructure.Decode(tolerations, &toleration.Rules)
	}
}

// ParseNetworking xxx
func ParseNetworking(podSpec map[string]interface{}, networking *model.Networking) {
	if dnsPolicy, ok := mapx.Get(podSpec, "dnsPolicy", "ClusterFirst").(string); ok {
		networking.DNSPolicy = dnsPolicy
	}
	networking.HostIPC = mapx.GetBool(podSpec, "hostIPC")
	networking.HostNetwork = mapx.GetBool(podSpec, "hostNetwork")
	networking.HostPID = mapx.GetBool(podSpec, "hostPID")
	networking.ShareProcessNamespace = mapx.GetBool(podSpec, "shareProcessNamespace")
	networking.Hostname = mapx.GetStr(podSpec, "hostname")
	networking.Subdomain = mapx.GetStr(podSpec, "subdomain")
	for _, ns := range mapx.GetList(podSpec, "dnsConfig.nameservers") {
		if ns, ok := ns.(string); ok {
			networking.NameServers = append(networking.NameServers, ns)
		}
	}
	for _, s := range mapx.GetList(podSpec, "dnsConfig.searches") {
		if s, ok := s.(string); ok {
			networking.Searches = append(networking.Searches, s)
		}
	}
	for _, opt := range mapx.GetList(podSpec, "dnsConfig.options") {
		if optM, ok := opt.(map[string]interface{}); ok {
			networking.DNSResolverOpts = append(networking.DNSResolverOpts, model.DNSResolverOpt{
				Name:  mapx.GetStr(optM, "name"),
				Value: mapx.GetStr(optM, "value"),
			})
		}
	}
	for _, hostAlias := range mapx.GetList(podSpec, "hostAliases") {
		if aliasM, ok := hostAlias.(map[string]interface{}); ok {
			hostnames := []string{}
			for _, hName := range mapx.GetList(aliasM, "hostnames") {
				if hName, ok := hName.(string); ok {
					hostnames = append(hostnames, hName)
				}
			}
			networking.HostAliases = append(networking.HostAliases, model.HostAlias{
				IP:    mapx.GetStr(aliasM, "ip"),
				Alias: strings.Join(hostnames, ","),
			})
		}
	}
}

// ParsePodSecurityCtx xxx
func ParsePodSecurityCtx(podSpec map[string]interface{}, security *model.PodSecurityCtx) {
	if secCtx, _ := mapx.GetItems(podSpec, "securityContext"); secCtx != nil {
		_ = mapstructure.Decode(secCtx, security)
	}
}

// ParseSpecReadinessGates xxx
func ParseSpecReadinessGates(podSpec map[string]interface{}, rg *model.ReadinessGates) {
	for _, cond := range mapx.GetList(podSpec, "readinessGates") {
		if condM, ok := cond.(map[string]interface{}); ok {
			rg.ReadinessGates = append(rg.ReadinessGates, mapx.GetStr(condM, "conditionType"))
		}
	}
}

// ParseSpecOther xxx
func ParseSpecOther(podSpec map[string]interface{}, other *model.SpecOther) {
	if restartPolicy, ok := mapx.Get(podSpec, "restartPolicy", "Always").(string); ok {
		other.RestartPolicy = restartPolicy
	}
	other.TerminationGracePeriodSecs = mapx.GetInt64(podSpec, "terminationGracePeriodSeconds")
	other.SAName = mapx.GetStr(podSpec, "serviceAccountName")
	for _, secret := range mapx.GetList(podSpec, "imagePullSecrets") {
		if secretM, ok := secret.(map[string]interface{}); ok {
			other.ImagePullSecrets = append(other.ImagePullSecrets, mapx.GetStr(secretM, "name"))
		}
	}
}
