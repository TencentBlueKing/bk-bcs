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

package template

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

const (
	templateFileVarPattern = `{{\s*([^}]*)\s*}}`
)

func toEntityTemplateIDs(templateIDs []*clusterRes.TemplateID) []entity.TemplateID {
	ids := make([]entity.TemplateID, 0, len(templateIDs))
	for _, id := range templateIDs {
		ids = append(ids, entity.TemplateID{
			TemplateSpace:   id.TemplateSpace,
			TemplateName:    id.TemplateName,
			TemplateVersion: id.Version,
		})
	}
	return ids
}

func buildChart(templates []*entity.TemplateVersion, req *clusterRes.CreateTemplateSetReq,
	creator string) *chart.Chart {
	cht := &chart.Chart{
		Metadata: &chart.Metadata{
			APIVersion:  "v2",
			Name:        req.GetName(),
			Version:     req.GetVersion(),
			AppVersion:  req.GetVersion(),
			Description: req.GetDescription(),
			Keywords:    req.GetKeywords(),
			Annotations: map[string]string{
				"category":          req.GetCategory(),
				"creator":           creator,
				"bcs_template_sets": buildTemplateSetsAnnotation(req.GetTemplates()),
			},
		},
	}
	cht.Raw = append(cht.Raw, &chart.File{
		Name: chartutil.ValuesfileName,
		Data: []byte(req.GetValues()),
	})
	for _, template := range templates {
		cht.Templates = append(cht.Templates, &chart.File{
			Name: fmt.Sprintf("templates/%s", template.TemplateName),
			Data: []byte(template.Content),
		})
	}
	return cht
}

func buildTemplateSetsAnnotation(templates []*clusterRes.TemplateID) string {
	b, _ := json.Marshal(templates)
	return string(b)
}

// parseMultiTemplateFileVar parse template file variables from multiple templates
func parseMultiTemplateFileVar(templates []entity.TemplateDeploy) []string {
	vars := make([]string, 0)
	for _, template := range templates {
		vars = append(vars, parseTemplateFileVar(template.Content)...)
	}
	return vars
}

// parseTemplateFileVar parse template file variables
func parseTemplateFileVar(template string) []string {
	re := regexp.MustCompile(templateFileVarPattern)
	vars := make([]string, 0)
	matches := re.FindAllStringSubmatch(template, -1)
	for _, match := range matches {
		if match == nil || len(match) < 2 {
			continue
		}
		if strings.HasPrefix(match[1], ".Values.") {
			vars = append(vars, strings.TrimSpace(match[1]))
		}
	}
	vars = slice.RemoveDuplicateValues(vars)
	return vars
}

// replaceTemplateFileVar replace template file variables
func replaceTemplateFileVar(template string, values map[string]string) string {
	re := regexp.MustCompile(templateFileVarPattern)
	return re.ReplaceAllStringFunc(template, func(s string) string {
		// 去掉 {{ 和 }}
		varName := strings.TrimSpace(s[2 : len(s)-2])
		return values[varName]
	})
}

// replaceTemplateFileToHelm replace template file to helm
func replaceTemplateFileToHelm(template string) string {
	re := regexp.MustCompile(templateFileVarPattern)
	return re.ReplaceAllStringFunc(template, func(s string) string {
		// 去掉 {{ 和 }}
		varName := strings.TrimSpace(s[2 : len(s)-2])
		if !strings.HasPrefix(varName, ".Values.") {
			if strings.HasPrefix(varName, ".") {
				varName = ".Values" + varName
			} else {
				varName = ".Values." + varName
			}
		}
		return "{{ " + varName + " }}"
	})
}

// patchTemplateAnnotations patch template annotations
func patchTemplateAnnotations(manifest map[string]interface{}, username, templateSpace, templateName,
	templateVersion string) map[string]interface{} {
	annos := mapx.GetMap(manifest, "metadata.annotations")
	if len(annos) == 0 {
		_ = mapx.SetItems(manifest, "metadata.annotations", map[string]interface{}{})
	}
	if mapx.GetStr(manifest, []string{"metadata", "annotations", resCsts.CreatorAnnoKey}) == "" {
		_ = mapx.SetItems(manifest, []string{"metadata", "annotations", resCsts.CreatorAnnoKey}, username)
	}
	_ = mapx.SetItems(manifest, []string{"metadata", "annotations", resCsts.UpdaterAnnoKey}, username)
	_ = mapx.SetItems(manifest, []string{"metadata", "annotations", resCsts.TemplateSourceType},
		resCsts.TemplateSourceTypeValue)
	_ = mapx.SetItems(manifest, []string{"metadata", "annotations", resCsts.TemplateNameAnnoKey}, fmt.Sprintf("%s/%s",
		templateSpace, templateName))
	_ = mapx.SetItems(manifest, []string{"metadata", "annotations", resCsts.TemplateVersionAnnoKey}, templateVersion)
	return manifest
}

