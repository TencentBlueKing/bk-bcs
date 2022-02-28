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

package project

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project/proto/bcsproject"
)

// setResp 设置 response 数据
func setResp(resp *proto.ProjectResponse, code uint32, prefixMsg string, msg string, data interface{}) {
	resp.Code = code
	// 处理message
	if prefixMsg != "" {
		msg = util.JoinString(prefixMsg, msg)
	}
	resp.Message = msg
	// 处理数据
	if val, ok := data.(*proto.Project); ok {
		resp.Data = val
	} else {
		resp.Data = nil
	}
}

// set response for list action
func setListResp(resp *proto.ListProjectsResponse, code uint32, prefixMsg string, msg string, data *proto.ListProjectData) {
	resp.Code = code
	// 处理message
	if prefixMsg != "" {
		msg = util.JoinString(prefixMsg, msg)
	}
	resp.Message = msg
	resp.Data = data
}
