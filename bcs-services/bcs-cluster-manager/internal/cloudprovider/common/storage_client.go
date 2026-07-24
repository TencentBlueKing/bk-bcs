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

package common

import (
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"
)

// StorageOptions storage options
type StorageOptions struct {
	Host  string
	Token string
	Debug bool
}

// StorageClient storage client
type StorageClient struct {
	host        string
	token       string
	serverDebug bool
}

// StorageCli global storage client
var StorageCli *StorageClient

// SetStorageClient set storage client
func SetStorageClient(options StorageOptions) error {
	cli, err := NewStorageClient(options)
	if err != nil {
		return err
	}

	StorageCli = cli
	return nil
}

// NewStorageClient create bcs storage client
func NewStorageClient(options StorageOptions) (*StorageClient, error) {
	c := &StorageClient{
		host:        options.Host,
		token:       options.Token,
		serverDebug: options.Debug,
	}

	return c, nil
}

// SyncClusterData sync cluster data to storage
func (s *StorageClient) SyncClusterData(clusterID string, data map[string]interface{}) error {
	if s == nil {
		return ErrServerNotInit
	}

	reqUrl := fmt.Sprintf("%s/bcsapi/v4/storage/clusters/%s", s.host, clusterID)
	respData := &StorageResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Put(reqUrl).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("Authorization", fmt.Sprintf("Bearer %s", s.token)).
		SetDebug(s.serverDebug).
		Send(SyncClusterDataRequest{Data: data}).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api SyncClusterData failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api SyncClusterData failed: %v", respData.Message)
		return errors.New(respData.Message)
	}

	// successfully request
	blog.Infof("call api SyncClusterData with url(%s) successfully", reqUrl)

	return nil
}

// DelClusterData del storage cluster data
func (s *StorageClient) DelClusterData(clusterID string) error {
	if s == nil {
		return ErrServerNotInit
	}

	reqUrl := fmt.Sprintf("%s/bcsapi/v4/storage/clusters/%s", s.host, clusterID)
	respData := &StorageResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqUrl).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("Authorization", fmt.Sprintf("Bearer %s", s.token)).
		SetDebug(s.serverDebug).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api DelClusterData failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api DelClusterData failed: %v", respData.Message)
		return errors.New(respData.Message)
	}

	// successfully request
	blog.Infof("call api DelClusterData with url(%s) successfully", reqUrl)

	return nil
}
