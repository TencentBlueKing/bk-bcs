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

package v1

import (
	"fmt"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/api/proto/logmanager"
)

// ListLogCollectionTask call ListLogCollectionTask of LogManager grpc server
func (m *LogManager) ListLogCollectionTask(req *proto.ListLogCollectionTaskReq) ([]*proto.ListLogCollectionTaskRespItem, error) {
	resp, err := m.client.ListLogCollectionTask(m.ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.ErrName != proto.ErrCode_ERROR_OK {
		return nil, fmt.Errorf(resp.Message)
	}
	return resp.Data, nil
}
