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

// Package example xxx
package example

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

var (
	// ResConfDIR 模板配置信息目录
	ResConfDIR = envs.ExampleFileBaseDir + "/config"

	// ResDemoManifestDIR Demo Manifest 目录
	ResDemoManifestDIR = envs.ExampleFileBaseDir + "/manifest"

	// ResRefsDIR 参考资料目录
	ResRefsDIR = envs.ExampleFileBaseDir + "/reference"

	// HasDemoManifestResKinds 支持获取示例的资源类型
	HasDemoManifestResKinds = []string{
		resCsts.Deploy, resCsts.STS, resCsts.DS, resCsts.CJ, resCsts.Job, resCsts.Po, resCsts.Ing, resCsts.SVC,
		resCsts.EP, resCsts.CM, resCsts.Secret, resCsts.PV, resCsts.PVC, resCsts.SC, resCsts.HPA, resCsts.SA,
		resCsts.GDeploy, resCsts.GSTS, resCsts.CObj, resCsts.BscpConfig,
	}
)

const (
	// RandomSuffixLength 资源名称随机后缀长度
	RandomSuffixLength = 8
	// SuffixCharset 后缀可选字符集（小写 + 数字）
	SuffixCharset = "abcdefghijklmnopqrstuvwxyz1234567890"
)

// LoadResConf 加载指定资源类型模板配置信息
func LoadResConf(ctx context.Context, kind string) (map[string]interface{}, error) {
	lang := i18n.GetLangFromContext(ctx)
	filepath := fmt.Sprintf("%s/%s/%s.json", ResConfDIR, lang, kind)
	conf := map[string]interface{}{}
	if strings.Contains(filepath, "..") {
		return conf, nil
	}

	content, err := os.ReadFile(filepath)
	if err != nil {
		return conf, err
	}
	err = json.Unmarshal(content, &conf)
	return conf, err
}

// LoadResRefs 加载指定资源类型的参考资料（Markdown 格式字符串）
func LoadResRefs(ctx context.Context, kind string) (string, error) {
	lang := i18n.GetLangFromContext(ctx)
	filepath := fmt.Sprintf("%s/%s/%s.md", ResRefsDIR, lang, kind)
	if strings.Contains(filepath, "..") {
		return "", nil
	}
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// LoadDemoManifest 加载指定资源类型的 Demo 配置信息
func LoadDemoManifest(
	ctx context.Context, path, clusterID, namespace, kind string,
) (map[string]interface{}, error) {
	filepath := fmt.Sprintf("%s/%s.yaml", ResDemoManifestDIR, path)
	manifest := map[string]interface{}{}
	if strings.Contains(filepath, "..") {
		return manifest, nil
	}

	content, err := os.ReadFile(filepath)
	if err != nil {
		return manifest, err
	}
	err = yaml.Unmarshal(content, manifest)
	if err != nil {
		return manifest, err
	}

	// 尝试使用 preferred 的 ApiVersion，若获取失败则保留默认的
	preferredApiVersion, err := res.GetResPreferredVersion(ctx, clusterID, kind)
	if err == nil {
		_ = mapx.SetItems(manifest, "apiVersion", preferredApiVersion)
	}

	// 避免名称重复，每次默认添加随机后缀
	randSuffix := stringx.Rand(RandomSuffixLength, SuffixCharset)
	rawName := mapx.GetStr(manifest, "metadata.name")
	if err = mapx.SetItems(manifest, "metadata.name", fmt.Sprintf("%s-%s", rawName, randSuffix)); err != nil {
		return manifest, err
	}

	// 若指定命名空间，且原示例配置中有命名空间的，则覆盖命名空间
	if _, getNSErr := mapx.GetItems(manifest, "metadata.namespace"); getNSErr == nil && namespace != "" {
		_ = mapx.SetItems(manifest, "metadata.namespace", namespace)
	}
	return manifest, err
}

// LoadDemoManifestString 加载指定资源类型的 Demo 配置信息（字符串）
func LoadDemoManifestString(
	ctx context.Context, path, clusterID, namespace, kind string,
) (string, error) {
	filepath := fmt.Sprintf("%s/%s.yaml", ResDemoManifestDIR, path)
	if strings.Contains(filepath, "..") {
		return "", nil
	}

	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
