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
	"fmt"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// GetK8SResTemplate xxx
func (h *Handler) GetK8SResTemplate(
	ctx context.Context, req *clusterRes.GetK8SResTemplateReq, resp *clusterRes.CommonResp,
) (err error) {
	// 不在列表里面的，则认为是自定义资源
	if !slice.StringInSlice(req.Kind, example.HasDemoManifestResKinds) {
		req.Kind = resCsts.CObj
	}
	conf, err := example.LoadResConf(ctx, req.Kind)
	if err != nil {
		return err
	}
	conf["references"], err = example.LoadResRefs(ctx, req.Kind)
	if err != nil {
		return err
	}
	for _, tmpl := range mapx.GetList(conf, "items") {
		t, _ := tmpl.(map[string]interface{})
		path := fmt.Sprintf("%s/%s", conf["class"], t["name"])
		if t["manifest"], err = example.LoadDemoManifest(
			ctx, path, req.ClusterID, req.Namespace, req.Kind,
		); err != nil {
			return err
		}
	}
	resp.Data, err = pbstruct.Map2pbStruct(conf)
	return err
}
