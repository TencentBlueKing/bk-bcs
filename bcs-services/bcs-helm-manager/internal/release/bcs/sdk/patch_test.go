/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sdk

import (
	"strings"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
)

// 开关只作用于 io.tencent.paas.updator，其余元数据一律保留：
//   - creator / version 虽然也是 BCS 注入的，但属可追溯信息，不在清理范围
//   - source_type / projectid 属「来源 / 归属」分组，同样不清理
//   - io.tencent.bcs.* 是 BCS 控制信息，与本开关无关
var keepKeys = []string{
	"io.tencent.paas.creator",
	"io.tencent.paas.version",
	"io.tencent.paas.source_type",
	"io.tencent.paas.projectid",
	"io.tencent.bcs.controller.name",
	"other.io/keep",
}

// 开关范围内、必须被删除的 key
const removedKey = "io.tencent.paas.updator"

// 锁定开关作用范围：只能含 updator。
// 若后续要扩大范围（例如把 creator / version 也纳入），必须是有意识的决策 —— 该用例会失败并强制review。
func TestRemovePaasMetadataScopeIsUpdatorOnly(t *testing.T) {
	if len(common.PaasMetadataKeys) != 1 || common.PaasMetadataKeys[0] != removedKey {
		t.Fatalf("PaasMetadataKeys should contain only %q, got %v", removedKey, common.PaasMetadataKeys)
	}
}

func TestRemovePaasMetadata(t *testing.T) {
	in := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo
  annotations:
    io.tencent.paas.creator: alice
    io.tencent.paas.updator: bob
    io.tencent.paas.version: "1.0"
    io.tencent.paas.source_type: helm
    other.io/keep: "yes"
  labels:
    io.tencent.paas.projectid: p1
    io.tencent.paas.updator: bob
    app: demo
spec:
  template:
    metadata:
      labels:
        io.tencent.bcs.controller.name: demo
`
	got, err := removePaasMetadata(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, k := range common.PaasMetadataKeys {
		if strings.Contains(got, k) {
			t.Errorf("key %q should be removed, got:\n%s", k, got)
		}
	}
	for _, k := range keepKeys {
		if !strings.Contains(got, k) {
			t.Errorf("key %q should be kept, got:\n%s", k, got)
		}
	}
	if !strings.Contains(got, "app: demo") {
		t.Errorf("non paas label should be kept, got:\n%s", got)
	}
}

// 没有命中任何待清理 key 时必须原样返回，不做无谓的重新格式化
func TestRemovePaasMetadataNoHit(t *testing.T) {
	in := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cm\n" +
		"  annotations:\n    foo: bar\n"
	got, err := removePaasMetadata(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != in {
		t.Errorf("manifest without paas keys should be returned as-is, got:\n%s", got)
	}
}

// annotations 被整个删空时输出必须是 {}，[] 会被 K8s 拒绝
func TestRemovePaasMetadataEmptyMap(t *testing.T) {
	in := "apiVersion: v1\nkind: Service\nmetadata:\n" +
		"  annotations:\n    io.tencent.paas.updator: bob\n"
	got, err := removePaasMetadata(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(got, "annotations: []") {
		t.Errorf("empty annotations should be {} instead of [], got:\n%s", got)
	}
	if strings.Contains(got, removedKey) {
		t.Errorf("updator should be removed, got:\n%s", got)
	}
	t.Logf("empty map result:\n%s", got)
}

// 嵌套路径（spec.template.metadata）下的 updator 同样要清理，
// 该用例同时覆盖此前 goccy AST 序列化触发的 strings.Repeat panic
func TestRemovePaasMetadataNestedPath(t *testing.T) {
	in := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: x
  annotations:
    io.tencent.paas.creator: alice
    io.tencent.paas.version: "1"
    io.tencent.paas.updator: bob
spec:
  template:
    metadata:
      annotations:
        io.tencent.paas.updator: bob
      labels:
        io.tencent.paas.updator: bob
        app: x
`
	got, err := removePaasMetadata(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(got, removedKey) {
		t.Errorf("updator should be removed at every path, got:\n%s", got)
	}
	for _, k := range []string{"io.tencent.paas.creator", "io.tencent.paas.version"} {
		if !strings.Contains(got, k) {
			t.Errorf("key %q should be kept, got:\n%s", k, got)
		}
	}
	if !strings.Contains(got, "app: x") {
		t.Errorf("non paas key should be kept, got:\n%s", got)
	}
}
