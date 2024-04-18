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

// Package options defines the config
package options

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// GitOpsOptions defines the gitops option
type GitOpsOptions struct {
	Server    string
	Token     string
	ProxyPath string
}

var (
	op = &GitOpsOptions{
		ProxyPath: "gitopsmanager/proxy",
	}
	cfgFilePath string
)

// ConfigfilePath return the config file path
func ConfigfilePath() string {
	return cfgFilePath
}

// GlobalOption return the global option
func GlobalOption() *GitOpsOptions {
	return op
}

// Parse the config from file
func Parse(cfgFile string) {
	cfgFilePath = cfgFile
	if fi, err := os.Stat(cfgFile); err == nil {
		if fi.IsDir() {
			blog.Fatalf("Config file '%s' cannot be directory", cfgFile)
		}
		var bs []byte
		bs, err = os.ReadFile(cfgFile)
		if err != nil {
			blog.Fatalf("Config file '%s' read failed: %s", cfgFile, err.Error())
		}
		if err = json.Unmarshal(bs, op); err != nil {
			blog.Fatalf("Config file '%s' unmarshal failed: %s", cfgFile, err.Error())
		}
		blog.V(3).Infof("Config loaded from file: %s", cfgFile)
	} else if errors.Is(err, os.ErrNotExist) {
		blog.V(3).Infof("Warning! config file '%s' not exist", cfgFile)
	} else {
		blog.Fatalf("Check config file '%s' exist failed: %s", cfgFile, err.Error())
	}
}
