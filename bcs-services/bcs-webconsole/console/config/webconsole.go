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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	k8sVersion "k8s.io/apimachinery/pkg/version"
)

const (
	// InternalMode xxx
	InternalMode = "internal" // 用户自己集群 inCluster 模式
	// ExternalMode xxx
	ExternalMode = "external" // 平台集群, 外部模式, 需要设置 AdminClusterId

	// kubectld 默认资源配置, 可通过配置文件覆盖
	defaultLimitCPU      = "500m"
	defaultLimitMemory   = "128Mi"
	defaultRequestCPU    = "200m"
	defaultRequestMemory = "64Mi"
)

// ResSpec 资源类型
type ResSpec struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

// ResourceConf 资源配置
type ResourceConf struct {
	Limits   *ResSpec `yaml:"limits"`
	Requests *ResSpec `yaml:"requests"`
}

// WebConsoleConf webconsole 配置
type WebConsoleConf struct {
	AdminClusterId      string                  `yaml:"admin_cluster_id"`
	Mode                string                  `yaml:"mode"`               // internal , external
	GuideDocLinks       []string                `yaml:"guide_doc_links"`    // 使用文档链接
	KubectldImage       string                  `yaml:"kubectld_image"`     // 镜像路径
	KubectldTags        []string                `yaml:"kubectld_tags"`      // 镜像tags
	KubectldResources   *ResourceConf           `yaml:"kubectld_resources"` // kubectld资源限制
	KubectldTagPatterns []*Version              `yaml:"-"`                  // 镜像解析后的版本
	Resources           v1.ResourceRequirements `yaml:"-"`
}

// Version kubectld 版本
type Version struct {
	Tag      string
	MajorVer *version.Version
}

// String xxx
func (v *Version) String() string {
	return fmt.Sprintf("%s<%s>", v.Tag, v.MajorVer.String())
}

// Init xxx
func (c *WebConsoleConf) Init() error {
	// only for development
	c.KubectldImage = ""
	c.AdminClusterId = ""
	c.Mode = InternalMode
	c.KubectldTags = []string{}
	c.GuideDocLinks = []string{}
	c.KubectldResources = &ResourceConf{
		Limits:   &ResSpec{CPU: defaultLimitCPU, Memory: defaultLimitMemory},
		Requests: &ResSpec{CPU: defaultRequestCPU, Memory: defaultRequestMemory},
	}
	c.Resources = v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.Quantity{},
			v1.ResourceMemory: resource.Quantity{},
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.Quantity{},
			v1.ResourceMemory: resource.Quantity{},
		},
	}
	return c.parseRes()
}

// parseRes 解析resources值
func (c *WebConsoleConf) parseRes() error {
	limitCPU, err := resource.ParseQuantity(c.KubectldResources.Limits.CPU)
	if err != nil {
		return errors.Wrap(err, "parse cpu limit")
	}
	if limitCPU.CmpInt64(0) != 1 {
		return errors.New("cpu limit must > 0")
	}
	c.Resources.Limits[v1.ResourceCPU] = limitCPU

	limitMemory, err := resource.ParseQuantity(c.KubectldResources.Limits.Memory)
	if err != nil {
		return errors.Wrap(err, "parse memory limit")
	}
	if limitMemory.CmpInt64(0) != 1 {
		return errors.New("memory limit must > 0")
	}
	c.Resources.Limits[v1.ResourceMemory] = limitMemory

	requestCPU, err := resource.ParseQuantity(c.KubectldResources.Requests.CPU)
	if err != nil {
		return errors.Wrap(err, "parse request cpu")
	}
	if requestCPU.CmpInt64(0) != 1 {
		return errors.New("request cpu must > 0")
	}
	c.Resources.Requests[v1.ResourceCPU] = requestCPU

	requestMemory, err := resource.ParseQuantity(c.KubectldResources.Requests.Memory)
	if err != nil {
		return errors.Wrap(err, "parse request memory")
	}
	if requestMemory.CmpInt64(0) != 1 {
		return errors.New("request memory must > 0")
	}
	c.Resources.Requests[v1.ResourceMemory] = requestMemory

	return nil
}

// IsExternalMode kubectl 是否使用外部集群
func (c *WebConsoleConf) IsExternalMode() bool {
	if c.AdminClusterId == "" {
		return false
	}
	return c.Mode == ExternalMode
}

// InitTagPatterns 初始化 tag
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

// parseVersion 解析版本, 优先使用gitVersion, 回退到 v{Major}.{Minor}.0
func parseVersion(versionInfo *k8sVersion.Info) (*version.Version, error) {
	v, err := version.NewVersion(versionInfo.GitVersion)
	if err == nil {
		return v, nil
	}

	// 回退使用 v{Major}.{Minor}.0
	majorVersion := fmt.Sprintf("v%s.%s.0", versionInfo.Major, versionInfo.Minor)
	v, err = version.NewVersion(majorVersion)
	if err == nil {
		return v, nil
	}

	return nil, errors.Errorf("Malformed version: GitVersion %s, MajorVersion %s", versionInfo.GitVersion, majorVersion)
}

// MatchTag 匹配镜像Tag
func (c *WebConsoleConf) MatchTag(versionInfo *k8sVersion.Info) (string, error) {
	v, err := parseVersion(versionInfo)
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
