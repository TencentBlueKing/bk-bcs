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

package network

import (
	"testing"

	"github.com/stretchr/testify/assert"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightV1IngManifest = map[string]interface{}{
	"apiVersion": "networking.k8s.io/v1",
	"kind":       "Ingress",
	"metadata": map[string]interface{}{
		"name": "ing-test-123456",
		"annotations": map[string]interface{}{
			"kubernetes.io/ingress.class":     resCsts.IngClsQCloud,
			"kubernetes.io/ingress.existLbId": "lb-abcd",
			"kubernetes.io/ingress.subnetId":  "subnet-12345",
		},
	},
	"spec": map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"host": "example1.com",
				"http": map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"backend": map[string]interface{}{
								"service": map[string]interface{}{
									"name": "svc-1",
									"port": map[string]interface{}{
										"number": int64(80),
									},
								},
							},
							"path":     "/api",
							"pathType": "Prefix",
						},
						map[string]interface{}{
							"backend": map[string]interface{}{
								"service": map[string]interface{}{
									"name": "svc-2",
									"port": map[string]interface{}{
										"number": int64(8080),
									},
								},
							},
							"path":     "/api/v1",
							"pathType": "Exact",
						},
						map[string]interface{}{
							"backend": map[string]interface{}{
								"service": map[string]interface{}{
									"name": "svc-3",
									"port": map[string]interface{}{
										"number": int64(8090),
									},
								},
							},
							"path":     "/api/v2",
							"pathType": "ImplementationSpecific",
						},
					},
				},
			},
		},
		"defaultBackend": map[string]interface{}{
			"service": map[string]interface{}{
				"name": "svc-4",
				"port": map[string]interface{}{
					"number": int64(443),
				},
			},
		},
		"tls": []interface{}{
			map[string]interface{}{
				"secretName": "secret-test-12345",
				"hosts": []interface{}{
					"1.1.1.1",
					"2.2.2.2",
				},
			},
		},
	},
}

var exceptedV1IngSpec = model.IngSpec{
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
	DefaultBackend: model.IngDefaultBackend{
		TargetSVC: "svc-4",
		Port:      443,
	},
	Network: model.IngNetwork{
		CLBUseType: resCsts.CLBUseTypeAutoCreate,
		ExistLBID:  "lb-abcd",
		SubNetID:   "subnet-12345",
	},
	Cert: model.IngCert{
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
}

func TestParseV1IngSpec(t *testing.T) {
	actualIngSpec := model.IngSpec{}
	ParseIngSpec(lightV1IngManifest, &actualIngSpec)
	assert.Equal(t, exceptedV1IngSpec, actualIngSpec)
}

var lightV1Beta1IngManifest = map[string]interface{}{
	"apiVersion": "networking.k8s.io/v1beta1",
	"kind":       "Ingress",
	"metadata": map[string]interface{}{
		"name": "ing-test-123456",
	},
	"spec": map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"host": "example1.com",
				"http": map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"backend": map[string]interface{}{
								"serviceName": "svc-1",
								"servicePort": int64(82),
							},
							"path":     "/api",
							"pathType": "Prefix",
						},
					},
				},
			},
		},
		"backend": map[string]interface{}{
			"serviceName": "svc-2",
			"servicePort": int64(8080),
		},
		"tls": []interface{}{
			map[string]interface{}{
				"secretName": "secret-test-54321",
				"hosts": []interface{}{
					"1.1.1.1",
					"2.2.2.2",
				},
			},
		},
	},
}

var exceptedV1Beta1IngSpec = model.IngSpec{
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
		CLBUseType: resCsts.CLBUseTypeUseExists,
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
}

func TestParseV1Beta1IngSpec(t *testing.T) {
	actualIngSpec := model.IngSpec{}
	ParseIngSpec(lightV1Beta1IngManifest, &actualIngSpec)
	assert.Equal(t, exceptedV1Beta1IngSpec, actualIngSpec)
}
