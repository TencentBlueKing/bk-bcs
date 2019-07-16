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

package resthdrs

import (
	"fmt"

	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
	"github.com/emicklei/go-restful"
)

type UpdateTkeLbForm struct {
	ClusterRegion string `json:"cluster_region" validate:"required"`
	SubnetId      string `json:"subnet_id" validate:"required"`
}

func UpdateTkeLbSubnet(request *restful.Request, response *restful.Response) {
	blog.Debug(fmt.Sprintf("Create or Update tke lb subnet"))
	form := UpdateTkeLbForm{}
	request.ReadEntity(&form)

	err := validate.Struct(&form)
	if err != nil {
		response.WriteEntity(FormatValidationError(err))
		return
	}

	err = sqlstore.SaveTkeLbSubnet(form.ClusterRegion, form.SubnetId)
	if err != nil {
		message := fmt.Sprintf("errcode: %d, can not update tke lb subnet, error: %s", common.BcsErrApiInternalDbError, err.Error())
		WriteClientError(response, "CANNOT_UPDATE_TKE_LB_SUBNET", message)
		return
	}
}
