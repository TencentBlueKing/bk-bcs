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

// Package configfilecheck xxx
package configfilecheck

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"os"
	"path"
	"regexp"
	"strings"
)

// Options bcs log options
type Options struct {
	Interval   int         `json:"interval" yaml:"interval"`
	FilePaths  []string    `json:"filePaths" yaml:"filePaths"`
	CheckRules []CheckRule `json:"checkRules" yaml:"checkRules"`
}

// CheckRule 配置文件检测规则
type CheckRule struct {
	RuleName string `json:"ruleName" yaml:"ruleName"`
	FilePath string `json:"filePath" yaml:"filePath"`

	// CheckString 是要检查的字符串
	CheckString string `json:"checkString" yaml:"checkString"`

	// ShouldContain 表示是否应该包含该字符串
	ShouldContain bool `json:"shouldContain" yaml:"shouldContain"`

	// RegexPattern 是用于匹配的正则表达式
	RegexPattern string `json:"regexPattern" yaml:"regexPattern"`
	ShouldMatch  bool   `json:"shouldMatch" yaml:"shouldMatch"`

	ShouldExist bool `json:"shouldExist" yaml:"shouldExist"`
}

// Validate validate options
func (o *Options) Validate() error {
	if o.FilePaths == nil || len(o.FilePaths) == 0 {
		o.FilePaths = []string{
			"/etc/resolv.conf",
			"/etc/ntp.conf",
			"/var/spool/cron/root",
		}
	}

	if o.CheckRules == nil || len(o.CheckRules) == 0 {
		o.CheckRules = []CheckRule{
			{RuleName: "dns服务器未初始化", FilePath: "/etc/resolv.conf", RegexPattern: `(?m)^nameserver 183`, ShouldExist: true, ShouldMatch: false},
		}
	}

	o.FilePaths = removeDuplicates(o.FilePaths)
	return nil
}

// Check xxx
func (v *CheckRule) Check() (string, error) {
	filePath := path.Join(util.GetHostPath(), v.FilePath)

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		if errors.IsNotFound(err) {
			if v.ShouldExist {
				return configFileNotFoundStatus, fmt.Errorf("open file %s failed: %s", filePath, err.Error())
			} else {
				return NormalStatus, nil
			}
		}
		return otherErrorStatus, fmt.Errorf("read file %s failed: %s", filePath, err.Error())
	}

	if v.ShouldContain && v.CheckString != "" {
		if !strings.Contains(string(content), v.CheckString) {
			return configNotFoundStatus, fmt.Errorf("file %s doesn't contain: %s", filePath, v.CheckString)
		}
	}

	if v.RegexPattern != "" {
		re, _ := regexp.Compile(v.RegexPattern)
		if !re.MatchString(string(content)) && v.ShouldMatch {
			return configNotFoundStatus, fmt.Errorf("file %s doesn't match %s", filePath, v.RegexPattern)
		} else if re.MatchString(string(content)) && !v.ShouldMatch {
			return errorConfigMatchedStatus, fmt.Errorf("file %s match wrong config %s", filePath, v.RegexPattern)
		}
	}
	return NormalStatus, nil
}

func removeDuplicates(filePaths []string) []string {
	result := make([]string, 0, 0)

	uniqMap := make(map[string]string)
	for _, filePath := range filePaths {
		uniqMap[filePath] = filePath
	}

	for _, filePath := range uniqMap {
		result = append(result, filePath)
	}

	return result
}
