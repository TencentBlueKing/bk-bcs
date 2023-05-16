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

package api

import (
	"encoding/json"

	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
)

// NewDescribeOSImagesRequest request
func NewDescribeOSImagesRequest() (request *DescribeOSImagesRequest) {
	request = &DescribeOSImagesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeOSImages")

	return
}

// NewDescribeOSImagesResponse response
func NewDescribeOSImagesResponse() (response *DescribeOSImagesResponse) {
	response = &DescribeOSImagesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeOSImagesRequestParams Predefined struct for user
type DescribeOSImagesRequestParams struct {
}

// DescribeOSImagesRequest xxx
type DescribeOSImagesRequest struct {
	*tchttp.BaseRequest
}

// ToJsonString toString
func (r *DescribeOSImagesRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString xxx
func (r *DescribeOSImagesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}

	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeOSImagesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

// DescribeOSImagesResponseParams xxx
type DescribeOSImagesResponseParams struct {
	// 镜像信息列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	OSImageSeriesSet []*OSImage `json:"OSImageSeriesSet,omitempty" name:"OSImageSeriesSet"`

	// 镜像数量
	// 注意：此字段可能返回 null，表示取不到有效值。
	TotalCount *uint64 `json:"TotalCount,omitempty" name:"TotalCount"`

	// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
}

// DescribeOSImagesResponse xxx
type DescribeOSImagesResponse struct {
	*tchttp.BaseResponse
	Response *DescribeOSImagesResponseParams `json:"Response"`
}

// ToJsonString xxx
func (r *DescribeOSImagesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString xxx
func (r *DescribeOSImagesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// OSImage os info
type OSImage struct {
	// os聚合名称
	SeriesName *string `json:"SeriesName,omitempty" name:"SeriesName"`
	// os别名
	Alias *string `json:"Alias,omitempty" name:"Alias"`
	// os架构
	Arch *string `json:"Arch,omitempty" name:"Arch"`
	// os名称
	OsName *string `json:"OsName,omitempty" name:"OsName"`
	// 操作系统类型(分为定制和非定制，取值分别为:DOCKER_CUSTOMIZE、GENERAL)
	OsCustomizeType *string `json:"OsCustomizeType,omitempty" name:"OsCustomizeType"`
	// os是否下线(online表示在线,offline表示下线)
	Status *string `json:"Status,omitempty" name:"Status"`
	// 镜像id
	ImageId *string `json:"ImageId,omitempty" name:"ImageId"`
}
