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

// Package project xxx
package project

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/project"
)

var (
	projectFeatTags = []string{constants.ProjectIDTag}
)

type general struct {
	ctx              context.Context
	projectID        string
	resourceFeatList []string
	data             map[string]interface{}
}

func (g *general) getFeatures() operator.M {
	features := make(operator.M)
	features[constants.ProjectIDTag] = g.projectID
	return features
}

func (g *general) putResources() error {
	return project.PutData(g.ctx, g.data, g.getFeatures(), g.resourceFeatList)
}

// HandlerCreateProjectInfoReq  CreateProjectInfoReq业务方法
func HandlerCreateProjectInfoReq(ctx context.Context, req *storage.PutProjectInfoRequest) (operator.M, error) {

	data := map[string]interface{}{
		"projectID": req.ProjectID,
		"data":      req.Data.AsMap(),
	}
	r := &general{
		ctx:              ctx,
		projectID:        req.ProjectID,
		resourceFeatList: projectFeatTags,
		data:             data,
	}

	return data, r.putResources()
}
