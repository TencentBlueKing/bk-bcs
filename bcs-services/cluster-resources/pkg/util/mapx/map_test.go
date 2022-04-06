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

var deploySpec = map[string]interface{}{
	"testKey":              "testValue",
	"replicas":             3,
	"revisionHistoryLimit": 10,
	"intKey4SetItem":       8,
	"selector": map[string]interface{}{
		"matchLabels": map[string]interface{}{
			"app": "nginx",
		},
	},
	"strategy": map[string]interface{}{
		"rollingUpdate": map[string]interface{}{
			"maxSurge":       "25%",
			"maxUnavailable": "25%",
		},
		"type": "RollingUpdate",
	},
	"template": map[string]interface{}{
		"metadata": map[string]interface{}{
			"creationTimestamp": nil,
			"labels": map[string]interface{}{
				"app": "nginx",
			},
		},
		"spec": map[string]interface{}{
			"containers": []map[string]interface{}{
				{
					"image":           "nginx:latest",
					"imagePullPolicy": "IfNotPresent",
					"name":            "nginx",
					"ports": map[string]interface{}{
						"containerPort": 80,
						"protocol":      "TCP",
					},
					"resources": map[string]interface{}{},
				},
			},
			"dnsPolicy":                     "ClusterFirst",
			"restartPolicy":                 "Always",
			"schedulerName":                 "default-scheduler",
			"securityContext":               map[string]interface{}{},
			"terminationGracePeriodSeconds": 30,
		},
	},
}

// paths 为以 '.' 连接的字符串
func TestGetItems(t *testing.T) {
	// depth 1，val type int
	ret, _ := mapx.GetItems(deploySpec, "replicas")
	assert.Equal(t, 3, ret)

	// depth 2, val type string
	ret, _ = mapx.GetItems(deploySpec, "strategy.type")
	assert.Equal(t, "RollingUpdate", ret)

	// depth 3, val type string
	ret, _ = mapx.GetItems(deploySpec, "template.spec.restartPolicy")
	assert.Equal(t, "Always", ret)
}

// paths 为 []string，成功的情况
func TestGetItemsSuccessCase(t *testing.T) {
	// depth 1，val type int
	ret, _ := mapx.GetItems(deploySpec, []string{"replicas"})
	assert.Equal(t, 3, ret)

	// depth 2，val type map[string]interface{}
	r, _ := mapx.GetItems(deploySpec, []string{"selector", "matchLabels"})
	_, ok := r.(map[string]interface{})
	assert.Equal(t, true, ok)

	// depth 2, val type string
	ret, _ = mapx.GetItems(deploySpec, []string{"strategy", "type"})
	assert.Equal(t, "RollingUpdate", ret)

	// depth 3, val type nil
	ret, _ = mapx.GetItems(deploySpec, []string{"template", "metadata", "creationTimestamp"})
	assert.Nil(t, ret)

	// depth 3, val type string
	ret, _ = mapx.GetItems(deploySpec, []string{"template", "spec", "restartPolicy"})
	assert.Equal(t, "Always", ret)
}

// paths 为 []string 或 其他，失败的情况
func TestGetItemsFailCase(t *testing.T) {
	// not paths error
	_, err := mapx.GetItems(deploySpec, []string{})
	assert.NotNil(t, err)

	// not map[string]interface{} type error
	_, err = mapx.GetItems(deploySpec, []string{"replicas", "testKey"})
	assert.NotNil(t, err)

	_, err = mapx.GetItems(deploySpec, []string{"template", "spec", "containers", "image"})
	assert.NotNil(t, err)

	// key not exist
	_, err = mapx.GetItems(deploySpec, []string{"templateKey", "spec"})
	assert.NotNil(t, err)

	_, err = mapx.GetItems(deploySpec, []string{"selector", "spec"})
	assert.NotNil(t, err)

	// paths type error
	_, err = mapx.GetItems(deploySpec, []int{123, 456})
	assert.NotNil(t, err)

	_, err = mapx.GetItems(deploySpec, 123)
	assert.NotNil(t, err)
}

func TestGet(t *testing.T) {
	ret := mapx.Get(deploySpec, []string{"replicas"}, 1)
	assert.Equal(t, 3, ret)

	ret = mapx.Get(deploySpec, []string{}, nil)
	assert.Nil(t, ret)

	ret = mapx.Get(deploySpec, "container.name", "defaultName")
	assert.Equal(t, "defaultName", ret)
}

// SetItems 成功的情况
func TestSetItemsSuccessCase(t *testing.T) {
	// depth 1，val type int
	err := mapx.SetItems(deploySpec, "intKey4SetItem", 5)
	assert.Nil(t, err)
	ret, _ := mapx.GetItems(deploySpec, []string{"intKey4SetItem"})
	assert.Equal(t, 5, ret)

	// depth 2, val type string
	err = mapx.SetItems(deploySpec, "strategy.type", "Rolling")
	assert.Nil(t, err)
	ret, _ = mapx.GetItems(deploySpec, []string{"strategy", "type"})
	assert.Equal(t, "Rolling", ret)

	// depth 3, val type string
	err = mapx.SetItems(deploySpec, []string{"template", "spec", "restartPolicy"}, "Never")
	assert.Nil(t, err)
	ret, _ = mapx.GetItems(deploySpec, []string{"template", "spec", "restartPolicy"})
	assert.Equal(t, "Never", ret)

	// key noy exists
	err = mapx.SetItems(deploySpec, []string{"selector", "testKey"}, "testVal")
	assert.Nil(t, err)
	ret, _ = mapx.GetItems(deploySpec, "selector.testKey")
	assert.Equal(t, "testVal", ret)
}

// SetItems 失败的情况
func TestSetItemsFailCase(t *testing.T) {
	// not paths error
	err := mapx.SetItems(deploySpec, []string{}, 1)
	assert.NotNil(t, err)

	// not map[string]interface{} type error
	err = mapx.SetItems(deploySpec, []string{"replicas", "testKey"}, 1)
	assert.NotNil(t, err)

	// key not exist
	err = mapx.SetItems(deploySpec, []string{"templateKey", "spec"}, 1)
	assert.NotNil(t, err)

	err = mapx.SetItems(deploySpec, "templateKey.spec", 123)
	assert.NotNil(t, err)

	// paths type error
	err = mapx.SetItems(deploySpec, []int{123, 456}, 1)
	assert.NotNil(t, err)

	err = mapx.SetItems(deploySpec, 123, 1)
	assert.NotNil(t, err)
}

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
