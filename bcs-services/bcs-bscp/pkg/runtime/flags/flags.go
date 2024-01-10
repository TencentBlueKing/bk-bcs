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

// Package flags NOTES
package flags

import (
	"strings"

	"github.com/spf13/pflag"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
)

// wordSepNormalizeFunc changes all flags that contain "_" separators
func wordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
	}
	return pflag.NormalizedName(name)
}

// SysFlags normalizes and parses the command line flags
func SysFlags(fs *pflag.FlagSet) *cc.SysOption {
	opt := new(cc.SysOption)
	fs.SetNormalizeFunc(wordSepNormalizeFunc)

	fs.StringArrayVarP(&opt.ConfigFiles, "config-file", "c", []string{},
		"the absolute path of the configuration file (repeatable)")
	fs.IPVarP(&opt.BindIP, "bind-ip", "b", []byte{}, "which IP the server is listen to")
	fs.BoolVarP(&opt.Versioned, "version", "v", false, "show version")

	return opt
}
