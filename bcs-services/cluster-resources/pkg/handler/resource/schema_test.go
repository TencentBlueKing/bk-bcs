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

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/validator"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestGetResFormSchema(t *testing.T) {
	hdlr := New()
	ctx := context.TODO()

	assert.Nil(t, i18n.InitMsgMap())

	for kind := range validator.FormSupportedResAPIVersion {
		req, resp := clusterRes.GetResFormSchemaReq{Kind: kind}, clusterRes.CommonResp{}
		err := hdlr.GetResFormSchema(ctx, &req, &resp)
		assert.Nil(t, err)
	}
}

func TestGetFormSupportedApiVersions(t *testing.T) {
	hdlr := New()
	ctx := context.TODO()

	for kind := range validator.FormSupportedResAPIVersion {
		req, resp := clusterRes.GetFormSupportedApiVersionsReq{Kind: kind}, clusterRes.CommonResp{}
		err := hdlr.GetFormSupportedAPIVersions(ctx, &req, &resp)
		assert.Nil(t, err)
	}
}
