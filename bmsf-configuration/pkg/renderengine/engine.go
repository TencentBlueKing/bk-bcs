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

package renderengine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"bk-bscp/pkg/logger"
)

// Engine render engine object
type Engine struct {
	PluginDir string
	plugin    string
}

// NewEngine create new engine
func NewEngine(pluginDir string) (*Engine, error) {
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("pluginDir %s does not existed", pluginDir)
	}
	return &Engine{
		PluginDir: pluginDir,
	}, nil
}

// FindPlugin find plugin
func (e *Engine) FindPlugin(plugin string) error {
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

// Execute execute rendering
func (e *Engine) Execute(in *RenderInConf, plugin string) (*RenderOutConf, error) {
	stdout := &bytes.Buffer{}
	var stderr io.Writer
	stdin, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("encode render in conf %+v failed, err %s", in, err.Error())
	}
	logger.V(2).Infof("execute command with stdin %s", string(stdin))
	c := exec.Command(filepath.Join(e.PluginDir, plugin))
	c.Stdin = bytes.NewBuffer(stdin)
	c.Stdout = stdout
	c.Stderr = stderr
	if err := c.Run(); err != nil {
		return nil, fmt.Errorf("execute command failed, err %s", err.Error())
	}
	outputBytes := stdout.Bytes()
	logger.V(2).Infof("execute command with stdout %s", string(outputBytes))

	out := &RenderOutConf{}
	if err := json.Unmarshal(outputBytes, out); err != nil {
		return nil, fmt.Errorf("decode output %s failed, err %s", string(outputBytes), err.Error())
	}
	return out, nil
}
