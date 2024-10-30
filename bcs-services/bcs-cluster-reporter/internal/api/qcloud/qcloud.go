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
 */

// Package qcloud xxx
package qcloud

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/rest"
)

// NodeMetaData node info struct
type NodeMetaData struct {
	InstanceID      string
	Region          string
	Zone            string
	InstanceType    string
	InstanceImageID string
}

// GetQcloudNodeMetadata get cvm info
func GetQcloudNodeMetadata() (*NodeMetaData, error) {
	var err error
	var result = &NodeMetaData{}

	result.InstanceID, err = GetMetadata("instance-id")
	if err != nil {
		return nil, err
	}
	result.Region, err = GetMetadata("placement/region")
	if err != nil {
		return nil, err
	}
	result.Zone, err = GetMetadata("placement/zone")
	if err != nil {
		return nil, err
	}
	result.InstanceImageID, err = GetMetadata("instance/image-id")
	if err != nil {
		return nil, err
	}
	result.InstanceType, err = GetMetadata("instance/instance-type")
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetMetadata get cvm info by qcloud api
func GetMetadata(item string) (string, error) {
	httpClient := &http.Client{}
	svcUrl, _ := url.Parse(fmt.Sprintf("http://metadata.tencentyun.com/latest/meta-data/%s", item))
	req := rest.NewRequest(httpClient, "GET", svcUrl, nil)
	data, err := req.Do()
	if err != nil {
		return "", err
	}
	return string(data), nil
}
