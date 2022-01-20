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

var lightDeploySpec = map[string]interface{}{
	"template": map[string]interface{}{
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{"image": "nginx:1.14.2"},
				map[string]interface{}{"image": "nginx:latest"},
				map[string]interface{}{"image": "nginx:1.14.2"},
			},
		},
	},
}

var lightCronJobSpec = map[string]interface{}{
	"jobTemplate": map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{"image": "busybox:1.0.0"},
						map[string]interface{}{"image": "busybox:1.0.0"},
						map[string]interface{}{"image": "busybox:1.2.3"},
						map[string]interface{}{"image": "busybox:latest"},
					},
				},
			},
		},
	},
}

func TestParseContainerImages(t *testing.T) {
	images := parseContainerImages(lightDeploySpec, "template.spec.containers")
	assert.Equal(t, 2, len(images))

	images = parseContainerImages(lightCronJobSpec, "jobTemplate.spec.template.spec.containers")
	assert.Equal(t, 3, len(images))
}

var failedPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Failed",
		"conditions": []map[string]interface{}{
			{
				"type":   "Initialized",
				"status": "True",
			},
		},
	},
}

var succeededPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Succeeded",
		"conditions": []map[string]interface{}{
			{
				"type":   "Initialized",
				"status": "True",
			},
		},
	},
}

var runningPodManifest1 = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Running",
		"conditions": []map[string]interface{}{
			{
				"type":   "Initialized",
				"status": "True",
			},
			{
				"type":   "Ready",
				"status": "True",
			},
		},
	},
}

var pendingPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Pending",
		"conditions": []map[string]interface{}{
			{
				"type":   "Initialized",
				"status": "False",
			},
		},
	},
}

var terminatingPodManifest = map[string]interface{}{
	"metadata": map[string]interface{}{
		"deletionTimestamp": "2021-01-01T10:00:00Z",
	},
	"status": map[string]interface{}{
		"phase": "Running",
	},
}

var unknownPodManifest1 = map[string]interface{}{
	"metadata": map[string]interface{}{
		"deletionTimestamp": "2021-01-01T10:00:00Z",
	},
	"status": map[string]interface{}{
		"phase":  "Running",
		"reason": "NodeLost",
	},
}

var completedPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Succeeded",
		"containerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"terminated": map[string]interface{}{
						"reason": "Completed",
					},
				},
			},
		},
	},
}

var waitReasonPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Pending",
		"containerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"waiting": map[string]interface{}{
						"message": "Error response from daemon: No command specified",
						"reason":  "CreateContainerError",
					},
				},
			},
		},
	},
}

var signalPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Failed",
		"containerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"terminated": map[string]interface{}{
						"signal": 1,
					},
				},
			},
		},
	},
}

var exitCodePodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Failed",
		"containerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"terminated": map[string]interface{}{
						"ExitCode": 1,
					},
				},
			},
		},
	},
}

var runningPodManifest2 = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Completed",
		"conditions": []map[string]interface{}{
			{
				"type":   "Ready",
				"status": "True",
			},
		},
		"containerStatuses": []interface{}{
			map[string]interface{}{
				"ready": true,
				"state": map[string]interface{}{
					"running": map[string]interface{}{
						"startAt": "2022-01-01T10:00:00Z",
					},
				},
			},
		},
	},
}

var notReadyPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Completed",
		"conditions": []map[string]interface{}{
			{
				"type":   "Ready",
				"status": "False",
			},
		},
		"containerStatuses": []interface{}{
			map[string]interface{}{
				"ready": true,
				"state": map[string]interface{}{
					"running": map[string]interface{}{
						"startAt": "2022-01-01T10:00:00Z",
					},
				},
			},
		},
	},
}

var unknownPodManifest2 = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": nil,
	},
}

var initSignalPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Pending",
		"initContainerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"terminated": map[string]interface{}{
						"exitCode": 0,
					},
				},
			},
			map[string]interface{}{
				"state": map[string]interface{}{
					"terminated": map[string]interface{}{
						"exitCode": 1,
						"signal":   1,
					},
				},
			},
		},
	},
}

var initExitCodePodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Pending",
		"initContainerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"terminated": map[string]interface{}{
						"exitCode": 1,
					},
				},
			},
		},
	},
}

var initTermPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Pending",
		"initContainerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"terminated": map[string]interface{}{
						"exitCode": 1,
						"reason":   "term init",
					},
				},
			},
		},
	},
}

var initWaitPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Pending",
		"initContainerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"waiting": map[string]interface{}{
						"reason": "wait init",
					},
				},
			},
		},
	},
}

var initRunPodManifest = map[string]interface{}{
	"spec": map[string]interface{}{
		"initContainers": []interface{}{
			map[string]interface{}{
				"name": "nginx",
			},
		},
	},
	"status": map[string]interface{}{
		"phase": "Pending",
		"initContainerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"running": map[string]interface{}{
						"startAt": "2022-01-01T10:00:00Z",
					},
				},
			},
		},
	},
}

func TestPodStatusParser(t *testing.T) {
	parser := podStatusParser{manifest: failedPodManifest}
	assert.Equal(t, "Failed", parser.parse())

	parser = podStatusParser{manifest: succeededPodManifest}
	assert.Equal(t, "Succeeded", parser.parse())

	parser = podStatusParser{manifest: runningPodManifest1}
	assert.Equal(t, "Running", parser.parse())

	parser = podStatusParser{manifest: pendingPodManifest}
	assert.Equal(t, "Pending", parser.parse())

	parser = podStatusParser{manifest: terminatingPodManifest}
	assert.Equal(t, "Terminating", parser.parse())

	parser = podStatusParser{manifest: unknownPodManifest1}
	assert.Equal(t, "Unknown", parser.parse())

	parser = podStatusParser{manifest: completedPodManifest}
	assert.Equal(t, "Completed", parser.parse())

	parser = podStatusParser{manifest: waitReasonPodManifest}
	assert.Equal(t, "CreateContainerError", parser.parse())

	parser = podStatusParser{manifest: signalPodManifest}
	assert.Equal(t, "Signal: 1", parser.parse())

	parser = podStatusParser{manifest: exitCodePodManifest}
	assert.Equal(t, "ExitCode: 1", parser.parse())

	parser = podStatusParser{manifest: runningPodManifest2}
	assert.Equal(t, "Running", parser.parse())

	parser = podStatusParser{manifest: notReadyPodManifest}
	assert.Equal(t, "NotReady", parser.parse())

	parser = podStatusParser{manifest: unknownPodManifest2}
	assert.Equal(t, "Unknown", parser.parse())

	parser = podStatusParser{manifest: initSignalPodManifest}
	assert.Equal(t, "Init: Signal 1", parser.parse())

	parser = podStatusParser{manifest: initExitCodePodManifest}
	assert.Equal(t, "Init: ExitCode 1", parser.parse())

	parser = podStatusParser{manifest: initTermPodManifest}
	assert.Equal(t, "Init: term init", parser.parse())

	parser = podStatusParser{manifest: initWaitPodManifest}
	assert.Equal(t, "Init: wait init", parser.parse())

	parser = podStatusParser{manifest: initRunPodManifest}
	assert.Equal(t, "Init: 0/1", parser.parse())
}

var lightIngManifest = map[string]interface{}{
	"apiVersion": "networking.k8s.io/v1",
	"kind":       "Ingress",
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

var lightSVCManifest = map[string]interface{}{
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

var lightEndpointsManifest = map[string]interface{}{
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
