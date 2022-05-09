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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightPodSpec = map[string]interface{}{
	"containers": []interface{}{},
	"dnsConfig": map[string]interface{}{
		"nameservers": []interface{}{
			"1.1.1.1",
			"2.2.2.2",
		},
		"options": []interface{}{
			map[string]interface{}{
				"name":  "testName",
				"value": "testValue",
			},
		},
		"searches": []interface{}{
			"3.3.3.3",
		},
	},
	"dnsPolicy": "ClusterFirst",
	"hostIPC":   true,
	"hostAliases": []interface{}{
		map[string]interface{}{
			"hostnames": []interface{}{
				"vm-1",
				"vm-2",
			},
			"ip": "5.5.5.5",
		},
	},
	"hostname": "vm-12345",
	"imagePullSecrets": []interface{}{
		map[string]interface{}{
			"name": "default-token-1",
		},
		map[string]interface{}{
			"name": "default-token-2",
		},
	},
	"nodeName": "vm-123",
	"nodeSelector": map[string]interface{}{
		"kubernetes.io/arch": "amd64",
	},
	"restartPolicy": "Always",
	"schedulerName": "default-scheduler",
	"securityContext": map[string]interface{}{
		"fsGroup":      int64(3333),
		"runAsGroup":   int64(2222),
		"runAsNonRoot": true,
		"runAsUser":    int64(1111),
		"seLinuxOptions": map[string]interface{}{
			"level": "4444",
			"role":  "5555",
			"type":  "6666",
			"user":  "7777",
		},
	},
	"serviceAccount":                "default",
	"serviceAccountName":            "default",
	"subdomain":                     "blueking",
	"terminationGracePeriodSeconds": int64(30),
	"tolerations": []interface{}{
		map[string]interface{}{
			"effect":   "PreferNoSchedule",
			"key":      "testTolKey1",
			"operator": "Exists",
		},
		map[string]interface{}{
			"effect":            "NoExecute",
			"key":               "testTolKey2",
			"operator":          "Equal",
			"tolerationSeconds": int64(120),
			"value":             "tolVal",
		},
	},
	"affinity": map[string]interface{}{
		"podAffinity": map[string]interface{}{
			"preferredDuringSchedulingIgnoredDuringExecution": []interface{}{
				map[string]interface{}{
					"podAffinityTerm": map[string]interface{}{
						"namespaces": []interface{}{
							"kube-system",
							"default",
						},
						"topologyKey": "topoKeyTest1",
						"labelSelector": map[string]interface{}{
							"matchExpressions": []interface{}{
								map[string]interface{}{
									"key":      "testKey",
									"operator": "Equal",
									"values": []interface{}{
										"testVal",
									},
								},
							},
							"matchLabels": map[string]interface{}{
								"labelKey": "labelVal",
							},
						},
					},
					"weight": int64(30),
				},
			},
			"requiredDuringSchedulingIgnoredDuringExecution": []interface{}{
				map[string]interface{}{
					"namespaces": []interface{}{
						"kube-node-lease",
						"default",
					},
					"topologyKey": "topoKeyTest0",
					"labelSelector": map[string]interface{}{
						"matchExpressions": []interface{}{
							map[string]interface{}{
								"key":      "testKey0",
								"operator": "In",
								"values": []interface{}{
									"testVal0",
									"testVal1",
								},
							},
						},
						"matchLabels": map[string]interface{}{
							"labelKey1": "labelVal1",
						},
					},
				},
			},
		},
		"podAntiAffinity": map[string]interface{}{
			"preferredDuringSchedulingIgnoredDuringExecution": []interface{}{
				map[string]interface{}{
					"podAffinityTerm": map[string]interface{}{
						"namespaces": []interface{}{
							"default",
							"kube-system",
						},
						"topologyKey": "topoKeyTest2",
						"labelSelector": map[string]interface{}{
							"matchExpressions": []interface{}{
								map[string]interface{}{
									"key":      "testKey2",
									"operator": "In",
									"values": []interface{}{
										"testVal2",
										"testVal2",
									},
								},
							},
							"matchLabels": map[string]interface{}{
								"testKey3": "testVal3",
							},
						},
					},
					"weight": int64(50),
				},
			},
			"requiredDuringSchedulingIgnoredDuringExecution": []interface{}{
				map[string]interface{}{
					"namespaces": []interface{}{
						"default",
					},
					"topologyKey": "topoKeyTest3",
					"labelSelector": map[string]interface{}{
						"matchExpressions": []interface{}{
							map[string]interface{}{
								"key":      "testKey3",
								"operator": "In",
								"values": []interface{}{
									"testVal3",
									"testVal4",
								},
							},
						},
						"matchLabels": map[string]interface{}{
							"testKey4": "testVal4",
						},
					},
				},
			},
		},
		"nodeAffinity": map[string]interface{}{
			"requiredDuringSchedulingIgnoredDuringExecution": map[string]interface{}{
				"nodeSelectorTerms": []interface{}{
					map[string]interface{}{
						"matchExpressions": []interface{}{
							map[string]interface{}{
								"key":      "testKey",
								"operator": "In",
								"values": []interface{}{
									"testValue1",
								},
							},
						},
						"matchFields": []interface{}{
							map[string]interface{}{
								"key":      "metadata.name",
								"operator": "In",
								"values": []interface{}{
									"testName",
								},
							},
						},
					},
				},
			},
			"preferredDuringSchedulingIgnoredDuringExecution": []interface{}{
				map[string]interface{}{
					"weight": int64(10),
					"preference": map[string]interface{}{
						"matchExpressions": []interface{}{
							map[string]interface{}{
								"key":      "testKey",
								"operator": "In",
								"values": []interface{}{
									"testVal1",
									"testVal2",
									"testVal3",
								},
							},
						},
						"matchFields": []interface{}{
							map[string]interface{}{
								"key":      "metadata.namespace",
								"operator": "In",
								"values": []interface{}{
									"testName1",
								},
							},
						},
					},
				},
			},
		},
	},
}

