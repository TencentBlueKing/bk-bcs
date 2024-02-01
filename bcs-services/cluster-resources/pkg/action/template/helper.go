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

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
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

func buildChart(templates []*entity.TemplateVersion, req *clusterRes.CreateTemplateSetReq, creator string) *chart.Chart {
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
		cht.Templates = append(cht.Files, &chart.File{
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