// convertManifestToString convert manifest to string
func convertManifestToString(
	ctx context.Context, manifests []map[string]interface{}, clusterID string) (interface{}, error) {
	result := make([]interface{}, 0)
	for _, v := range manifests {
		name := mapx.GetStr(v, "metadata.name")
		kind := mapx.GetStr(v, "kind")
		d, err := yaml.Marshal(v)
		if err != nil {
			return nil, err
		}
		// 获取线上的版本
		d2 := getK8sResource(ctx, kind, name, clusterID, v)
		result = append(result, map[string]interface{}{
			"name":            name,
			"kind":            kind,
			"content":         string(d),
			"previousContent": d2,
		})
	}
	return result, nil
}

func isNSRequired(kind string) bool {
	return !slice.StringInSlice(kind, []string{resCsts.PV, resCsts.SC, resCsts.ClusterRole, resCsts.ClusterRoleBinding})
}

// 获取k8s线上资源
func getK8sResource(ctx context.Context, kind, name, clusterID string, v map[string]interface{}) string {
	// deploy templates
	clusterConf := res.NewClusterConf(clusterID)
	// 获取线上的版本
	groupVersion := mapx.GetStr(v, "apiVersion")
	namespace := mapx.GetStr(v, "metadata.namespace")
	if kind == "" {
		return ""
	}
	k8sRes, err := res.GetGroupVersionResource(ctx, clusterConf, kind, groupVersion)
	if err != nil {
		// 报错不影响 preview 接口返回
		klog.Errorf("get group version resource err: %v", err)
		return ""
	}
	content, err := cli.NewResClient(clusterConf, k8sRes).Get(ctx, namespace, name, metav1.GetOptions{})
	if err != nil {
		// 报错不影响 preview 接口返回
		klog.Errorf("new res client get err: %v", err)
		return ""
	}

	// 裁剪线上版本内容
	object := pruneResource(content.Object)

	d2, err := yaml.Marshal(object)
	if err != nil {
		// 报错不影响 preview 接口返回
		klog.Errorf("yaml marshal err: %v", err)
		return ""
	}
	return string(d2)
}

// 裁剪资源，不需要的字段去除
func pruneResource(m map[string]interface{}) map[string]interface{} {
	// 需要删除的map key字段
	pruneKey := [][]string{
		{"metadata", "generation"},
		{"metadata", "creationTimestamp"},
		{"metadata", "resourceVersion"},
		{"metadata", "uid"},
		{"metadata", "managedFields"},
		{"spec", "template", "metadata", "creationTimestamp"},
		{"spec", "template", "spec", "schedulerName"},
		{"spec", "template", "spec", "securityContext"},
		{"status"},
	}
	// 仅能删除map[string] interface{}类型key值
	// 无法对[]interface{}里面的字段进行删除
	for _, value := range pruneKey {
		lastKey := m
		isDelete := true
		for i := 0; i < len(value)-1; i++ {
			l, ok := lastKey[value[i]].(map[string]interface{})
			// 只要取不出就直接忽略这个key值
			if !ok {
				isDelete = false
				break
			}
			lastKey = l
		}
		if isDelete {
			delete(lastKey, value[len(value)-1])
		}
	}

	return m
}

// 校验chart helm语法是否正确，并自动填充Helm模板内容
func validAndFillChart(cht *chart.Chart, value string) (map[string]string, error) {
	var m map[string]interface{}
	err := yaml.Unmarshal([]byte(value), &m)
	if err != nil {
		return nil, err
	}

	var options chartutil.ReleaseOptions
	valuesToRender, err := chartutil.ToRenderValues(cht, m, options, nil)
	if err != nil {
		return nil, err
	}
	var e engine.Engine
	e.LintMode = true
	result, err := e.Render(cht, valuesToRender)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// helm 语法模式 模板文件内容进行helm template 渲染, 简单语法模式自动跳过
func renderTemplateForHelmMode(
	td []entity.TemplateDeploy, value string, variables map[string]string) (map[string]string, error) {
	cht := chart.Chart{
		Raw:       []*chart.File{},
		Metadata:  &chart.Metadata{},
		Templates: []*chart.File{},
	}

	for _, v := range td {
		// helm 模式才转
		if v.RenderMode == string(constants.HelmRenderMode) {
			// 先填充来自variable的变量
			v.Content = replaceTemplateFileVar(v.Content, variables)
			cht.Templates = append(cht.Templates, &chart.File{
				Name: path.Join(v.TemplateSpace, v.TemplateName),
				Data: []byte(v.Content),
			})
		}
	}
	return validAndFillChart(&cht, value)
}
