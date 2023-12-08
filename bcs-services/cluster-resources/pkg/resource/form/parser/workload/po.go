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
	podSpec, _ := tmplSpec.(map[string]interface{})
	ParseNodeSelect(podSpec, &spec.NodeSelect)
	ParseAffinity(podSpec, &spec.Affinity)
	ParseToleration(podSpec, &spec.Toleration)
	ParseNetworking(podSpec, &spec.Networking)
	ParsePodSecurityCtx(podSpec, &spec.Security)
	ParseSpecReadinessGates(podSpec, &spec.ReadinessGates)
	ParseSpecOther(podSpec, &spec.Other)
}

// ParseNodeSelect xxx
// 类型优先级：指定节点 > 调度规则 > 任意节点
func ParseNodeSelect(podSpec map[string]interface{}, nodeSelect *model.NodeSelect) {
	nodeSelect.Type = resCsts.NodeSelectTypeAnyAvailable
	nodeSelector, _ := mapx.GetItems(podSpec, "nodeSelector")
	if nodeSelector != nil {
		nodeSelect.Type = resCsts.NodeSelectTypeSchedulingRule
		for k, v := range nodeSelector.(map[string]interface{}) {
			nodeSelect.Selector = append(nodeSelect.Selector, model.NodeSelector{Key: k, Value: v.(string)})
		}
	}
	nodeName, _ := mapx.GetItems(podSpec, "nodeName")
	if nodeName != nil {
		nodeSelect.Type = resCsts.NodeSelectTypeSpecificNode
		nodeSelect.NodeName = nodeName.(string)
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
		t, _ := term.(map[string]interface{})
		aff := model.NodeAffinity{Priority: resCsts.AffinityPriorityRequired}
		aff.Selector.Expressions = parseAffinityExpSelector(mapx.GetList(t, "matchExpressions"))
		aff.Selector.Fields = parseAffinityFieldSelector(mapx.GetList(t, "matchFields"))
		*nodeAffinity = append(*nodeAffinity, aff)
	}
	for _, exec := range mapx.GetList(
		affinity, "preferredDuringSchedulingIgnoredDuringExecution",
	) {
		e, _ := exec.(map[string]interface{})
		aff := model.NodeAffinity{Priority: resCsts.AffinityPriorityPreferred}
		if weight, ok := e["weight"]; ok {
			aff.Weight = weight.(int64)
		}
		aff.Selector.Expressions = parseAffinityExpSelector(mapx.GetList(e, "preference.matchExpressions"))
		aff.Selector.Fields = parseAffinityFieldSelector(mapx.GetList(e, "preference.matchFields"))
		*nodeAffinity = append(*nodeAffinity, aff)
	}
}

func parseAffinityExpSelector(matchExps []interface{}) []model.ExpSelector {
	selectors := []model.ExpSelector{}
	for _, exps := range matchExps {
		es, _ := exps.(map[string]interface{})
		values := []string{}
		for _, v := range mapx.GetList(es, "values") {
			values = append(values, v.(string))
		}
		selectors = append(selectors, model.ExpSelector{
			Key: es["key"].(string), Op: es["operator"].(string), Values: strings.Join(values, ","),
		})
	}
	return selectors
}

func parseAffinityFieldSelector(matchFields []interface{}) []model.FieldSelector {
	selectors := []model.FieldSelector{}
	for _, fields := range matchFields {
		fs, _ := fields.(map[string]interface{})
		values := []string{}
		for _, v := range mapx.GetList(fs, "values") {
			values = append(values, v.(string))
		}
		selectors = append(selectors, model.FieldSelector{
			Key: fs["key"].(string), Op: fs["operator"].(string), Values: strings.Join(values, ","),
		})
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
		e, _ := exec.(map[string]interface{})
		namespaces := []string{}
		for _, ns := range mapx.GetList(e, priorityArgs.NSPaths) {
			namespaces = append(namespaces, ns.(string))
		}
		aff := model.PodAffinity{
			Type:        typeArgs.Type,
			Priority:    priorityArgs.Priority,
			Namespaces:  namespaces,
			TopologyKey: mapx.GetStr(e, priorityArgs.TopoKeyPaths),
		}
		if weight, ok := e["weight"]; ok {
			aff.Weight = weight.(int64)
		}
		aff.Selector.Expressions = parseAffinityExpSelector(mapx.GetList(e, priorityArgs.ExpPaths))
		for k, v := range mapx.GetMap(e, priorityArgs.LabelPaths) {
			aff.Selector.Labels = append(aff.Selector.Labels, model.LabelSelector{
				Key: k, Value: v.(string),
			})
		}
		*podAffinity = append(*podAffinity, aff)
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
	networking.DNSPolicy = mapx.Get(podSpec, "dnsPolicy", "ClusterFirst").(string)
	networking.HostIPC = mapx.GetBool(podSpec, "hostIPC")
	networking.HostNetwork = mapx.GetBool(podSpec, "hostNetwork")
	networking.HostPID = mapx.GetBool(podSpec, "hostPID")
	networking.ShareProcessNamespace = mapx.GetBool(podSpec, "shareProcessNamespace")
	networking.Hostname = mapx.GetStr(podSpec, "hostname")
	networking.Subdomain = mapx.GetStr(podSpec, "subdomain")
	for _, ns := range mapx.GetList(podSpec, "dnsConfig.nameservers") {
		networking.NameServers = append(networking.NameServers, ns.(string))
	}
	for _, s := range mapx.GetList(podSpec, "dnsConfig.searches") {
		networking.Searches = append(networking.Searches, s.(string))
	}
	for _, opt := range mapx.GetList(podSpec, "dnsConfig.options") {
		networking.DNSResolverOpts = append(networking.DNSResolverOpts, model.DNSResolverOpt{
			Name:  opt.(map[string]interface{})["name"].(string),
			Value: opt.(map[string]interface{})["value"].(string),
		})
	}
	for _, hostAlias := range mapx.GetList(podSpec, "hostAliases") {
		alias, _ := hostAlias.(map[string]interface{})
		hostnames := []string{}
		for _, hName := range mapx.GetList(alias, "hostnames") {
			hostnames = append(hostnames, hName.(string))
		}
		networking.HostAliases = append(networking.HostAliases, model.HostAlias{
			IP: alias["ip"].(string), Alias: strings.Join(hostnames, ","),
		})
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
		rg.ReadinessGates = append(rg.ReadinessGates, mapx.GetStr(cond.(map[string]interface{}), "conditionType"))
	}
}

// ParseSpecOther xxx
func ParseSpecOther(podSpec map[string]interface{}, other *model.SpecOther) {
	other.RestartPolicy = mapx.Get(podSpec, "restartPolicy", "Always").(string)
	other.TerminationGracePeriodSecs = mapx.GetInt64(podSpec, "terminationGracePeriodSeconds")
	other.SAName = mapx.GetStr(podSpec, "serviceAccountName")
	for _, secret := range mapx.GetList(podSpec, "imagePullSecrets") {
		other.ImagePullSecrets = append(other.ImagePullSecrets, secret.(map[string]interface{})["name"].(string))
	}
}
