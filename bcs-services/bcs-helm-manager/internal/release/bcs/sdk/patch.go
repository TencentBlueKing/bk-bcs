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

	goyaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
	cmdtpl "github.com/vmware-tanzu/carvel-ytt/pkg/cmd/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/cmd/ui"
	"github.com/vmware-tanzu/carvel-ytt/pkg/files"
	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
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
	manifest, err := p.do(renderedManifests)
	if err != nil {
		return nil, err
	}
	// ??????????????????
	splitManifests := stringx.SplitManifests(manifest.String())
	if err != nil {
		return nil, fmt.Errorf("SplitYAML error, %s", err.Error())
	}
	var yList []string
	for _, s := range splitManifests {
		// ??? metadata.labels ????????? `io.tencent.bcs.controller.name`
		// ??? spec.template.metadata.labels ????????? `io.tencent.bcs.controller.name`
		j, err := inject4MetadataLabels(s)
		if err != nil {
			blog.Errorf("inject4MetadataLabels error, %s", err.Error())
			yList = append(yList, s)
			continue
		}
		yList = append(yList, strings.TrimRight(j, "\n"))
	}
	yl := stringx.JoinStringBySeparator(yList, "", false)
	// ????????????
	yl += "\n"
	// ????????????
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

// inject4MetadataLabels ???????????????????????????metadata??????label
func inject4MetadataLabels(manifest string) (string, error) {
	// ?????????????????????, goyaml ??????????????? |2- ??????????????????
	var n yaml.MapSlice
	if err := yaml.Unmarshal([]byte(manifest), &n); err != nil {
		return manifest, err
	}
	out, err := yaml.Marshal(&n)
	if err != nil {
		return manifest, err
	}
	s := string(out)

	// ????????????????????????????????? key:val
	kinds := []string{"Deployment", "StatefulSet", "Job", "DaemonSet"}

	// parse name
	namePath, err := goyaml.PathString("$.metadata.name")
	if err != nil {
		return s, err
	}
	var name string
	if err := namePath.Read(strings.NewReader(s), &name); err != nil {
		return s, err
	}

	// parse kind
	kindPath, err := goyaml.PathString("$.kind")
	if err != nil {
		return s, err
	}
	var kind string
	if err := kindPath.Read(strings.NewReader(s), &kind); err != nil {
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
