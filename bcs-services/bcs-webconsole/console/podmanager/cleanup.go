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

// Package podmanager xxx
package podmanager

import (
	"context"
	"fmt"
	"strconv"
	"time"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/go-redis/redis/v8"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

// CleanUpManager Pod 清理
// 1. 定时统计活跃集群和命名空间信息
// 2. 定时扫描configmap / pod, 对比活跃pod数据, 如果不存在或者已经不活跃, 执行删除操作
// 3. 定期清理redis缓存数据
type CleanUpManager struct {
	ctx         context.Context
	redisClient *redis.Client
	podKey      string
	clusterKey  string
}

// NewCleanUpManager 定期清理控制
func NewCleanUpManager(ctx context.Context) *CleanUpManager {
	redisClient := storage.GetDefaultRedisSession().Client

	return &CleanUpManager{
		ctx:         ctx,
		redisClient: redisClient,
		podKey:      fmt.Sprintf(webConsolePodHeartbeatKey, config.G.Base.RunEnv),
		clusterKey:  fmt.Sprintf(webConsoleClusterHeartbeatKey, config.G.Base.RunEnv),
	}
}

// Heartbeat : 记录pod心跳, 定时上报存活, 清理时需要使用
func (p *CleanUpManager) Heartbeat(podCtx *types.PodContext) error {
	now := time.Now().Unix()

	// 同步Pod信息
	uid := getUid(podCtx.ProjectCode, podCtx.AdminClusterId, podCtx.Username)
	if err := p.redisClient.ZAdd(p.ctx, p.podKey, &redis.Z{Score: float64(now), Member: uid}).Err(); err != nil {
		return err
	}

	// 同步集群信息
	err := p.redisClient.ZAdd(p.ctx, p.clusterKey, &redis.Z{Score: float64(now), Member: podCtx.AdminClusterId}).Err()
	if err != nil {
		return err
	}

	return nil
}

// fetchAliveRes 查询活跃资源, 可为集群和Pod uid
func (p *CleanUpManager) fetchAliveRes(key string, expireSeconds int64) ([]string, error) {
	now := time.Now().Unix()
	min := strconv.FormatInt(now-expireSeconds, 10)
	max := strconv.FormatInt(now+expireSeconds, 10)

	vals, err := p.redisClient.ZRangeByScore(p.ctx, key, &redis.ZRangeBy{Min: min, Max: max}).Result()
	if err != nil {
		return nil, err
	}

	// 清理一年前的数据
	delTime := strconv.FormatInt(now-3600*24*365, 10)
	p.redisClient.ZRem(p.ctx, key, &redis.ZRangeBy{Min: "0", Max: delTime})

	return vals, nil
}

// CleanupRes 单个集群清理
func (p *CleanUpManager) CleanupRes() error {
	aliveClusters, err := p.fetchAliveRes(p.clusterKey, clusterExpireSeconds)
	if err != nil {
		return err
	}

	if config.G.WebConsole.AdminClusterId != "" {
		aliveClusters = append(aliveClusters, config.G.WebConsole.AdminClusterId)
	}

	alivePodUid, err := p.fetchAliveRes(p.podKey, UserCtxExpireTime)
	if err != nil {
		return err
	}

	alivePodMap := map[string]struct{}{}
	for _, v := range alivePodUid {
		alivePodMap[v] = struct{}{}
	}

	// 只获取固定命名空间
	namespace := GetNamespace()

	for _, clusterId := range aliveClusters {
		if err := p.cleanUserPodByCluster(clusterId, namespace, alivePodMap); err != nil {
			return err
		}
		if err := p.cleanConfigMap(clusterId, namespace, alivePodMap); err != nil {
			return err
		}
	}

	return nil
}

// cleanUserPodByCluster 清理用户下的相关集群pod
func (p *CleanUpManager) cleanUserPodByCluster(clusterId string, namespace string,
	alivePodMap map[string]struct{}) error {
	k8sClient, err := k8sclient.GetK8SClientByClusterId(clusterId)
	if err != nil {
		return err
	}

	podList, err := k8sClient.CoreV1().Pods(namespace).List(p.ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	metrics.CollectPodCount(clusterId, namespace, float64(len(podList.Items)))

	// 过期时间
	now := time.Now()
	expireTime := now.Add(-UserPodExpireTime)

	for _, pod := range podList.Items {
		if pod.Status.Phase == v1.PodPending {
			continue
		}

		startTime := pod.Status.StartTime.Time
		// 小于一个周期的pod不清理
		if expireTime.Before(startTime) {
			logger.Infof("pod %s exist time %s < %s, just ignore", pod.Name, now.Sub(startTime), UserPodExpireTime)
			continue
		}

		// 有心跳上报的不清理
		if _, ok := alivePodMap[pod.Name]; ok {
			continue
		}

		// 命令行持久化保存
		if err := historyMgr.persistenceBashHistory(
			k8sClient, pod.Name, pod.Namespace, pod.Spec.Containers[0].Name, clusterId); err != nil {
			logger.Errorf("persistence history failed, err: %s", err)
		}

		// 删除pod
		if err := k8sClient.CoreV1().Pods(namespace).Delete(p.ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
			logger.Errorf("delete pod(%s) failed, err: %s", pod.Name, err)
			continue
		}

		logger.Infof("delete pod %s done", pod.Name)
	}

	return nil
}

// cleanConfigMap 清理configmap
func (p *CleanUpManager) cleanConfigMap(clusterId string, namespace string, alivePodMap map[string]struct{}) error {
	k8sClient, err := k8sclient.GetK8SClientByClusterId(clusterId)
	if err != nil {
		return err
	}

	configMapList, err := k8sClient.CoreV1().ConfigMaps(namespace).List(p.ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// 过期时间
	now := time.Now()
	expireTime := now.Add(-UserPodExpireTime)

	for _, configMap := range configMapList.Items {
		// 小于一个周期的pod不清理
		createTime := configMap.GetCreationTimestamp().Time
		if expireTime.Before(createTime) {
			logger.Infof("configmap %s exist time %s < %s, just ignore", configMap.Name, now.Sub(createTime), UserPodExpireTime)
			continue
		}

		shouldDelete := false

		uid, ok := configMap.Labels[uidKey]
		if !ok {
			shouldDelete = true
		}

		if _, ok := alivePodMap[uid]; !ok {
			shouldDelete = true
		}

		if !shouldDelete {
			continue
		}

		if err := k8sClient.CoreV1().ConfigMaps(namespace).Delete(p.ctx, configMap.Name, metav1.DeleteOptions{}); err != nil {
			logger.Errorf("delete configmap(%s) failed, err: %s", configMap.Name, err)
			continue
		}

		logger.Infof("configmap %s deleted", configMap.Name)
	}

	return nil
}

// Run xxx
func (p *CleanUpManager) Run() error {
	interval := time.NewTicker(CleanUserPodInterval)
	defer interval.Stop()

	sessionCleanupMgr := sessions.NewStore()

	for {
		select {
		case <-p.ctx.Done():
			logger.Info("close CleanUpManager done")
			return nil
		case <-interval.C:
			// 清理 pods 数据
			now := time.Now()
			if err := p.CleanupRes(); err != nil {
				logger.Errorf("clean webconsole pod failed, duration=%s, err=%s", err, time.Since(now))
			} else {
				logger.Infof("clean webconsole pod done, duration=%s", time.Since(now))
			}

			// 清理 sessions 数据
			now = time.Now()
			if err := sessionCleanupMgr.Cleanup(p.ctx); err != nil {
				logger.Errorf("clean sessions failed, duration=%s, err=%s", err, time.Since(now))
			} else {
				logger.Infof("clean sessions done, duration=%s", time.Since(now))
			}
		}
	}
}
