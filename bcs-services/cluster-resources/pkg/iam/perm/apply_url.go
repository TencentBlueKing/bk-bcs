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

package perm

import (
	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
)

// ApplyURLGenerator 权限申请链接生成器
type ApplyURLGenerator struct {
	cli *bkiam.IAM
}

// NewApplyURLGenerator ...
func NewApplyURLGenerator() *ApplyURLGenerator {
	return &ApplyURLGenerator{cli: conf.G.IAM.Cli}
}

// Gen 生成权限申请跳转链接
func (g *ApplyURLGenerator) Gen(username string, actionReqList []ActionResourcesRequest) (string, error) {
	application := g.makeApplication(actionReqList)
	return g.cli.GetApplyURL(application, "", username)
}

func (g *ApplyURLGenerator) makeApplication(actionReqList []ActionResourcesRequest) bkiam.Application {
	actions := []bkiam.ApplicationAction{}
	for _, req := range actionReqList {
		actions = append(actions, req.ToAction())
	}
	return bkiam.Application{
		SystemID: conf.G.IAM.SystemID,
		Actions:  actions,
	}
}
