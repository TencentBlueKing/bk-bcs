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
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"

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
		if match[1] == "" {
			continue
		}
		vars = append(vars, strings.TrimSpace(match[1]))
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

// patchTemplateAnnotations patch template annotations
func patchTemplateAnnotations(
	manifest map[string]interface{}, username, templateName, templateVersion string) map[string]interface{} {
	annos := mapx.GetMap(manifest, "metadata.annotations")
	if len(annos) == 0 {
		_ = mapx.SetItems(manifest, "metadata.annotations", map[string]interface{}{})
	}
	if mapx.GetStr(manifest, []string{"metadata", "annotations", resCsts.CreatorAnnoKey}) == "" {
		_ = mapx.SetItems(manifest, []string{"metadata", "annotations", resCsts.CreatorAnnoKey}, username)
	}
	if mapx.GetStr(manifest, []string{"metadata", "annotations", resCsts.UpdaterAnnoKey}) == "" {
		_ = mapx.SetItems(manifest, []string{"metadata", "annotations", resCsts.UpdaterAnnoKey}, username)
	}
	_ = mapx.SetItems(manifest, []string{"metadata", "annotations", resCsts.TemplateSourceType}, "template")
	_ = mapx.SetItems(manifest, []string{"metadata", "annotations", resCsts.TemplateNameAnnoKey}, templateName)
	_ = mapx.SetItems(manifest, []string{"metadata", "annotations", resCsts.TemplateVersionAnnoKey}, templateVersion)
	return manifest
}

// convertManifestToString convert manifest to string
func convertManifestToString(manifests []map[string]interface{}) (interface{}, error) {
	result := make([]interface{}, 0)
	for _, v := range manifests {
		name := mapx.GetStr(v, "metadata.name")
		kind := mapx.GetStr(v, "kind")
		d, err := yaml.Marshal(v)
		if err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"name":    name,
			"kind":    kind,
			"content": string(d),
		})
	}
	return result, nil
}

func isNSRequired(kind string) bool {
	return !slice.StringInSlice(kind, []string{resCsts.PV, resCsts.SC, resCsts.ClusterRole, resCsts.ClusterRoleBinding})
}
