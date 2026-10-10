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
	"github.com/goccy/go-yaml/ast"
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
// skipPaasAnnotations 为 true 时，会在渲染完成后删除 BCS 注入的 paas 系列元数据。
func newPatcher(templates []*release.File, keys map[string]string, skipPaasAnnotations bool) *patcher {
	fs := make([]*files.File, 0, 5)
	for _, f := range templates {
		fs = append(fs, files.MustNewFileFromSource(files.NewBytesSource(f.Name, replacePatchTplKey(keys, f.Content))))
	}

	return &patcher{
		files:               fs,
		skipPaasAnnotations: skipPaasAnnotations,
	}
}

type patcher struct {
	files []*files.File
	// skipPaasAnnotations 为 true 时删除 BCS 注入的 io.tencent.paas 系列元数据
	skipPaasAnnotations bool
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

		// 用户选择不注入时，删除 BCS 注入的 io.tencent.paas 系列元数据（annotations + labels）
		if p.skipPaasAnnotations {
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

// paasMetaPaths 需要清理 paas 元数据的位置。
/*
需要要同时处理 annotations 和 labels：实测 Deployment 的 metadata.annotations 与
metadata.labels 两处都有 io.tencent.paas.creator / updator，只清 annotations 会漏，
用户 kubectl get -o yaml 仍看得到人名。
spec.template 下的两份是防御性覆盖：实测当前模板未把这三个 key 注入到 pod template 层，
路径不存在或无命中时 removePaasMetadataByPath 自动跳过，无副作用，也不会触发 Pod 滚动。
*/
var paasMetaPaths = []string{
	"$.metadata.annotations",
	"$.metadata.labels",
	"$.spec.template.metadata.annotations",
	"$.spec.template.metadata.labels",
}

/*
// removePaasMetadata 删除 BCS 注入的 io.tencent.paas 系列元数据（annotations 与 labels）。
// 必须与"不放占位符"配合使用：仅不放占位符会让值变成空串而不是消失，
// 且会把 creator 原有值覆盖为空，造成不可逆的数据丢失。
//
// 未删到任何内容时原样返回 manifest，避免对无关资源（CRD / Secret 等）做无谓的重新格式化。
*/
func removePaasMetadata(manifest string) (string, error) {
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

	f, err := parser.ParseBytes([]byte(s), 0)
	if err != nil {
		return manifest, err
	}

	removed := false
	for _, path := range paasMetaPaths {
		ok, err := removePaasMetadataByPath(f, s, path)
		if err != nil {
			return manifest, err
		}
		if ok {
			removed = true
		}
	}

	// 与 inject4MetadataLabels 保持一致的收尾原则：
	// 没有命中就返回原始 manifest，不对无关资源做无谓的重新格式化
	if !removed {
		return manifest, nil
	}
	return f.String(), nil
}

// removePaasMetadataByPath 删除指定路径下属于 paas 系列的 key，返回是否真的删到了
func removePaasMetadataByPath(f *ast.File, s, path string) (bool, error) {
	p, err := goyaml.PathString(path)
	if err != nil {
		return false, err
	}

	items := make(map[string]string)
	// 路径不存在时直接跳过（如 ConfigMap 没有 spec.template）
	if err := p.Read(strings.NewReader(s), &items); err != nil {
		return false, nil
	}

	removed := false
	for k := range items {
		if stringx.StringInSlice(k, common.PaasMetadataKeys) {
			delete(items, k)
			removed = true
		}
	}
	if !removed {
		return false, nil
	}

	// 全部删空时写入空 map，避免出现 `annotations: []` 导致 K8s 校验失败。
	// yaml.Marshal 一个空 map 会输出 `{}`，正好是期望的形态。
	b, err := yaml.Marshal(items)
	if err != nil {
		return false, err
	}
	if err := p.ReplaceWithReader(f, bytes.NewReader(b)); err != nil {
		return false, err
	}
	return true, nil
}