var exceptedSelect = model.NodeSelect{
	Type:     NodeSelectTypeSpecificNode,
	NodeName: "vm-123",
	Selector: []model.NodeSelector{
		{Key: "kubernetes.io/arch", Value: "amd64"},
	},
}

func TestParseNodeSelect(t *testing.T) {
	actualSelect := model.NodeSelect{}
	ParseNodeSelect(lightPodSpec, &actualSelect)
	assert.Equal(t, exceptedSelect, actualSelect)
}

var exceptedAffinity = model.Affinity{
	NodeAffinity: []model.NodeAffinity{
		{
			Priority: AffinityPriorityRequired,
			Selector: model.NodeAffinitySelector{
				Expressions: []model.ExpSelector{
					{Key: "testKey", Op: "In", Values: "testValue1"},
				},
				Fields: []model.FieldSelector{
					{Key: "metadata.name", Op: "In", Values: "testName"},
				},
			},
		},
		{
			Priority: AffinityPriorityPreferred,
			Weight:   10,
			Selector: model.NodeAffinitySelector{
				Expressions: []model.ExpSelector{
					{Key: "testKey", Op: "In", Values: "testVal1,testVal2,testVal3"},
				},
				Fields: []model.FieldSelector{
					{Key: "metadata.namespace", Op: "In", Values: "testName1"},
				},
			},
		},
	},
	PodAffinity: []model.PodAffinity{
		{
			Type:     AffinityTypeAffinity,
			Priority: AffinityPriorityPreferred,
			Namespaces: []string{
				"kube-system",
				"default",
			},
			Weight:      30,
			TopologyKey: "topoKeyTest1",
			Selector: model.PodAffinitySelector{
				Expressions: []model.ExpSelector{
					{Key: "testKey", Op: "Equal", Values: "testVal"},
				},
				Labels: []model.LabelSelector{
					{Key: "labelKey", Value: "labelVal"},
				},
			},
		},
		{
			Type:     AffinityTypeAffinity,
			Priority: AffinityPriorityRequired,
			Namespaces: []string{
				"kube-node-lease",
				"default",
			},
			TopologyKey: "topoKeyTest0",
			Selector: model.PodAffinitySelector{
				Expressions: []model.ExpSelector{
					{Key: "testKey0", Op: "In", Values: "testVal0,testVal1"},
				},
				Labels: []model.LabelSelector{
					{Key: "labelKey1", Value: "labelVal1"},
				},
			},
		},
		{
			Type:     AffinityTypeAntiAffinity,
			Priority: AffinityPriorityPreferred,
			Namespaces: []string{
				"default",
				"kube-system",
			},
			Weight:      50,
			TopologyKey: "topoKeyTest2",
			Selector: model.PodAffinitySelector{
				Expressions: []model.ExpSelector{
					{Key: "testKey2", Op: "In", Values: "testVal2,testVal2"},
				},
				Labels: []model.LabelSelector{
					{Key: "testKey3", Value: "testVal3"},
				},
			},
		},
		{
			Type:     AffinityTypeAntiAffinity,
			Priority: AffinityPriorityRequired,
			Namespaces: []string{
				"default",
			},
			TopologyKey: "topoKeyTest3",
			Selector: model.PodAffinitySelector{
				Expressions: []model.ExpSelector{
					{Key: "testKey3", Op: "In", Values: "testVal3,testVal4"},
				},
				Labels: []model.LabelSelector{
					{Key: "testKey4", Value: "testVal4"},
				},
			},
		},
	},
}

