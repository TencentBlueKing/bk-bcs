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

package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	// ErrCodeOK is error code for render which render success.
	ErrCodeOK = 0

	// ErrCodeLoadStdinFailed is error code for render which load stdin failed.
	ErrCodeLoadStdinFailed = 1

	// ErrCodeSetStdoutFailed is error code for render which set stdout failed.
	ErrCodeSetStdoutFailed = 2

	// ErrCodeRenderFailed is error code for render which redner template failed.
	ErrCodeRenderFailed = 3

	// ErrCodeParseTemplateFailed is error code for render which parse template failed.
	ErrCodeParseTemplateFailed = 4
)

const (
	// EnginePluginGolang is engine plugin name of golang template.
	EnginePluginGolang = "gotemplate"

	// EnginePluginMako is engine plugin name of mako template.
	EnginePluginMako = "makotemplate"
)

// RenderInConf stdin for renderer.
type RenderInConf struct {
	// Template encoded template content
	Template string `json:"template"`

	// Vars is template render vars.
	Vars interface{} `json:"vars"`
}

// RenderOutConf stdout for renderer.
type RenderOutConf struct {
	// Code error code.
	Code int32 `json:"code"`

	// Message error message.
	Message string `json:"message"`

	// Content is stdout for template render.
	Content string `json:"content"`
}

// Engine render engine object.
type Engine struct {
	// PluginDir is template engine plugin dir.
	PluginDir string
	plugin    string
}

// NewEngine create new Engine instance.
func NewEngine(pluginDir string) (*Engine, error) {
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin dir %s does not existed", pluginDir)
	}
	return &Engine{PluginDir: pluginDir}, nil
}

// ValidatePlugin validates target plugin.
func (e *Engine) ValidatePlugin(plugin string) error {
	path := filepath.Join(e.PluginDir, plugin)
	f, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !f.Mode().IsRegular() {
		return fmt.Errorf("%s is not regular file", path)
	}
	return nil
}

// Execute execute rendering action.
func (e *Engine) Execute(in *RenderInConf, plugin string) (*RenderOutConf, error) {
	stdin, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("encode render in conf %+v failed, err %+v", in, err)
	}

	var stderr io.Writer
	stdout := &bytes.Buffer{}

	c := exec.Command(filepath.Join(e.PluginDir, plugin))
	c.Stdin = bytes.NewBuffer(stdin)
	c.Stdout = stdout
	c.Stderr = stderr

	if err := c.Run(); err != nil {
		return nil, fmt.Errorf("execute command failed, err %+v", err)
	}
	outputBytes := stdout.Bytes()

	out := &RenderOutConf{}
	if err := json.Unmarshal(outputBytes, out); err != nil {
		return nil, fmt.Errorf("decode output %s failed, err %+v", string(outputBytes), err)
	}
	return out, nil
}
