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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/workload"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestFormDataRenderPreview(t *testing.T) {
	hdlr := New()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_deployment", "")
	// 类型强制转换，确保解析器正确解析
	res.ConvertInt2Int64(manifest)
	formData, _ := pbstruct.Map2pbStruct(workload.ParseDeploy(manifest))
	req, resp := clusterRes.FormRenderPreviewReq{Kind: res.Deploy, FormData: formData}, clusterRes.CommonResp{}
	err := hdlr.FormDataRenderPreview(ctx, &req, &resp)
	assert.Nil(t, err)
}
