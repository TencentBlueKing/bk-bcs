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

package formdata

import (
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// IngV1 ...
var IngV1 = model.Ing{
	Metadata: model.Metadata{
		APIVersion: "networking.k8s.io/v1",
		Kind:       resCsts.Ing,
		Name:       "ing-v1-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
		},
		Annotations: []model.Annotation{
			{"anno-key-1", "anno-val-1"},
		},
	},
	Controller: model.IngController{
		Type: resCsts.IngClsQCloud,
	},
	Spec: model.IngSpec{
		RuleConf: model.IngRuleConf{
			Rules: []model.IngRule{
				{
					Domain: "example1.com",
					Paths: []model.IngPath{
						{
							Type:      "Prefix",
							Path:      "/api",
							TargetSVC: "svc-1",
							Port:      80,
						},
						{
							Type:      "Exact",
							Path:      "/api/v1",
							TargetSVC: "svc-2",
							Port:      8080,
						},
						{
							Type:      "ImplementationSpecific",
							Path:      "/api/v2",
							TargetSVC: "svc-3",
							Port:      8090,
						},
					},
				},
			},
		},
		Network: model.IngNetwork{
			CLBUseType: resCsts.CLBUseTypeUseExists,
			ExistLBID:  "lb-abcd",
			SubNetID:   "subnet-12345",
		},
		Cert: model.IngCert{
			AutoRewriteHTTP: true,
			TLS: []model.IngTLS{
				{
					SecretName: "secret-test-12345",
					Hosts: []string{
						"1.1.1.1",
						"2.2.2.2",
					},
				},
			},
		},
	},
}

// IngV1beta1 ...
var IngV1beta1 = model.Ing{
	Metadata: model.Metadata{
		APIVersion: "extensions/v1beta1",
		Kind:       resCsts.Ing,
		Name:       "ing-v1beta1-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
		},
		Annotations: []model.Annotation{
			{"anno-key-1", "anno-val-1"},
		},
	},
	Spec: model.IngSpec{
		RuleConf: model.IngRuleConf{
			Rules: []model.IngRule{
				{
					Domain: "example1.com",
					Paths: []model.IngPath{
						{
							Type:      "Prefix",
							Path:      "/api",
							TargetSVC: "svc-1",
							Port:      82,
						},
					},
				},
			},
		},
		Network: model.IngNetwork{
			CLBUseType: resCsts.CLBUseTypeAutoCreate,
			ExistLBID:  "",
			SubNetID:   "",
		},
		DefaultBackend: model.IngDefaultBackend{
			TargetSVC: "svc-2",
			Port:      8080,
		},
		Cert: model.IngCert{
			TLS: []model.IngTLS{
				{
					SecretName: "secret-test-54321",
					Hosts: []string{
						"1.1.1.1",
						"2.2.2.2",
					},
				},
			},
		},
	},
}

// SVCComplex ...
var SVCComplex = model.SVC{
	Metadata: model.Metadata{
		APIVersion: "v1",
		Kind:       resCsts.SVC,
		Name:       "svc-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
		},
		Annotations: []model.Annotation{
			{"anno-key-1", "anno-val-1"},
		},
	},
	Spec: model.SVCSpec{
		PortConf: model.SVCPortConf{
			Type: resCsts.SVCTypeLoadBalancer,
			LB: model.SVCLB{
				UseType:   resCsts.CLBUseTypeAutoCreate,
				ExistLBID: "lb-12345",
				SubNetID:  "subnet-id-1234",
			},
			Ports: []model.SVCPort{
				{
					Name:       "aaa",
					Port:       80,
					Protocol:   "TCP",
					TargetPort: "http",
				},
				{
					Name:       "bbb",
					Port:       81,
					Protocol:   "TCP",
					TargetPort: "8081",
					NodePort:   30000,
				},
				{
					Name:       "ccc",
					Port:       82,
					Protocol:   "TCP",
					TargetPort: "8082",
				},
			},
		},
		Selector: model.SVCSelector{
			Labels: []model.LabelSelector{
				{
					Key:   "select-123",
					Value: "456",
				},
			},
		},
		SessionAffinity: model.SessionAffinity{
			Type:       resCsts.SessionAffinityTypeClientIP,
			StickyTime: 10800,
		},
		IP: model.IPConf{
			External: []string{
				"2.2.2.2",
				"3.3.3.3",
			},
		},
	},
}

// EPComplex ...
var EPComplex = model.EP{
	Metadata: model.Metadata{
		APIVersion: "v1",
		Kind:       resCsts.EP,
		Name:       "ep-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
		},
		Annotations: []model.Annotation{
			{"anno-key-1", "anno-val-1"},
		},
	},
	Spec: model.EPSpec{
		SubSets: []model.SubSet{
			{
				Addresses: []string{
					"1.0.0.1",
					"1.0.0.2",
				},
				Ports: []model.EPPort{
					{
						Name:     "web",
						Protocol: "TCP",
						Port:     8080,
					},
					{
						Name:     "abc",
						Protocol: "UDP",
						Port:     8090,
					},
				},
			},
		},
	},
}
