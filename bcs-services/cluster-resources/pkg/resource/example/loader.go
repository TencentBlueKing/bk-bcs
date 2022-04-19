/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package example

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
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
		res.Deploy, res.STS, res.DS, res.CJ, res.Job, res.Po, res.Ing, res.SVC,
		res.EP, res.CM, res.Secret, res.PV, res.PVC, res.SC, res.HPA, res.SA, res.CObj,
	}
)

const (
	// RandomSuffixLength 资源名称随机后缀长度
	RandomSuffixLength = 8
	// SuffixCharset 后缀可选字符集（小写 + 数字）
	SuffixCharset = "abcdefghijklmnopqrstuvwxyz1234567890"
)

// LoadResConf 加载指定资源类型模板配置信息
func LoadResConf(kind string) (map[string]interface{}, error) {
	filepath := fmt.Sprintf("%s/%s.json", ResConfDIR, kind)
	conf := map[string]interface{}{}

	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return conf, err
	}
	err = json.Unmarshal(content, &conf)
	return conf, err
}

// LoadResRefs 加载指定资源类型的参考资料（Markdown 格式字符串）
func LoadResRefs(kind string) (string, error) {
	filepath := fmt.Sprintf("%s/%s.md", ResRefsDIR, kind)
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// LoadDemoManifest 加载指定资源类型的 Demo 配置信息
func LoadDemoManifest(path, namespace string) (map[string]interface{}, error) {
	filepath := fmt.Sprintf("%s/%s.yaml", ResDemoManifestDIR, path)
	manifest := map[string]interface{}{}

	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return manifest, err
	}
	err = yaml.Unmarshal(content, manifest)
	if err != nil {
		return manifest, err
	}

	// 避免名称重复，每次默认添加随机后缀
	randSuffix := stringx.Rand(RandomSuffixLength, SuffixCharset)
	rawName := mapx.Get(manifest, "metadata.name", "")
	if err = mapx.SetItems(manifest, "metadata.name", fmt.Sprintf("%s-%s", rawName, randSuffix)); err != nil {
		return manifest, err
	}

	// 若指定命名空间，且原示例配置中有命名空间的，则覆盖命名空间
	if _, err = mapx.GetItems(manifest, "metadata.namespace"); err == nil && namespace != "" {
		_ = mapx.SetItems(manifest, "metadata.namespace", namespace)
	}
	return manifest, err
}
