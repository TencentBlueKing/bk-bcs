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

package mapx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

var formData = map[string]interface{}{
	"metadata": map[string]interface{}{
		"annotations": []interface{}{},
		"labels": []interface{}{
			map[string]interface{}{
				"key":   "app",
				"value": "busybox",
			},
		},
		"name":      "busybox-deployment-12345",
		"namespace": "default",
	},
	"spec": map[string]interface{}{
		"affinity": map[string]interface{}{
			"nodeAffinity": []interface{}{},
			"podAffinity":  []interface{}{},
		},
		"nodeSelect": map[string]interface{}{
			"nodeName": "",
			"selector": []interface{}{},
			"type":     "anyAvailable",
		},
		"other": map[string]interface{}{
			"imagePullSecrets":           []interface{}{},
			"restartPolicy":              "",
			"saName":                     "",
			"terminationGracePeriodSecs": 0,
		},
		"security": map[string]interface{}{
			"runAsUser":    1111,
			"runAsGroup":   2222,
			"fsGroup":      3333,
			"runAsNonRoot": true,
			"seLinuxOpt": map[string]interface{}{
				"level": "",
				"role":  "",
				"type":  "",
				"user":  "",
			},
		},
		"toleration": map[string]interface{}{
			"rules": []interface{}{},
		},
	},
	"volume": map[string]interface{}{
		"hostPath": []interface{}{},
		"nfs": []interface{}{
			map[string]interface{}{
				"name":     "nfs",
				"path":     "/data",
				"readOnly": false,
				"server":   "1.1.1.1",
			},
		},
	},
	"containerGroup": map[string]interface{}{
		"containers": []interface{}{
			map[string]interface{}{
				"basic": map[string]interface{}{
					"image":      "busybox:latest",
					"name":       "busybox",
					"pullPolicy": "IfNotPresent",
				},
				"command": map[string]interface{}{
					"args": []interface{}{
						"echo hello",
					},
					"command": []interface{}{
						"/bin/bash",
						"-c",
					},
					"Stdin":      false,
					"stdinOnce":  true,
					"tty":        false,
					"workingDir": "/data/dev",
				},
				"envs": map[string]interface{}{
					"vars": []interface{}{},
				},
				"healthz": map[string]interface{}{
					"livenessProbe": map[string]interface{}{
						"command": []interface{}{
							"echo hello",
						},
						"failureThreshold": 3,
						"initialDelaySecs": 0,
						"path":             "",
						"periodSecs":       10,
						"port":             0,
						"successThreshold": 1,
						"timeoutSecs":      3,
						"type":             "exec",
					},
					"readinessProbe": map[string]interface{}{
						"command":          []interface{}{},
						"failureThreshold": 0,
						"initialDelaySecs": 0,
						"path":             "",
						"periodSecs":       0,
						"port":             0,
						"successThreshold": 0,
						"timeoutSecs":      0,
						"type":             "",
					},
				},
				"mount": map[string]interface{}{
					"volumes": []interface{}{},
				},
				"resource": map[string]interface{}{
					"limits": map[string]interface{}{
						"cpu":    500,
						"memory": 1024,
					},
					"requests": map[string]interface{}{
						"cpu":    0,
						"memory": 0,
					},
				},
			},
		},
		"initContainers": []interface{}{},
	},
}

var noZeroFormData = map[string]interface{}{
	"metadata": map[string]interface{}{
		"labels": []interface{}{
			map[string]interface{}{
				"key":   "app",
				"value": "busybox",
			},
		},
		"name":      "busybox-deployment-12345",
		"namespace": "default",
	},
	"spec": map[string]interface{}{
		"nodeSelect": map[string]interface{}{
			"type": "anyAvailable",
		},
		"security": map[string]interface{}{
			"runAsUser":    1111,
			"runAsGroup":   2222,
			"fsGroup":      3333,
			"runAsNonRoot": true,
		},
	},
	"volume": map[string]interface{}{
		"nfs": []interface{}{
			map[string]interface{}{
				"name":   "nfs",
				"path":   "/data",
				"server": "1.1.1.1",
			},
		},
	},
	"containerGroup": map[string]interface{}{
		"containers": []interface{}{
			map[string]interface{}{
				"basic": map[string]interface{}{
					"image":      "busybox:latest",
					"name":       "busybox",
					"pullPolicy": "IfNotPresent",
				},
				"command": map[string]interface{}{
					"args": []interface{}{
						"echo hello",
					},
					"command": []interface{}{
						"/bin/bash",
						"-c",
					},
					"stdinOnce":  true,
					"workingDir": "/data/dev",
				},
				"healthz": map[string]interface{}{
					"livenessProbe": map[string]interface{}{
						"command": []interface{}{
							"echo hello",
						},
						"failureThreshold": 3,
						"periodSecs":       10,
						"successThreshold": 1,
						"timeoutSecs":      3,
						"type":             "exec",
					},
				},
				"resource": map[string]interface{}{
					"limits": map[string]interface{}{
						"cpu":    500,
						"memory": 1024,
					},
				},
			},
		},
	},
}

// 清理 Map 空子项测试
func TestRemoveZeroSubItem(t *testing.T) {
	mapx.RemoveZeroSubItem(formData)
	assert.Equal(t, noZeroFormData, formData)
}
