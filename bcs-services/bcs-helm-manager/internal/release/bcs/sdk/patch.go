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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"

	cmdtpl "github.com/vmware-tanzu/carvel-ytt/pkg/cmd/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/cmd/ui"
	"github.com/vmware-tanzu/carvel-ytt/pkg/files"
)

const (
	resourceFilename = "resource.yaml"
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
	return p.do(renderedManifests)
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
