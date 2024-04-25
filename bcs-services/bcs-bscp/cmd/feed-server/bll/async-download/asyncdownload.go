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

// Package asyncdownload NOTES
package asyncdownload

import (
	clientset "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/client-set"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/lcache"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pkgtypes "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// New initialize the release service instance.
func New(cs *clientset.ClientSet, cache *lcache.Cache) (*AsyncDownload, error) {
	return &AsyncDownload{
		cs:    cs,
		cache: cache,
	}, nil
}

// AsyncDownload defines async download related operations.
type AsyncDownload struct {
	cs    *clientset.ClientSet
	cache *lcache.Cache
}

// CreateAsyncDownloadTask create sync download task record.
func (ad *AsyncDownload) CreateAsyncDownloadTask(kt *kit.Kit, opts *types.AsyncDownloadTask) error {
	return ad.cache.AsyncDownload.SetAsyncDownloadTask(kt, &pkgtypes.AsyncDownloadTaskCache{
		BizID:    opts.BizID,
		AppID:    opts.AppID,
		TaskID:   opts.TaskID,
		FilePath: opts.FilePath,
		FileName: opts.FileName,
	})
}

// GetAsyncDownloadTask get sync download task record.
func (ad *AsyncDownload) GetAsyncDownloadTask(kt *kit.Kit, bizID uint32, taskID string) (
	*types.AsyncDownloadTask, error) {

	data, err := ad.cache.AsyncDownload.GetAsyncDownloadTask(kt, bizID, taskID)
	if err != nil {
		return nil, err
	}

	task := &types.AsyncDownloadTask{
		BizID:    data.BizID,
		AppID:    data.AppID,
		TaskID:   data.TaskID,
		FilePath: data.FilePath,
		FileName: data.FileName,
	}

	return task, nil
}
