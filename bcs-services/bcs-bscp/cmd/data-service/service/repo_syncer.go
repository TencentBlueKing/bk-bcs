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

package service

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/dao"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/repository"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/space"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// NewRepoSyncer new repo syncer
func NewRepoSyncer(set dao.Set, repo repository.Provider, spaceMgr *space.Manager, sd serviced.Service) *RepoSyncer {
	return &RepoSyncer{
		set:      set,
		repo:     repo,
		spaceMgr: spaceMgr,
		state:    sd,
		metric:   initMetric(),
	}
}

// RepoSyncer is repo syncer which sync file from master to slave repo
type RepoSyncer struct {
	set      dao.Set
	repo     repository.Provider
	spaceMgr *space.Manager
	state    serviced.Service
	metric   *metric
}

// Run runs the repo syncer
func (s *RepoSyncer) Run() {
	logs.Infof("begin run repo syncer, sync period is %d seconds", cc.DataService().Repo.SyncPeriodSeconds)
	kt := kit.New()
	ctx, cancel := context.WithCancel(kt.Ctx)
	kt.Ctx = ctx

	go func() {
		notifier := shutdown.AddNotifier()
		<-notifier.Signal
		cancel()
		notifier.Done()
	}()

	go s.collectMetrics()

	// sync incremental files
	go s.syncIncremental(kt)

	go func() {
		// sync all files at once after service starts a while
		time.Sleep(time.Minute)
		if s.state.IsMaster() {
			s.syncAll(kt)
		} else {
			logs.Infof("current service instance is slave, skip the task of syncing all repo files")
		}

		// sync all files periodically
		ticker := time.NewTicker(time.Duration(cc.DataService().Repo.SyncPeriodSeconds) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-kt.Ctx.Done():
				logs.Infof("stop repo syncer success")
				return
			case <-ticker.C:
				if !s.state.IsMaster() {
					logs.Infof("current service instance is slave, skip the task of syncing all repo files")
					continue
				}
				s.syncAll(kt)
			}
		}
	}()
}

// collectMetrics collects metrics for repo syncer periodically
func (s *RepoSyncer) collectMetrics() {
	syncMgr := s.repo.SyncManager()
	client := syncMgr.QueueClient()
	syncQueue := syncMgr.QueueName()
	ackQueue := syncMgr.AckQueueName()
	ctx := context.Background()
	for {
		time.Sleep(time.Second * 5)
		if syncQueueLen, err := client.LLen(ctx, syncQueue); err != nil {
			logs.Errorf("get sync queue length failed, err: %v", err)
		} else {
			s.metric.syncQueueLen.Set(float64(syncQueueLen))
		}

		if syncQueueLen, err := client.LLen(ctx, ackQueue); err != nil {
			logs.Errorf("get sync queue length failed, err: %v", err)
		} else {
			s.metric.syncQueueLen.Set(float64(syncQueueLen))
		}
	}
}

type syncStat struct {
	bizID       int32
	total       int32
	success     int32
	failed      int32
	skip        int32
	costSeconds float64
}

type noFiles struct {
	bizID     int32
	fileSigns []string
}

var (
	stats          []syncStat
	noFileInMaster []noFiles
	syncFailedCnt  int32
)

const failedLimit = 100

