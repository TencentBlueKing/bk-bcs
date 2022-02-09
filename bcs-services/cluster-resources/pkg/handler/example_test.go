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

package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestGetK8SResTemplate(t *testing.T) {
	crh := NewClusterResourcesHandler()
	ctx := context.TODO()

	for _, kind := range example.HasDemoManifestResKinds {
		req, resp := clusterRes.GetK8SResTemplateReq{Kind: kind}, clusterRes.CommonResp{}
		err := crh.GetK8SResTemplate(ctx, &req, &resp)
		assert.Nil(t, err)
		assert.Equal(t, kind, resp.Data.AsMap()["kind"])
	}
}
