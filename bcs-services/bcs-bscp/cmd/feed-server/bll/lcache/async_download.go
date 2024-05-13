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

package lcache

import (
	"fmt"
	"reflect"
	"time"

	"github.com/bluele/gcache"
	prm "github.com/prometheus/client_golang/prometheus"

	clientset "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/client-set"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/cache-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// newAsyncDownload create async download cache instance.
func newAsyncDownload(mc *metric, cs *clientset.ClientSet) *AsyncDownload {
	ad := new(AsyncDownload)
	ad.mc = mc
	opt := cc.FeedServer().FSLocalCache
	client := gcache.New(int(opt.AsyncDownloadCacheSize)).
		LRU().
		EvictedFunc(ad.evictRecorder).
		Expiration(time.Duration(opt.AsyncDownloadCacheTTLSec) * time.Second).
		Build()

	ad.client = client
	ad.cs = cs
	ad.collectHitRate()

	return ad
}

// AsyncDownload is the instance of the async download cache.
type AsyncDownload struct {
	mc     *metric
	client gcache.Cache
	cs     *clientset.ClientSet
}

// GetAsyncDownloadTask Get the async download task cache.
func (ad *AsyncDownload) GetAsyncDownloadTask(kt *kit.Kit, bizID uint32, taskID string) (*types.AsyncDownloadTaskCache,
	error) {
	cacheKey := fmt.Sprintf("%d-%s", bizID, taskID)
	val, err := ad.client.GetIFPresent(cacheKey)
	if err == nil {
		ad.mc.hitCounter.With(prm.Labels{"resource": "async_download_task", "biz": tools.Itoa(bizID)}).Inc()

		// hit from cache.
		meta, yes := val.(*types.AsyncDownloadTaskCache)
		if !yes {
			return nil, fmt.Errorf("unsupported async download task cache value type: %v", reflect.TypeOf(val).String())
		}
		return meta, nil
	}

	if err != gcache.KeyNotFoundError {
		// this is not a not found error, log it.
		logs.Errorf("get biz: %d, task: %s cache from local cache failed, err: %v, rid: %s", bizID,
			taskID, err, kt.Rid)
		// do not return here, try to refresh cache for now.
	}

	start := time.Now()

	// get the cache from cache service directly.
	opt := &pbcs.GetAsyncDownloadTaskReq{
		BizId:  bizID,
		TaskId: taskID,
	}

	resp, err := ad.cs.CS().GetAsyncDownloadTask(kt.RpcCtx(), opt)
	if err != nil {
		ad.mc.errCounter.With(prm.Labels{"resource": "async_download_task", "biz": tools.Itoa(bizID)}).Inc()
		return nil, err
	}

	task := new(types.AsyncDownloadTaskCache)
	err = jsoni.UnmarshalFromString(resp.JsonRaw, &task)
	if err != nil {
		return nil, err
	}

	if err = ad.client.Set(cacheKey, task); err != nil {
		logs.Errorf("refresh biz: %d, task: %s cache failed, err: %v, rid: %s", bizID, taskID, err, kt.Rid)
		// do not return, ignore the error directly.
	}

	ad.mc.refreshLagMS.With(prm.Labels{"resource": "async_download_task", "biz": tools.Itoa(bizID)}).Observe(
		tools.SinceMS(start))

	return task, nil

}

// SetAsyncDownloadTask Set the async download task cache.
func (ad *AsyncDownload) SetAsyncDownloadTask(kt *kit.Kit, task *types.AsyncDownloadTaskCache) error {

	// get the cache from cache service directly.
	opt := &pbcs.SetAsyncDownloadTaskReq{
		BizId:    task.BizID,
		TaskId:   task.TaskID,
		AppId:    task.AppID,
		FilePath: task.FilePath,
		FileName: task.FileName,
	}

	_, err := ad.cs.CS().SetAsyncDownloadTask(kt.RpcCtx(), opt)
	if err != nil {
		ad.mc.errCounter.With(prm.Labels{"resource": "async_download_task", "biz": tools.Itoa(task.BizID)}).Inc()
		return err
	}

	cacheKey := fmt.Sprintf("%d-%s", task.BizID, task.TaskID)

	if err = ad.client.Set(cacheKey, task); err != nil {
		logs.Errorf("refresh biz: %d, task: %s cache failed, err: %v, rid: %s", task.BizID, task.TaskID, err, kt.Rid)
		// do not return, ignore the error directly.
	}

	return nil

}

func (ad *AsyncDownload) evictRecorder(key interface{}, _ interface{}) {
	taskID, yes := key.(string)
	if !yes {
		return
	}

	ad.mc.evictCounter.With(prm.Labels{"resource": "async_download_task"}).Inc()

	if logs.V(2) {
		logs.Infof("evict async download task cache, task: %d", taskID)
	}
}

func (ad *AsyncDownload) collectHitRate() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			ad.mc.hitRate.With(prm.Labels{"resource": "async_download_task"}).Set(ad.client.HitRate())
		}
	}()
}