// syncAll syncs all files from master to slave repo
func (s *RepoSyncer) syncAll(kt *kit.Kit) {
	logs.Infof("start to sync all repo files")
	start := time.Now()
	// clear stats data
	stats = make([]syncStat, 0)
	// get all biz
	bizs := s.spaceMgr.AllCMDBSpaces()
	// sync files for all bizs
	// we think the file count would not be too large for every biz, eg:<100000
	// so, we directly retrieve all file signatures under one biz from the db
	// this syncs biz serially (one by one) , and sync files under every biz concurrently
	for biz := range bizs {
		b, _ := strconv.Atoi(biz)
		bizID := uint32(b)
		var allSigns []string
		var normalSigns, releasedNormalSigns, tmplSigns, releasedTmplSigns []string
		var err error
		// 未发布的普通配置项
		if normalSigns, err = s.set.Content().ListAllCISigns(kt, bizID); err != nil {
			logs.Errorf("list normal ci signs failed, err: %v, rid: %s", err, kt.Rid)
		} else {
			allSigns = append(allSigns, normalSigns...)
		}

		// 已发布的普通配置项
		if releasedNormalSigns, err = s.set.ReleasedCI().ListAllCISigns(kt, bizID); err != nil {
			logs.Errorf("list released normal ci signs failed, err: %v, rid: %s", err, kt.Rid)
		} else {
			allSigns = append(allSigns, releasedNormalSigns...)
		}

		// 未发布的模版配置项
		if tmplSigns, err = s.set.TemplateRevision().ListAllCISigns(kt, bizID); err != nil {
			logs.Errorf("list template ci signs failed, err: %v, rid: %s", err, kt.Rid)
		} else {
			allSigns = append(allSigns, tmplSigns...)
		}

		// 已发布的模版配置项
		if releasedTmplSigns, err = s.set.ReleasedAppTemplate().ListAllCISigns(kt, bizID); err != nil {
			logs.Errorf("list released template ci signs failed, err: %v, rid: %s", err, kt.Rid)
		} else {
			allSigns = append(allSigns, releasedTmplSigns...)
		}

		allSigns = tools.RemoveDuplicateStrings(allSigns)
		s.syncOneBiz(kt, bizID, allSigns)
		if atomic.LoadInt32(&syncFailedCnt) > failedLimit {
			logs.Infof("sync all repo files failed too many times(> %d), stop the sync task, "+
				"you should check the health of repo service, cost time: %s, rid: %s", failedLimit,
				time.Since(start), kt.Rid)
			return
		}
	}

	logs.Infof("sync all repo files finished, cost time: %s, rid: %s, stats: %#v", time.Since(start), kt.Rid, stats)
	if len(noFileInMaster) > 0 {
		logs.Infof("sync all repo files found some files not in master, please check the master repo, rid: %s, "+
			"info: %#v", kt.Rid, noFileInMaster)
	}
}

// syncOneBiz syncs all files under one biz concurrently
func (s *RepoSyncer) syncOneBiz(kt *kit.Kit, bizID uint32, signs []string) {
	start := time.Now()
	syncMgr := s.repo.SyncManager()
	var success, failed, skip int32
	noFilesCh := make(chan string, 1)
	var nofiles []string

	// save info for no file in master
	go func() {
		for file := range noFilesCh {
			nofiles = append(nofiles, file)
		}
		if len(nofiles) > 0 {
			noFileInMaster = append(noFileInMaster, noFiles{
				bizID:     int32(bizID),
				fileSigns: nofiles,
			})
		}
	}()

	// sync files concurrently
	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(10)
	for _, si := range signs {
		sign := si
		g.Go(func() error {
			if atomic.LoadInt32(&syncFailedCnt) > failedLimit {
				return nil
			}
			kt2 := kt.Clone()
			kt2.BizID = bizID
			isSkip, err := syncMgr.Sync(kt2, sign)
			if err != nil {
				logs.Errorf("sync file sign %s for biz %d failed, err: %v, rid: %s", sign, bizID, err, kt2.Rid)
				atomic.AddInt32(&failed, 1)
				atomic.AddInt32(&syncFailedCnt, 1)
				if err == repository.ErrNoFileInMaster {
					noFilesCh <- sign
				}
				return err
			}
			if isSkip {
				atomic.AddInt32(&skip, 1)
			} else {
				atomic.AddInt32(&success, 1)
			}
			return nil
		})
	}
	_ = g.Wait()

	close(noFilesCh)
	cost := time.Since(start)
	stat := syncStat{
		bizID:       int32(bizID),
		total:       int32(len(signs)),
		success:     success,
		failed:      failed,
		skip:        skip,
		costSeconds: cost.Seconds(),
	}
	stats = append(stats, stat)
	logs.Infof("sync biz [%d] repo files finished, cost time: %s, rid: %s, stat: %#v", bizID, cost, kt.Rid, stat)
}

