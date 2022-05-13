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

package namespace

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/resp"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// Handler ...
type Handler struct{}

// New ...
func New() *Handler {
	return &Handler{}
}

// ListNS ...
func (h *Handler) ListNS(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) error {
	ret, err := cli.NewNSCliByClusterID(ctx, req.ClusterID).List(
		ctx, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	if err != nil {
		return err
	}

	respDataBuilder, err := respUtil.NewRespDataBuilder(ret, res.NS, req.Format)
	if err != nil {
		return err
	}
	respData, err := respDataBuilder.BuildList()
	if err != nil {
		return err
	}

	resp.Data, err = pbstruct.Map2pbStruct(respData)
	return err
}
