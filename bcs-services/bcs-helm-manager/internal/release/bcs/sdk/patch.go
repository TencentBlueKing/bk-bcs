/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sdk

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	goyaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
	cmdtpl "github.com/vmware-tanzu/carvel-ytt/pkg/cmd/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/cmd/ui"
	"github.com/vmware-tanzu/carvel-ytt/pkg/files"
	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/stringx"
)

const (
	resourceFilename = "resource.yaml"
	labelKey         = "io.tencent.bcs.controller.name"
)

func replacePatchTplKey(keys map[string]string, data []byte) []byte {
	for k, v := range keys {
		if !common.IsPatchTemplateKey(k) {
			continue
		}

		data = []byte(strings.ReplaceAll(string(data), k, v))
	}

	return common.EmptyAllPatchTemplateKey(data)
}

// newPatcher 构建 post-render patcher。
// helmNative 为 true 时，会在渲染完成后删除 BCS 注入的 io.tencent.paas.updator 元数据。
func newPatcher(templates []*release.File, keys map[string]string, helmNative bool) *patcher {
	fs := make([]*files.File, 0, 5)
	for _, f := range templates {
		fs = append(fs, files.MustNewFileFromSource(files.NewBytesSource(f.Name, replacePatchTplKey(keys, f.Content))))
	}

	return &patcher{
		files:      fs,
		helmNative: helmNative,
	}
}

type patcher struct {
	files []*files.File
	// helmNative 为 true 时删除 BCS 注入的 io.tencent.paas.updator 元数据
	helmNative bool
}

