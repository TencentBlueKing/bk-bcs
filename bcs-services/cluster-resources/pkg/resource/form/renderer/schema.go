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

package renderer

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/feature"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/validator"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

const (
	// RandomSuffixLength 资源名称随机后缀长度
	RandomSuffixLength = 8
	// SuffixCharset 后缀可选字符集（小写 + 数字）
	SuffixCharset = "abcdefghijklmnopqrstuvwxyz1234567890"
)

// SchemaRenderer 渲染并加载表单 Schema 模板
type SchemaRenderer struct {
	ctx            context.Context
	clusterID      string
	apiVersion     string
	kind           string
	values         map[string]interface{}
	isTemplateFile bool
}

// NewSchemaRenderer xxx
func NewSchemaRenderer(ctx context.Context, clusterID, apiVersion, kind, namespace, action string,
	isTemplateFile bool) *SchemaRenderer {
	// 若没有指定命名空间，则使用 default
	if namespace == "" {
		namespace = "default"
	}
	// 避免名称重复，每次默认添加随机后缀
	randSuffix := stringx.Rand(RandomSuffixLength, SuffixCharset)
	resName := fmt.Sprintf("%s-%s", strings.ToLower(kind), randSuffix)
	selectorLabel := genWorkloadSelectorLabel(resName)
	// 尝试从 context 中获取集群类型，若获取失败，则默认独立集群
	clusterType := cluster.ClusterTypeSingle
	if clusterInfo, err := cluster.FromContext(ctx); err == nil && clusterInfo != nil {
		clusterType = clusterInfo.Type
	}

	return &SchemaRenderer{
		ctx:            ctx,
		clusterID:      clusterID,
		apiVersion:     apiVersion,
		kind:           kind,
		isTemplateFile: isTemplateFile,
		values: map[string]interface{}{
			"kind":          kind,
			"namespace":     namespace,
			"resName":       resName,
			"selectorLabel": selectorLabel,
			"lang":          i18n.GetLangFromContext(ctx),
			"action":        action,
			// 集群类型：目前可选值有 Single 独立集群，Shared 共享集群
			"clusterType": clusterType,
		},
	}
}

// Render 将模板渲染成 Schema
func (r *SchemaRenderer) Render() (map[string]interface{}, error) {
	var err error
	// 1. 检查指定资源类型是否支持表单化
	supportedAPIVersions, ok := validator.FormSupportedResAPIVersion[r.kind]
	if !ok {
		return nil, errorx.New(errcode.Unsupported, i18n.GetMsg(r.ctx, "资源类型 `%s` 不支持表单化"), r.kind)
	}

	if r.clusterID != "" {
		// 2. 预设 apiVersion，默认值为集群该类型资源的 PreferredVersion，如果获取不到且不是支持表单化的自定义资源，则抛出错误
		apiVersion, ierr := res.GetResPreferredVersion(r.ctx, r.clusterID, r.kind)
		if ierr != nil && !validator.IsFormSupportedCObjKinds(r.kind) {
			return nil, ierr
		}
		// 若 PreferredVersion 不支持表单化，则渲染为支持表单化的首个 apiVersion
		if !slice.StringInSlice(apiVersion, supportedAPIVersions) {
			apiVersion = supportedAPIVersions[0]
		}
		r.values["apiVersion"] = apiVersion

		// 3. 填充特性门控信息
		serverVerInfo, ierr := res.GetServerVersion(r.ctx, r.clusterID)
		if ierr != nil {
			return nil, ierr
		}
		r.values["featureGates"] = feature.GenFeatureGates(serverVerInfo)
	}

	// 自定义 apiVersion
	if r.apiVersion != "" {
		r.values["apiVersion"] = r.apiVersion
	}

	// 表单模板 Schema 包含原始 Schema + Layout 信息，两者格式不同，因此分别加载
	schema := map[string]interface{}{}
	if err = r.renderSubTypeTmpl2Map("schema", &schema); err != nil {
		return nil, err
	}

	var layout []interface{}
	if err = r.renderSubTypeTmpl2Map("layout", &layout); err != nil {
		return nil, err
	}

	return map[string]interface{}{"apiVersion": r.values["apiVersion"], "kind": r.kind, "schema": schema,
		"layout": layout, "rules": genSchemaRules(r.ctx)}, nil
}

func (r *SchemaRenderer) renderSubTypeTmpl2Map(dir string, ret interface{}) error {
	// 1. 加载模板并初始化
	if r.isTemplateFile {
		dir = "filetmpl/" + dir
	}
	tmpl, err := initTemplate(fmt.Sprintf("%s/%s/", envs.FormTmplFileBaseDir, dir), "*")
	if err != nil {
		return errorx.New(errcode.General, i18n.GetMsg(r.ctx, "加载模板失败：%v"), err)
	}

	// 2. 渲染模板并转换成 Map 格式（模板名称格式：{r.kind}.yaml）
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, r.kind+".yaml", r.values)
	if err != nil {
		return errorx.New(errcode.General, i18n.GetMsg(r.ctx, "渲染模板失败：%v"), err)
	}
	return yaml.Unmarshal(buf.Bytes(), ret)
}

