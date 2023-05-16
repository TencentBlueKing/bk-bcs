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

package auth

import (
	"fmt"

	bcsIAM "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	iam "github.com/TencentBlueKing/iam-go-sdk"
	iamBackendClient "github.com/TencentBlueKing/iam-go-sdk/client"
	iamExpr "github.com/TencentBlueKing/iam-go-sdk/expression"
	iamOP "github.com/TencentBlueKing/iam-go-sdk/expression/operator"
	"github.com/mitchellh/mapstructure"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
)

// ListAuthorizedProjectIDs 过滤有权限的项目 ID，如果为 any，则返回 true
func ListAuthorizedProjectIDs(username string) ([]string, bool, error) {
	// 组装 iam request
	iamReq := makeIAMRequest(username, ProjectView)
	if err := iamReq.Validate(); err != nil {
		return []string{}, false, err
	}
	// 获取 policy
	policy, err := makeIAMPolicy(iamReq)
	if err != nil || len(policy) == 0 {
		return []string{}, false, err
	}
	f, err := makeFilter(policy)
	if err != nil || len(f) == 0 {
		return []string{}, false, err
	}
	// 解析policy values，获取project id
	if f["op"] == iamOP.Any {
		return []string{}, true, nil
	}
	// value 为 []interface{}
	val, _ := f["value"].([]interface{})
	var ids []string
	for _, v := range val {
		vStr, _ := v.(string)
		ids = append(ids, vStr)
	}
	return ids, false, nil
}

// makeIAMRequest 生成 iam request
func makeIAMRequest(username, actionID string) iam.Request {
	subject := iam.Subject{Type: "user", ID: username}
	action := iam.Action{ID: actionID}
	return iam.NewRequest(bcsIAM.SystemIDBKBCS, subject, action, nil)
}

// makeIAMPolicy 生成查询策略
func makeIAMPolicy(iamReq iam.Request) (map[string]interface{}, error) {
	backendClient := iamBackendClient.NewIAMBackendClient(
		config.GlobalConf.IAM.GatewayHost,
		true,
		bcsIAM.SystemIDBKBCS,
		config.GlobalConf.App.Code, config.GlobalConf.App.Secret,
	)
	policy, err := backendClient.PolicyQuery(iamReq)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

func makeFilter(policy map[string]interface{}) (map[string]interface{}, error) {
	expr := iamExpr.ExprCell{}
	if err := mapstructure.Decode(policy, &expr); err != nil {
		return nil, err
	}
	// 处理 OP
	switch expr.OP {
	case iamOP.Eq:
		return map[string]interface{}{
			"value": []interface{}{expr.Value}, "op": iamOP.In,
		}, nil
	case iamOP.In, iamOP.Any:
		return map[string]interface{}{
			"value": expr.Value, "op": expr.OP,
		}, nil
	case iamOP.OR, iamOP.AND:
		return parseContent(expr.Content), nil
	default:
		return nil, errorx.NewIAMOPErr(fmt.Sprintf("not support op: %v", expr.OP))
	}
}

// parseContent 解析 content, 仅处理嵌套的第一级
func parseContent(c []iamExpr.ExprCell) map[string]interface{} {
	var ids []interface{}
	for _, expr := range c {
		if expr.Field != "project.id" {
			continue
		}
		switch expr.OP {
		case iamOP.Any:
			return map[string]interface{}{
				"value": expr.Value, "op": expr.OP,
			}
		case iamOP.In, iamOP.Eq:
			v, ok := expr.Value.([]interface{})
			if !ok {
				continue
			}
			ids = append(ids, v...)
		}
	}
	// 返回解析后的数据
	return map[string]interface{}{
		"value": ids, "op": iamOP.In,
	}
}
