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

package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
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
	ret, _ := util.GetItems(deploySpec, "replicas")
	assert.Equal(t, 3, ret)

	// depth 2, val type string
	ret, _ = util.GetItems(deploySpec, "strategy.type")
	assert.Equal(t, "RollingUpdate", ret)

	// depth 3, val type string
	ret, _ = util.GetItems(deploySpec, "template.spec.restartPolicy")
	assert.Equal(t, "Always", ret)
}

// paths 为 []string，成功的情况
func TestGetItemsSuccessCase(t *testing.T) {
	// depth 1，val type int
	ret, _ := util.GetItems(deploySpec, []string{"replicas"})
	assert.Equal(t, 3, ret)

	// depth 2，val type map[string]interface{}
	r, _ := util.GetItems(deploySpec, []string{"selector", "matchLabels"})
	_, ok := r.(map[string]interface{})
	assert.Equal(t, true, ok)

	// depth 2, val type string
	ret, _ = util.GetItems(deploySpec, []string{"strategy", "type"})
	assert.Equal(t, "RollingUpdate", ret)

	// depth 3, val type nil
	ret, _ = util.GetItems(deploySpec, []string{"template", "metadata", "creationTimestamp"})
	assert.Nil(t, ret)

	// depth 3, val type string
	ret, _ = util.GetItems(deploySpec, []string{"template", "spec", "restartPolicy"})
	assert.Equal(t, "Always", ret)
}

// paths 为 []string 或 其他，失败的情况
func TestGetItemsFailCase(t *testing.T) {
	// not paths error
	_, err := util.GetItems(deploySpec, []string{})
	assert.NotNil(t, err)

	// not map[string]interface{} type error
	_, err = util.GetItems(deploySpec, []string{"replicas", "testKey"})
	assert.NotNil(t, err)

	_, err = util.GetItems(deploySpec, []string{"template", "spec", "containers", "image"})
	assert.NotNil(t, err)

	// key not exist
	_, err = util.GetItems(deploySpec, []string{"templateKey", "spec"})
	assert.NotNil(t, err)

	_, err = util.GetItems(deploySpec, []string{"selector", "spec"})
	assert.NotNil(t, err)

	// paths type error
	_, err = util.GetItems(deploySpec, []int{123, 456})
	assert.NotNil(t, err)

	_, err = util.GetItems(deploySpec, 123)
	assert.NotNil(t, err)
}

func TestGetWithDefault(t *testing.T) {
	ret := util.GetWithDefault(deploySpec, []string{"replicas"}, 1)
	assert.Equal(t, 3, ret)

	ret = util.GetWithDefault(deploySpec, []string{}, nil)
	assert.Nil(t, ret)

	ret = util.GetWithDefault(deploySpec, "container.name", "defaultName")
	assert.Equal(t, "defaultName", ret)
}

// SetItems 成功的情况
func TestSetItemsSuccessCase(t *testing.T) {
	// depth 1，val type int
	err := util.SetItems(deploySpec, "intKey4SetItem", 5)
	assert.Nil(t, err)
	ret, _ := util.GetItems(deploySpec, []string{"intKey4SetItem"})
	assert.Equal(t, 5, ret)

	// depth 2, val type string
	err = util.SetItems(deploySpec, "strategy.type", "Rolling")
	assert.Nil(t, err)
	ret, _ = util.GetItems(deploySpec, []string{"strategy", "type"})
	assert.Equal(t, "Rolling", ret)

	// depth 3, val type string
	err = util.SetItems(deploySpec, []string{"template", "spec", "restartPolicy"}, "Never")
	assert.Nil(t, err)
	ret, _ = util.GetItems(deploySpec, []string{"template", "spec", "restartPolicy"})
	assert.Equal(t, "Never", ret)

	// key noy exists
	err = util.SetItems(deploySpec, []string{"selector", "testKey"}, "testVal")
	assert.Nil(t, err)
	ret, _ = util.GetItems(deploySpec, "selector.testKey")
	assert.Equal(t, "testVal", ret)
}

// SetItems 失败的情况
func TestSetItemsFailCase(t *testing.T) {
	// not paths error
	err := util.SetItems(deploySpec, []string{}, 1)
	assert.NotNil(t, err)

	// not map[string]interface{} type error
	err = util.SetItems(deploySpec, []string{"replicas", "testKey"}, 1)
	assert.NotNil(t, err)

	// key not exist
	err = util.SetItems(deploySpec, []string{"templateKey", "spec"}, 1)
	assert.NotNil(t, err)

	err = util.SetItems(deploySpec, "templateKey.spec", 123)
	assert.NotNil(t, err)

	// paths type error
	err = util.SetItems(deploySpec, []int{123, 456}, 1)
	assert.NotNil(t, err)

	err = util.SetItems(deploySpec, 123, 1)
	assert.NotNil(t, err)
}
