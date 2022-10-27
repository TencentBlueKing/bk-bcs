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

package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestGetResSelectItems(t *testing.T) {
	hdlr := New()
	ctx := handler.NewInjectedContext("", "", "")

	req, resp := clusterRes.GetResSelectItemsReq{Kind: resCsts.Deploy}, clusterRes.CommonResp{}
	err := hdlr.GetResSelectItems(ctx, &req, &resp)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "需要指定命名空间")

	req.Namespace = envs.TestNamespace
	err = hdlr.GetResSelectItems(ctx, &req, &resp)
	assert.Nil(t, err)

	req.Kind = resCsts.Po
	err = hdlr.GetResSelectItems(ctx, &req, &resp)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "当前资源类型 Pod 不受支持")
}
