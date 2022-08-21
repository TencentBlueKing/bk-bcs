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
 *
 */

package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/perm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// QueryPermByActionID implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) QueryPermByActionID(ctx context.Context,
	req *cmproto.QueryPermByActionIDRequest, resp *cmproto.QueryPermByActionIDResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := perm.NewCheckUserPermAction(cm.model, cm.iam)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("QueryPermByActionID", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action QueryPermByActionID, req %v, resp %v", reqID, req, resp)
	return nil
}
