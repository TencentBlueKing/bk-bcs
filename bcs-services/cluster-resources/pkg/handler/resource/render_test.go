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

package resource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/workload"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestFormDataRenderPreview(t *testing.T) {
	hdlr := New()
	ctx := context.WithValue(context.TODO(), ctxkey.UsernameKey, envs.AnonymousUsername)

	manifest, _ := example.LoadDemoManifest(ctx, "workload/simple_deployment", "", "", resCsts.Deploy)
	// 类型强制转换，确保解析器正确解析
	res.ConvertInt2Int64(manifest)
	formData, _ := pbstruct.Map2pbStruct(workload.ParseDeploy(manifest))
	req, resp := clusterRes.FormRenderPreviewReq{Kind: resCsts.Deploy, FormData: formData}, clusterRes.CommonResp{}
	err := hdlr.FormDataRenderPreview(ctx, &req, &resp)
	assert.Nil(t, err)
}

func TestFormToYAML(t *testing.T) {
	hdlr := New()
	ctx := context.WithValue(context.TODO(), ctxkey.UsernameKey, envs.AnonymousUsername)

	manifest, _ := example.LoadDemoManifest(ctx, "workload/simple_deployment", "", "", resCsts.Deploy)
	// 类型强制转换，确保解析器正确解析
	res.ConvertInt2Int64(manifest)
	formData, _ := pbstruct.Map2pbStruct(workload.ParseDeploy(manifest))
	resources := []*clusterRes.FormData{{Kind: resCsts.Deploy, FormData: formData}}
	req, resp := clusterRes.FormToYAMLReq{Resources: resources}, clusterRes.CommonResp{}
	err := hdlr.FormToYAML(ctx, &req, &resp)
	assert.Nil(t, err)
}

func TestYAMLToForm(t *testing.T) {
	hdlr := New()
	ctx := context.WithValue(context.TODO(), ctxkey.UsernameKey, envs.AnonymousUsername)

	manifest, _ := example.LoadDemoManifestString(ctx, "workload/simple_deployment", "", "", resCsts.Deploy)
	req, resp := clusterRes.YAMLToFormReq{Yaml: manifest}, clusterRes.CommonResp{}
	err := hdlr.YAMLToForm(ctx, &req, &resp)
	assert.Nil(t, err)
}
