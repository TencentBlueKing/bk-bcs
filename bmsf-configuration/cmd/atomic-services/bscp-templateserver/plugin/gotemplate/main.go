/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"bk-bscp/cmd/atomic-services/bscp-templateserver/plugin"
)

type renderer struct {
	input  *plugin.RenderInConf
	output *plugin.RenderOutConf
}

func newRenderer(input *plugin.RenderInConf) *renderer {
	return &renderer{input: input, output: &plugin.RenderOutConf{Code: plugin.ErrCodeOK, Message: "OK"}}
}

func (r *renderer) render() {
	t, err := template.New("").Parse(r.input.Template)
	if err != nil {
		r.output.Code = plugin.ErrCodeParseTemplateFailed
		r.output.Message = err.Error()
		return
	}

	// final config content size may over the limit, don't block it here.
	buffer := bytes.NewBuffer(nil)

	// rendering template.
	if err := t.Execute(buffer, r.input.Vars); err != nil {
		r.output.Code = plugin.ErrCodeRenderFailed
		r.output.Message = err.Error()
		return
	}

	r.output.Code = plugin.ErrCodeOK
	r.output.Message = "OK"
	r.output.Content = buffer.String()
}

// LoadStdin load stdin for template plugin.
func LoadStdin() (*plugin.RenderInConf, error) {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("read stdin failed, err %s", err.Error())
	}

	conf := &plugin.RenderInConf{}
	if err := json.Unmarshal(bytes, conf); err != nil {
		return nil, fmt.Errorf("decode stdin to json failed, err %s", err.Error())
	}
	return conf, nil
}

// SetStdout set output in stdout after template render.
func SetStdout(code int32, msg, content string) {
	out := &plugin.RenderOutConf{
		Code:    code,
		Message: msg,
		Content: content,
	}

	bytes, err := json.Marshal(out)
	if err != nil {
		fmt.Printf("{\"code\": %d, \"message\":\"encoding output failed, %+v\"}",
			plugin.ErrCodeSetStdoutFailed, err.Error())
	}
	fmt.Printf("%s", string(bytes))
}

func main() {
	in, err := LoadStdin()
	if err != nil {
		SetStdout(plugin.ErrCodeLoadStdinFailed, err.Error(), "")
		os.Exit(0)
	}

	renderer := newRenderer(in)
	renderer.render()
	SetStdout(renderer.output.Code, renderer.output.Message, renderer.output.Content)
}
