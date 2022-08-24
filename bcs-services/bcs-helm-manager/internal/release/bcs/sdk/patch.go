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

package sdk

import (
	"bytes"
	"fmt"
	"strings"

	cmdtpl "github.com/vmware-tanzu/carvel-ytt/pkg/cmd/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/cmd/ui"
	"github.com/vmware-tanzu/carvel-ytt/pkg/files"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/mapx"
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

func newPatcher(templates []*release.File, keys map[string]string) *patcher {
	fs := make([]*files.File, 0, 5)
	for _, f := range templates {
		fs = append(fs, files.MustNewFileFromSource(files.NewBytesSource(f.Name, replacePatchTplKey(keys, f.Content))))
	}

	return &patcher{
		files: fs,
	}
}

type patcher struct {
	files []*files.File
}

// Run implements the post-render Run method, do the render
func (p *patcher) Run(renderedManifests *bytes.Buffer) (*bytes.Buffer, error) {
	// 处理 yaml 转换 json，添加指定的 key 和 value
	splitedStrArr := stringx.SplitYaml2Array(renderedManifests.String(), "")
	var yList []string
	for _, s := range splitedStrArr {
		j, err := stringx.Yaml2Json(s)
		if err != nil {
			return nil, err
		}
		// 向 metadata.labels 中注入 `io.tencent.bcs.controller.name`
		// 向 spec.template.metadata.labels 中注入 `io.tencent.bcs.controller.name`
		j = inject4MetadataLabels(j)
		y, err := stringx.Json2Yaml(j)
		if err != nil {
			return nil, err
		}
		yList = append(yList, string(y))
	}
	yl := stringx.JoinStringBySeparator(yList, "", true)
	// 写回数据
	buf := new(bytes.Buffer)
	buf.WriteString(yl)
	return p.do(buf)
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
		return nil, fmt.Errorf("no data output from patcher")
	}
	return bytes.NewBuffer(out.Files[0].Bytes()), nil
}

// 兼容逻辑，目的是向metadata注入label
func inject4MetadataLabels(j map[interface{}]interface{}) map[interface{}]interface{} {
	// 限制下面几个注入指定的 key:val
	kinds := []string{"Deployment", "DaemonSet", "Job", "DaemonSet"}
	// 允许metadata中label注入的资源类型
	for _, kind := range kinds {
		if j["kind"] != kind {
			continue
		}
		// 断言格式
		name, _ := mapx.GetItems(j, []string{"metadata", "name"})
		mapx.SetItems(j, []string{"metadata", "labels", labelKey}, name)
		mapx.SetItems(j, []string{"spec", "template", "metadata", "labels", labelKey}, name)
	}
	if j["kind"] == "Service" {
		name, _ := mapx.GetItems(j, []string{"metadata", "name"})
		mapx.SetItems(j, []string{"metadata", "labels", labelKey}, name)
	}
	return j
}
