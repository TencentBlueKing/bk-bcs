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
	"fmt"
	"sort"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
)

const (
	InternalMode = "internal" // 用户自己集群 inCluster 模式
	ExternalMode = "external" // 平台集群, 外部模式, 需要设置 AdminClusterId
)

type WebConsoleConf struct {
	AdminClusterId      string     `yaml:"admin_cluster_id"`
	Mode                string     `yaml:"mode"`           // internal , external
	KubectldImage       string     `yaml:"kubectld_image"` // 镜像路径
	KubectldTags        []string   `yaml:"kubectld_tags"`  // 镜像tags
	KubectldTagPatterns []*Version `yaml:"-"`              // 镜像解析后的版本
	GuideDocLink        string     `yaml:"guide_doc_link"` // 使用文档链接
}

type Version struct {
	Tag      string
	MajorVer *version.Version
}

func (v *Version) String() string {
	return fmt.Sprintf("%s<%s>", v.Tag, v.MajorVer.String())
}

func (c *WebConsoleConf) Init() error {
	// only for development
	c.KubectldImage = ""
	c.AdminClusterId = ""
	c.Mode = InternalMode
	c.KubectldTags = []string{}
	c.GuideDocLink = ""

	return nil
}

func (c *WebConsoleConf) InitTagPatterns() error {
	c.KubectldTagPatterns = []*Version{}
	for _, tag := range c.KubectldTags {
		v, err := version.NewVersion(tag)
		if err != nil {
			return err
		}
		// 只使用 major 版本做匹配
		segments := v.Segments()
		v, err = version.NewSemver(fmt.Sprintf("%d.%d.0", segments[0], segments[1]))
		if err != nil {
			return err
		}
		c.KubectldTagPatterns = append(c.KubectldTagPatterns, &Version{tag, v})
	}

	// 由大到小排序
	sort.Slice(c.KubectldTagPatterns, func(i, j int) bool {
		return c.KubectldTagPatterns[i].MajorVer.GreaterThanOrEqual(c.KubectldTagPatterns[j].MajorVer)
	})

	return nil
}

func (c *WebConsoleConf) MatchTag(tag string) (string, error) {
	v, err := version.NewVersion(tag)
	if err != nil {
		return "", err
	}

	version := &Version{}
	for _, version = range c.KubectldTagPatterns {
		if version.MajorVer.LessThanOrEqual(v) {
			return version.Tag, nil
		}
	}

	// 返回最小的tag
	if version.Tag != "" {
		return version.Tag, nil
	}

	return "", errors.New("have not valid tag")
}
