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

package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var lightIngManifest = map[string]interface{}{
	"apiVersion": "networking.k8s.io/v1",
	"kind":       "Ingress",
	"metadata": map[string]interface{}{
		"creationTimestamp": "2022-01-01T10:00:00Z",
	},
	"spec": map[string]interface{}{
		"ingressClassName": "nginx",
		"rules": []interface{}{
			map[string]interface{}{
				"host": "bcs-cr.example.com",
				"http": map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"backend": map[string]interface{}{
								"service": map[string]interface{}{
									"name": "bcs-cr-test",
									"port": map[string]interface{}{
										"number": 20001,
									},
								},
							},
							"path":     "/metric/",
							"pathType": "Prefix",
						},
						map[string]interface{}{
							"backend": map[string]interface{}{
								"service": map[string]interface{}{
									"name": "bcs-cr-test",
									"port": map[string]interface{}{
										"number": 20000,
									},
								},
							},
							"path":     "/",
							"pathType": "Prefix",
						},
					},
				},
			},
		},
	},
	"status": map[string]interface{}{
		"loadBalancer": map[string]interface{}{
			"ingress": []interface{}{
				map[string]interface{}{
					"ip": "127.0.0.1",
				},
				map[string]interface{}{
					"hostname": "localhost",
				},
			},
		},
	},
}

var lightV1beta1IngManifest = map[string]interface{}{
	"apiVersion": "extensions/v1beta1",
	"kind":       "Ingress",
	"spec": map[string]interface{}{
		"ingressClassName": "nginx",
		"tls": []interface{}{
			map[string]interface{}{
				"host": []interface{}{
					"bcs-cr-tls.example.com",
				},
				"secretName": "secret-tls",
			},
		},
		"rules": []interface{}{
			map[string]interface{}{
				"host": "bcs-cr-tls.example.com",
				"http": map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"backend": map[string]interface{}{
								"serviceName": "bcs-cr-test",
								"servicePort": 20000,
							},
							"path": "/",
						},
					},
				},
			},
		},
	},
}

func TestParseIngHosts(t *testing.T) {
	assert.Equal(t, []string{"bcs-cr.example.com"}, parseIngHosts(lightIngManifest))
	assert.Equal(t, []string{"bcs-cr-tls.example.com"}, parseIngHosts(lightV1beta1IngManifest))
}

func TestParseIngAddrs(t *testing.T) {
	assert.Equal(t, []string{"127.0.0.1", "localhost"}, parseIngAddrs(lightIngManifest))
	assert.Equal(t, []string(nil), parseIngAddrs(lightV1beta1IngManifest))
}

func TestParseIngDefaultPort(t *testing.T) {
	assert.Equal(t, "80", getIngDefaultPort(lightIngManifest))
	assert.Equal(t, "80, 443", getIngDefaultPort(lightV1beta1IngManifest))
}

func TestParseIngRules(t *testing.T) {
	excepted := []map[string]interface{}{
		{
			"host":        "bcs-cr.example.com",
			"path":        "/metric/",
			"pathType":    "Prefix",
			"serviceName": "bcs-cr-test",
			"port":        20001,
		},
		{
			"host":        "bcs-cr.example.com",
			"path":        "/",
			"pathType":    "Prefix",
			"serviceName": "bcs-cr-test",
			"port":        20000,
		},
	}
	assert.Equal(t, excepted, parseV1IngRules(lightIngManifest))

	excepted = []map[string]interface{}{
		{
			"host":        "bcs-cr-tls.example.com",
			"path":        "/",
			"pathType":    "--",
			"serviceName": "bcs-cr-test",
			"port":        20000,
		},
	}
	assert.Equal(t, excepted, parseV1beta1IngRules(lightV1beta1IngManifest))
}

func TestFormatIng(t *testing.T) {
	ret := FormatIng(lightIngManifest)
	assert.Equal(t, []string{"bcs-cr.example.com"}, ret["hosts"])
	assert.Equal(t, []string{"127.0.0.1", "localhost"}, ret["addresses"])
	assert.Equal(t, "80", ret["defaultPorts"])
	excepted := []map[string]interface{}{
		{
			"host":        "bcs-cr.example.com",
			"path":        "/metric/",
			"pathType":    "Prefix",
			"serviceName": "bcs-cr-test",
			"port":        20001,
		},
		{
			"host":        "bcs-cr.example.com",
			"path":        "/",
			"pathType":    "Prefix",
			"serviceName": "bcs-cr-test",
			"port":        20000,
		},
	}
	assert.Equal(t, excepted, ret["rules"])
}

var lightSVCManifest = map[string]interface{}{
	"metadata": map[string]interface{}{
		"creationTimestamp": "2022-01-01T10:00:00Z",
	},
	"spec": map[string]interface{}{
		"ports": []interface{}{
			map[string]interface{}{
				"nodePort":   30600,
				"port":       8080,
				"protocol":   "TCP",
				"targetPort": 8080,
			},
			map[string]interface{}{
				"port":       8090,
				"protocol":   "TCP",
				"targetPort": 8090,
			},
		},
	},
	"status": map[string]interface{}{
		"loadBalancer": map[string]interface{}{
			"ingress": []interface{}{
				map[string]interface{}{
					"ip": "127.0.0.1",
				},
				map[string]interface{}{
					"hostname": "localhost",
				},
			},
		},
	},
}

func TestParseSVCExternalIPs(t *testing.T) {
	assert.Equal(t, []string{"127.0.0.1", "localhost"}, parseSVCExternalIPs(lightSVCManifest))
}

func TestParseSVCPorts(t *testing.T) {
	assert.Equal(t, []string{"8080:30600/TCP", "8090/TCP"}, parseSVCPorts(lightSVCManifest))
}

func TestFormatSVC(t *testing.T) {
	ret := FormatSVC(lightSVCManifest)
	assert.Equal(t, []string{"127.0.0.1", "localhost"}, ret["externalIP"])
	assert.Equal(t, []string{"8080:30600/TCP", "8090/TCP"}, ret["ports"])
}

var lightEndpointsManifest = map[string]interface{}{
	"metadata": map[string]interface{}{
		"creationTimestamp": "2022-01-01T10:00:00Z",
	},
	"subsets": []interface{}{
		map[string]interface{}{
			"addresses": []interface{}{
				map[string]interface{}{
					"ip": "127.0.0.1",
				},
				map[string]interface{}{
					"ip": "127.0.0.2",
				},
			},
			"ports": []interface{}{
				map[string]interface{}{
					"port": 80,
				},
				map[string]interface{}{
					"port": 90,
				},
			},
		},
	},
}

func TestParseEndpoints(t *testing.T) {
	excepted := []string{"127.0.0.1:80", "127.0.0.1:90", "127.0.0.2:80", "127.0.0.2:90"}
	assert.Equal(t, excepted, parseEndpoints(lightEndpointsManifest))
}

func TestFormatEP(t *testing.T) {
	ret := FormatEP(lightEndpointsManifest)
	assert.Equal(t, []string{"127.0.0.1:80", "127.0.0.1:90", "127.0.0.2:80", "127.0.0.2:90"}, ret["endpoints"])
	assert.Equal(t, "2022-01-01 10:00:00", ret["createTime"])
}