// genSchemaRules 生成 JsonSchema 校验规则
// nolint lll
func genSchemaRules(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		"required": map[string]interface{}{
			"validator": "{{ $self.value != '' }}",
			"message":   i18n.GetMsg(ctx, "值不能为空"),
		},
		"nameRegex": map[string]interface{}{
			"validator": "/^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/",
			"message":   i18n.GetMsg(ctx, "仅支持小写字母，数字及 '-' 且需以字母数字开头和结尾"),
		},
		"nameRegexWithVar": map[string]interface{}{
			"validator": "/^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*|{{.*}})$/",
			"message":   i18n.GetMsg(ctx, "仅支持小写字母，数字及 '-' 且需以字母数字开头和结尾"),
		},
		"numberRegex": map[string]interface{}{
			"validator": "/^[0-9]+(\\.[0-9])?[0-9]*$/",
			"message":   i18n.GetMsg(ctx, "仅可包含数字字符与小数点"),
		},
		"maxLength64": map[string]interface{}{
			"validator": "{{ $self.value.length < 64 }}",
			"message":   i18n.GetMsg(ctx, "超过长度限制（64）"),
		},
		"maxLength128": map[string]interface{}{
			"validator": "{{ $self.value.length < 128 }}",
			"message":   i18n.GetMsg(ctx, "超过长度限制（128）"),
		},
		"maxLength250": map[string]interface{}{
			"validator": "{{ $self.value.length < 250 }}",
			"message":   i18n.GetMsg(ctx, "超过长度限制（250）"),
		},
		// 规则：https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
		"labelKeyRegex": map[string]interface{}{
			"validator": "/^[a-z0-9A-Z]([-_a-z0-9A-Z]*[a-z0-9A-Z])?((\\.|\\/)[a-z0-9A-Z]([-_a-z0-9A-Z]*[a-z0-9A-Z])?)*$/",
			"message":   i18n.GetMsg(ctx, "仅支持字母，数字，'-'，'_'，'.' 及 '/' 且需以字母数字开头和结尾"),
		},
		"labelKeyRegexWithVar": map[string]interface{}{
			"validator": "/^([a-z0-9A-Z]([-_a-z0-9A-Z]*[a-z0-9A-Z])?((\\.|\\/)[a-z0-9A-Z]([-_a-z0-9A-Z]*[a-z0-9A-Z])?)*|{{.*}})$/",
			"message":   i18n.GetMsg(ctx, "仅支持字母，数字，'-'，'_'，'.' 及 '/' 且需以字母数字开头和结尾"),
		},
		// NOTE 标签值允许为空
		"labelValRegex": map[string]interface{}{
			"validator": "/^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$/",
			"message":   i18n.GetMsg(ctx, "需以字母数字开头和结尾，可包含 '-'，'_'，'.' 和字母数字"),
		},
		"labelValRegexWithVar": map[string]interface{}{
			"validator": "/^((([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?|{{.*}})$/",
			"message":   i18n.GetMsg(ctx, "需以字母数字开头和结尾，可包含 '-'，'_'，'.' 和字母数字"),
		},
		"sliceLength1": map[string]interface{}{
			"validator": "{{ $self.value.length > 0 }}",
			"message":   i18n.GetMsg(ctx, "不应少于 1 个项"),
		},
		"scheduleValid": map[string]interface{}{
			"validator": `/^(\*|([0-5]?\d)(,[0-5]?\d)*|([0-5]?\d)-([0-5]?\d))(\/([0-5]?\d)(,[0-5]?\d)*)?\s+` +
				`(\*|([01]?\d|2[0-3])(,[01]?\d|2[0-3])*|([01]?\d|2[0-3])-` +
				`([01]?\d|2[0-3]))(\/([01]?\d|2[0-3])(,[01]?\d|2[0-3])*)?\s+` +
				`(\*|(0?[1-9]|[12][0-9]|3[01])(,(0?[1-9]|[12][0-9]|3[01]))*|` +
				`(0?[1-9]|[12][0-9]|3[01])-(0?[1-9]|[12][0-9]|3[01])|\?)` +
				`(\/(0?[1-9]|[12][0-9]|3[01])(,(0?[1-9]|[12][0-9]|3[01]))*)?\s+` +
				`(\*|(0?[1-9]|1[0-2]|JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC)` +
				`(,(0?[1-9]|1[0-2]|JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC))*|` +
				`(0?[1-9]|1[0-2]|JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC)-` +
				`(0?[1-9]|1[0-2]|JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC))` +
				`(\/((0?[1-9]|1[0-2])(,((0?[1-9]|1[0-2]|JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC)))*))?\s+` +
				`(\*|([0-6]|SUN|MON|TUE|WED|THU|FRI|SAT)(,[0-6]|SUN|MON|TUE|WED|THU|FRI|SAT)*|` +
				`([0-6]|SUN|MON|TUE|WED|THU|FRI|SAT)-([0-6]|SUN|MON|TUE|WED|THU|FRI|SAT)|\?)` +
				`(\/([0-6](,([0-6]|SUN|MON|TUE|WED|THU|FRI|SAT))*))?$/`,
			"message": i18n.GetMsg(ctx, "格式无效，请参考 Cron 格式"),
		},
	}
}