// Run implements the post-render Run method, do the render
func (p *patcher) Run(renderedManifests *bytes.Buffer) (*bytes.Buffer, error) {
	manifest, err := p.do(renderedManifests)
	if err != nil {
		return nil, err
	}
	// 注入指定的值
	splitManifests := stringx.SplitManifests(manifest.String())
	if err != nil {
		return nil, fmt.Errorf("SplitYAML error, %s", err.Error())
	}
	var yList []string
	for _, s := range splitManifests {
		// 向 metadata.labels 中注入 `io.tencent.bcs.controller.name`
		// 向 spec.template.metadata.labels 中注入 `io.tencent.bcs.controller.name`
		var j string
		j, err = inject4MetadataLabels(s)
		if err != nil {
			blog.Errorf("inject4MetadataLabels error, %s", err.Error())
			yList = append(yList, s)
			continue
		}

		// 用户选择不注入时，删除 BCS 注入的 io.tencent.paas.updator 元数据（annotations + labels）
		if p.helmNative {
			j, err = removePaasMetadata(j)
			if err != nil {
				blog.Errorf("removePaasMetadata error, %s", err.Error())
				yList = append(yList, s)
				continue
			}
		}

		yList = append(yList, strings.TrimRight(j, "\n"))
	}
	yl := stringx.JoinStringBySeparator(yList, "", false)
	// 添加换行
	yl += "\n"
	// 写回数据
	buf := new(bytes.Buffer)
	_, err = buf.WriteString(yl)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (p *patcher) do(data *bytes.Buffer) (*bytes.Buffer, error) {
	if data == nil {
		return nil, fmt.Errorf("empty resource data")
	}

	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	fakeUI := ui.NewCustomWriterTTY(false, stdout, stderr)
	opts := cmdtpl.NewOptions()
	out := opts.RunWithFiles(cmdtpl.Input{
		Files: append(p.files, files.MustNewFileFromSource(files.NewBytesSource(resourceFilename, data.Bytes()))),
	}, fakeUI)
	if out.Err != nil {
		return nil, out.Err
	}
	if len(out.Files) == 0 {
		return data, nil
	}
	return bytes.NewBuffer(out.Files[0].Bytes()), nil
}

// inject4MetadataLabels 兼容逻辑，目的是向metadata注入label
func inject4MetadataLabels(manifest string) (string, error) {
	// 转换为常用格式, goyaml 库不能识别 |2- 之类的描述符
	var n yaml.MapSlice
	if err := yaml.Unmarshal([]byte(manifest), &n); err != nil {
		return manifest, err
	}
	out, err := yaml.Marshal(&n)
	if err != nil {
		return manifest, err
	}
	s := string(out)

	// 限制下面几个注入指定的 key:val
	kinds := []string{"Deployment", "StatefulSet", "Job", "DaemonSet"}

	// parse name
	namePath, err := goyaml.PathString("$.metadata.name")
	if err != nil {
		return s, err
	}
	var name string
	if err = namePath.Read(strings.NewReader(s), &name); err != nil {
		return s, err
	}

	// parse kind
	kindPath, err := goyaml.PathString("$.kind")
	if err != nil {
		return s, err
	}
	var kind string
	if err = kindPath.Read(strings.NewReader(s), &kind); err != nil {
		return s, err
	}

	// parse label
	labelPath, err := goyaml.PathString(fmt.Sprintf("$.metadata.labels.'%s'", labelKey))
	if err != nil {
		return s, err
	}
	// parse spec label
	specLabelPath, err := goyaml.PathString(fmt.Sprintf("$.spec.template.metadata.labels.'%s'", labelKey))
	if err != nil {
		return s, err
	}

	// parse origin yaml
	f, err := parser.ParseBytes([]byte(s), 0)
	if err != nil {
		return s, err
	}

	// inject service metadata
	if kind == "Service" {
		if err := labelPath.ReplaceWithReader(f, strings.NewReader(name)); err != nil {
			return s, err
		}
		return f.String(), nil
	}
	if stringx.StringInSlice(kind, kinds) {
		if err := labelPath.ReplaceWithReader(f, strings.NewReader(name)); err != nil {
			return s, err
		}
		if err := specLabelPath.ReplaceWithReader(f, strings.NewReader(name)); err != nil {
			return s, err
		}
		return f.String(), nil
	}

	return manifest, nil
}

// paasMetaPaths 需要清理 paas 元数据的位置，元素为逐级的 key 路径。
// annotations 与 labels 都要处理：updator 在 metadata 与 spec.template 两处都存在，只清一处会漏。
// spec.template 下的两份是防御性覆盖，路径不存在时自动跳过，也不会触发 Pod 滚动。
var paasMetaPaths = [][]string{
	{"metadata", "annotations"},
	{"metadata", "labels"},
	{"spec", "template", "metadata", "annotations"},
	{"spec", "template", "metadata", "labels"},
}

// removePaasMetadata 删除 BCS 注入的 io.tencent.paas.updator 元数据（annotations 与 labels）。
func removePaasMetadata(manifest string) (string, error) {
	var n yaml.MapSlice
	if err := yaml.Unmarshal([]byte(manifest), &n); err != nil {
		return manifest, err
	}

	root := interface{}(n)
	removed := false
	for _, path := range paasMetaPaths {
		v, ok := removePaasMetadataInPath(root, path)
		if ok {
			removed = true
			root = v
		}
	}

	if !removed {
		return manifest, nil
	}

	out, err := yaml.Marshal(root)
	if err != nil {
		return manifest, err
	}
	return string(out), nil
}

// String() 内部 strings.Repeat 会panic进而让进程崩溃。
func removePaasMetadataInPath(node interface{}, keys []string) (interface{}, bool) {
	if len(keys) == 0 {
		return removePaasMetadataInMap(node)
	}

	removed := false
	switch v := node.(type) {
	case yaml.MapSlice:
		for i := range v {
			k, ok := v[i].Key.(string)
			if !ok || k != keys[0] {
				continue
			}
			nv, r := removePaasMetadataInPath(v[i].Value, keys[1:])
			if r {
				v[i].Value = nv
				removed = true
			}
		}
		return v, removed
	case map[interface{}]interface{}:
		for k, val := range v {
			ks, ok := k.(string)
			if !ok || ks != keys[0] {
				continue
			}
			nv, r := removePaasMetadataInPath(val, keys[1:])
			if r {
				v[k] = nv
				removed = true
			}
		}
		return v, removed
	}
	return node, false
}

// removePaasMetadataInMap 删除同一层级 map 中属于 common.PaasMetadataKeys 的 key，返回是否真的删到了
func removePaasMetadataInMap(node interface{}) (interface{}, bool) {
	switch v := node.(type) {
	case yaml.MapSlice:
		out := make(yaml.MapSlice, 0, len(v))
		removed := false
		for _, item := range v {
			k, ok := item.Key.(string)
			if ok && stringx.StringInSlice(k, common.PaasMetadataKeys) {
				removed = true
				continue
			}
			out = append(out, item)
		}
		// 删空后改用普通 map：空 MapSlice 会被 marshal 成 `annotations: []`，K8s 会拒绝
		if removed && len(out) == 0 {
			return map[string]interface{}{}, true
		}
		return out, removed
	case map[interface{}]interface{}:
		removed := false
		for k := range v {
			ks, ok := k.(string)
			if ok && stringx.StringInSlice(ks, common.PaasMetadataKeys) {
				delete(v, k)
				removed = true
			}
		}
		if removed && len(v) == 0 {
			return map[string]interface{}{}, true
		}
		return v, removed
	}
	return node, false
}