// syncIncremental syncs incremental files
func (s *RepoSyncer) syncIncremental(kt *kit.Kit) {
	syncMgr := s.repo.SyncManager()
	client := syncMgr.QueueClient()
	syncQueue := syncMgr.QueueName()
	ackQueue := syncMgr.AckQueueName()

	// consider the service crash, the ackQueue may have some messages, we move them to syncQueue and handle them again
	mvCnt := 0
	for {
		kt2 := kt.Clone()
		msg, err := client.RPopLPush(kt2.Ctx, ackQueue, syncQueue)
		if err != nil {
			logs.Errorf("move msg from ackQueue to syncQueue failed, err: %v, rid: %s", err, kt2.Rid)
		}
		// msg is empty which means occurring redis.Nil and ackQueue is empty
		if msg == "" {
			break
		}
		mvCnt++
	}
	if mvCnt > 0 {
		logs.Errorf("have moved %d msg from ackQueue to syncQueue", mvCnt)
	}

	// sync files concurrently
	workerCount := 10
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				kt2 := kt.Clone()
				// block pop msg from syncQueue and push one to ackQueue
				msg, err := client.BRPopLPush(kt2.Ctx, syncQueue, ackQueue, 0)
				if err != nil {
					logs.Errorf("pop msg from syncQueue failed, err: %v, rid: %s", err, kt2.Rid)
					time.Sleep(time.Second)
					continue
				}

				var bizID uint32
				var sign string
				bizID, sign, err = syncMgr.ParseQueueMsg(msg)
				if err != nil {
					logs.Errorf("parse queue msg failed, err: %v, rid: %s", err, kt2.Rid)
					// remove the invalid msg from ackQueue
					if err = client.LRem(kt2.Ctx, ackQueue, 0, msg); err != nil {
						logs.Errorf("remove msg from ackQueue failed, err: %v, rid: %s", err, kt2.Rid)
					}
					continue
				}

				// sync the file with the specific bizID and signature
				kt2.BizID = bizID
				if _, err = syncMgr.Sync(kt2, sign); err != nil {
					logs.Errorf("sync file failed, err: %v, rid: %s", err, kt2.Rid)
					// if failed for no file in master, continue and not to retry it
					if err == repository.ErrNoFileInMaster {
						continue
					}
					// if failed for other reasons, retry after a delay
					go retryMessage(kt2, syncMgr, ackQueue, msg, sign)
					continue
				} else {
					// if success, remove the msg from ackQueue
					if err = client.LRem(kt2.Ctx, ackQueue, 0, msg); err != nil {
						logs.Errorf("remove msg from ackQueue failed, err: %v, rid: %s", err, kt2.Rid)
					}
				}
			}
		}()
	}

	// wait for all workers to finish (they won't, as this is an infinite loop)
	wg.Wait()
}

// retryMessage retry handle the failed msg after a delay
// if retry failed beyond the limit, finish the retry and all-sync mechanism will handle it
func retryMessage(kt *kit.Kit, syncMgr *repository.SyncManager, ackQueue, msg, sign string) {
	limit := 3
	count := 0
	for {
		// wait for 1 minute before retrying
		time.Sleep(1 * time.Minute)

		if count == 0 {
			// remove the msg from ackQueue,
			if err := syncMgr.QueueClient().LRem(kt.Ctx, ackQueue, 0, msg); err != nil {
				logs.Errorf("remove msg from ackQueue during retry failed, err: %v, rid: %s", err, kt.Rid)
				continue
			}
		}
		count++

		if _, err := syncMgr.Sync(kt, sign); err != nil {
			logs.Errorf("sync retry count %d failed, err: %v, rid: %s", count, err, kt.Rid)
			// retry failed beyond the limit, return and all-sync mechanism will handle it
			if count >= limit {
				return
			}
		}
		return
	}
}
