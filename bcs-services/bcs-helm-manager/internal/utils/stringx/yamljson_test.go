/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package stringx

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"

	goyaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestYaml2Json(t *testing.T) {
	y := `
test:
  test: 1
  test1: "a"
`
	j, err := Yaml2Json(y)
	assert.Nil(t, err)
	assert.Equal(t, map[interface{}]interface{}{"test": map[interface{}]interface{}{"test": 1, "test1": "a"}}, j)
}

func TestJson2Yaml(t *testing.T) {
	j := map[interface{}]interface{}{"test": map[interface{}]interface{}{"test": 1, "test1": "a"}}
	y, err := Json2Yaml(j)
	expectedYaml := `test:
  test: 1
  test1: a
`
	assert.Nil(t, err)
	assert.Equal(t, expectedYaml, string(y))
}

func TestYamlNode(t *testing.T) {
	y := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap-test
data:
  group.yaml: |2-
    a: b
    c: a
---
apiVersion: v1
kind: Service
metadata:
# comment 1
  name: test-service
  namespace: default
  labels:
    k1: v1
    c: d
---
apiVersion: v1
kind: Service
metadata:
# comment 1
  name: test-service2
  namespace: default
  labels:
    k1: v1
    c: d
`
	var n yaml.MapSlice
	err := yaml.Unmarshal([]byte(y), &n)
	if err != nil {
		t.Error(err)
	}
	// fmt.Println(n, err)
	// fmt.Println(n[2])
	var out []byte
	out, err = yaml.Marshal(&n)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(out), err)
}

// 乱序
func TestYamlMap(t *testing.T) {
	y := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap-test
data:
  group.yaml: |2-
    a: b
    c: a
---
apiVersion: v1
kind: Service
metadata:
# comment 1
  name: test-service
  namespace: default
  labels:
    k1: v1
    c: d
`
	n := make(map[interface{}]interface{}, 0)
	err := yaml.Unmarshal([]byte(y), &n)
	fmt.Println(n, err)

	out, err := yaml.Marshal(&n)
	fmt.Println(string(out), err)
}

func TestYamlPath(t *testing.T) {
	y := `
apiVersion: v1
kind: Service
metadata:
# comment 1
  name: test-service
  namespace: default
  labels:
    k1: v1
    c.a: d
`
	path, _ := goyaml.PathString("$.metadata.name")
	path2, _ := goyaml.PathString("$.metadata.namespace")
	labelpath, _ := goyaml.PathString(fmt.Sprintf("$.metadata.labels.'%s'", "c.a"))
	var name string
	path.Read(strings.NewReader(y), &name)
	f, _ := parser.ParseBytes([]byte(y), 0)
	path.ReplaceWithReader(f, strings.NewReader("test"))
	path2.ReplaceWithReader(f, strings.NewReader("test-ns"))
	labelpath.ReplaceWithReader(f, strings.NewReader(name))
	fmt.Println(name)
	fmt.Println(f.String())
}

func TestYamlYaml(t *testing.T) {
	y := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap-test
  labels:
    k1: v1
data:
  group.yaml: |2-
    a: b
    c: 1
`
	var n yaml.MapSlice
	_ = yaml.Unmarshal([]byte(y), &n)
	out, _ := yaml.Marshal(&n)
	path, _ := goyaml.PathString("$.metadata.name")
	kindpath, _ := goyaml.PathString("$.kind")
	var name string
	var kind string
	path.Read(bytes.NewReader(out), &name)
	kindpath.Read(bytes.NewReader(out), &kind)
	fmt.Println(kind)
	f, err := parser.ParseBytes(out, 0)
	if err != nil {
		log.Fatal(err)
	}
	path.ReplaceWithReader(f, strings.NewReader("test"))
	fmt.Println(name)
	fmt.Println(f.String())
}
