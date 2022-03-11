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
 *
 */
package config

import (
	"regexp"
)

const (
	InternalMode = "internal" // 用户自己集群 inCluster 模式
	ExternalMode = "external" // 平台集群, 外部模式, 需要设置 AdminClusterId
)

type WebConsoleConf struct {
	AdminClusterId          string                      `yaml:"admin_cluster_id"`
	Mode                    string                      `yaml:"mode"`               // internal , external
	KubectldImage           string                      `yaml:"kubectld_image"`     // 镜像路径
	KubectldTagMatch        map[string][]string         `yaml:"kubectld_tag_match"` // 镜像Tag对应关系
	KubectldTagMatchPattern map[string][]*regexp.Regexp `yaml:"-"`                  // 镜像Tag对应关系,编译后的正则
	KubectldTag             string                      `yaml:"kubectld_tag"`       // 镜像默认tag
	GuideDocLink            string                      `yaml:"guide_doc_link"`     // 使用文档链接
}

func (c *WebConsoleConf) Init() error {
	// only for development
	c.KubectldImage = ""
	c.AdminClusterId = ""
	c.Mode = InternalMode
	c.KubectldTagMatch = nil
	c.KubectldTagMatchPattern = nil
	c.KubectldTag = ""
	c.GuideDocLink = ""

	return nil
}

func (c *WebConsoleConf) InitMatchPattern() error {
	c.KubectldTagMatchPattern = map[string][]*regexp.Regexp{}
	for tag, patterns := range c.KubectldTagMatch {
		matchPatterns := []*regexp.Regexp{}
		for _, pattern := range patterns {
			p, err := regexp.Compile(pattern)
			if err != nil {
				return err
			}
			matchPatterns = append(matchPatterns, p)
		}
		c.KubectldTagMatchPattern[tag] = matchPatterns
	}

	return nil
}
