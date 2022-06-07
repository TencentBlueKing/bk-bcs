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

package podmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CleanUpManager Pod 清理
type CleanUpManager struct {
	ctx         context.Context
	redisClient *redis.Client
}

func NewCleanUpManager(ctx context.Context) *CleanUpManager {
	redisClient := storage.GetDefaultRedisSession().Client

	return &CleanUpManager{
		ctx:         ctx,
		redisClient: redisClient,
	}
}

// Heartbeat : 记录pod心跳, 定时上报存活, 清理时需要使用
func (p *CleanUpManager) Heartbeat(podCtx *types.PodContext) error {
	podCleanUpCtx := types.TimestampPodContext{
		PodContext: *podCtx,
		Timestamp:  time.Now().Unix(),
	}
	payload, err := json.Marshal(podCleanUpCtx)
	if err != nil {
		return err
	}
	key := fmt.Sprintf(webConsoleHeartbeatKey, config.G.Base.RunEnv)
	if _, err := p.redisClient.HSet(p.ctx, key, podCtx.PodName, payload).Result(); err != nil {
		return err
	}

	return nil
}

// getActiveUserPod 获取活跃 kubectld pod
func (p *CleanUpManager) getActiveUserPod() (map[string][]*types.TimestampPodContext, error) {
	podExpireTime := time.Now().Unix() - UserCtxExpireTime
	key := fmt.Sprintf(webConsoleHeartbeatKey, config.G.Base.RunEnv)

	values, err := p.redisClient.HGetAll(p.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	expirePods := []string{}
	results := map[string][]*types.TimestampPodContext{}
	for k, v := range values {
		podCtx := types.TimestampPodContext{}
		if err := json.Unmarshal([]byte(v), &podCtx); err != nil {
			logger.Warnf("failed to unmarshal user pod, %s, %s, just ignore", err, v)
			expirePods = append(expirePods, k)
			continue
		}
		if podCtx.Timestamp < podExpireTime {
			expirePods = append(expirePods, k)
			continue
		}

		results[podCtx.ClusterId] = append(results[podCtx.ClusterId], &podCtx)
	}

	// 清理过期数据
	if len(expirePods) > 0 {
		p.redisClient.HDel(p.ctx, key, expirePods...)
	}

	return results, nil
}

// CleanUserPod 单个集群清理
func (p *CleanUpManager) CleanUserPod() error {
	alivePods, err := p.getActiveUserPod()
	if err != nil {
		return err
	}

	// 只获取固定命名空间
	namespace := GetNamespace()

	if config.G.WebConsole.AdminClusterId != "" {
		values := alivePods[config.G.WebConsole.AdminClusterId]

		alivePodMap := getAlivePodMap(values)
		p.cleanUserPodByCluster(config.G.WebConsole.AdminClusterId, namespace, alivePodMap)
	}

	for clusterId, values := range alivePods {
		if clusterId == config.G.WebConsole.AdminClusterId {
			continue
		}

		alivePodMap := getAlivePodMap(values)
		p.cleanUserPodByCluster(clusterId, namespace, alivePodMap)
	}

	return nil
}

// 清理用户下的相关集群pod
func (p *CleanUpManager) cleanUserPodByCluster(clusterId string, namespace string, alivePodMap map[string]*types.TimestampPodContext) error {
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

		// 小于一个周期的pod不清理
		if expireTime.Before(pod.Status.StartTime.Time) {
			logger.Infof("pod %s exist time %s < %s, just ignore", pod.Name, now.Sub(pod.Status.StartTime.Time), UserPodExpireTime)
			continue
		}

		// 有心跳上报的不清理
		if _, ok := alivePodMap[pod.Name]; ok {
			continue
		}

		// 删除configMap
		if err := p.cleanConfigMapByPod(k8sClient, pod); err != nil {
			logger.Errorf("delete pod(%s) failed, err: %s", pod.Name, err)
			continue
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

// cleanConfigMapByPod 删除configMap
func (p *CleanUpManager) cleanConfigMapByPod(k8sClient *kubernetes.Clientset, pod v1.Pod) error {
	for _, volume := range pod.Spec.Volumes {
		if volume.ConfigMap == nil {
			continue
		}

		if err := k8sClient.CoreV1().ConfigMaps(pod.Namespace).Delete(
			p.ctx,
			volume.ConfigMap.LocalObjectReference.Name,
			metav1.DeleteOptions{},
		); err != nil {
			return errors.Wrapf(err, "delete configmap %s", volume.ConfigMap.LocalObjectReference.Name)
		}
		logger.Infof("delete configmap %s done", volume.ConfigMap.LocalObjectReference.Name)
	}
	return nil
}

func (p *CleanUpManager) Run() error {
	interval := time.NewTicker(CleanUserPodInterval)
	defer interval.Stop()

	sessionCleanupMgr := sessions.NewRedisStore("cleanup", "cleanup")

	for {
		select {
		case <-p.ctx.Done():
			logger.Info("close CleanUpManager done")
			return nil
		case <-interval.C:
			// 清理 pods 数据
			now := time.Now()
			if err := p.CleanUserPod(); err != nil {
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

func getAlivePodMap(pods []*types.TimestampPodContext) map[string]*types.TimestampPodContext {
	alivePodMap := map[string]*types.TimestampPodContext{}
	for _, p := range pods {
		alivePodMap[p.PodName] = p
	}
	return alivePodMap
}