func TestParseAffinity(t *testing.T) {
	affinity := model.Affinity{}
	ParseAffinity(lightPodSpec, &affinity)
	assert.Equal(t, exceptedAffinity, affinity)
}

var exceptedToleration = model.Toleration{
	Rules: []model.TolerationRule{
		{
			Key:    "testTolKey1",
			Op:     "Exists",
			Effect: "PreferNoSchedule",
		},
		{
			Key:            "testTolKey2",
			Op:             "Equal",
			Effect:         "NoExecute",
			Value:          "tolVal",
			TolerationSecs: 120,
		},
	},
}

func TestParseToleration(t *testing.T) {
	toleration := model.Toleration{}
	ParseToleration(lightPodSpec, &toleration)
	assert.Equal(t, exceptedToleration, toleration)
}

var exceptedNetworking = model.Networking{
	DNSPolicy:             "ClusterFirst",
	HostIPC:               true,
	HostNetwork:           false,
	HostPID:               false,
	ShareProcessNamespace: false,
	HostName:              "vm-12345",
	Subdomain:             "blueking",
	NameServers: []string{
		"1.1.1.1",
		"2.2.2.2",
	},
	Searches: []string{
		"3.3.3.3",
	},
	DNSResolverOpts: []model.DNSResolverOpt{
		{Name: "testName", Value: "testValue"},
	},
	HostAliases: []model.HostAlias{
		{IP: "5.5.5.5", Alias: "vm-1,vm-2"},
	},
}

func TestParseNetworking(t *testing.T) {
	networking := model.Networking{}
	ParseNetworking(lightPodSpec, &networking)
	assert.Equal(t, exceptedNetworking, networking)
}

var exceptedPodSecCtx = model.PodSecurityCtx{
	RunAsUser:    1111,
	RunAsNonRoot: true,
	RunAsGroup:   2222,
	FSGroup:      3333,
	SELinuxOpt: model.SELinuxOpt{
		Level: "4444",
		Role:  "5555",
		Type:  "6666",
		User:  "7777",
	},
}

func TestParsePodSecurityCtx(t *testing.T) {
	podSecCtx := model.PodSecurityCtx{}
	ParsePodSecurityCtx(lightPodSpec, &podSecCtx)
	assert.Equal(t, exceptedPodSecCtx, podSecCtx)
}

var exceptedSpecOther = model.SpecOther{
	RestartPolicy:              "Always",
	TerminationGracePeriodSecs: 30,
	ImagePullSecrets: []string{
		"default-token-1",
		"default-token-2",
	},
	SAName: "default",
}

func TestParseSpecOther(t *testing.T) {
	specOther := model.SpecOther{}
	ParseSpecOther(lightPodSpec, &specOther)
	assert.Equal(t, exceptedSpecOther, specOther)
}
