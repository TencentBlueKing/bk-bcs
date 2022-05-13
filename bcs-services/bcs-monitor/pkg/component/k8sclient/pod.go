/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云-监控平台 (Blueking - Monitor) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package k8sclient

import (
	"bufio"
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 需要取指针, 不用用常量
var (
	// 最多返回 10W 条日志
	MAX_TAIL_LINES = 100000

	// 默认返回 100 条日志
	DEFAULT_TAIL_LINES = int64(100)
)

// LogQuery 日志查询参数， 精简后的 v1.PodLogOptions
type LogQuery struct {
	ContainerName string `form:"container_name"`
	Previous      bool   `form:"previous"`
	StartedAt     string `form:"started_at"`
	FinishedAt    string `form:"finished_at"`
}

func (q *LogQuery) MakeOptions() (*v1.PodLogOptions, error) {
	opt := &v1.PodLogOptions{
		Container: q.ContainerName,
		Previous:  q.Previous,
	}
	if q.StartedAt == "" || q.FinishedAt == "" {
		opt.TailLines = &DEFAULT_TAIL_LINES
	} else {

		// 开始时间, 只做校验
		if _, err := time.Parse(time.RFC3339Nano, q.FinishedAt); err != nil {
			return nil, err
		}

		// 结束时间, 需要用做查询
		t, err := time.Parse(time.RFC3339Nano, q.StartedAt)
		if err != nil {
			return nil, err
		}

		opt.SinceTime = &metav1.Time{Time: t}
	}
	return opt, nil
}

// Log 格式化的日志
type Log struct {
	Log  string `json:"log"`
	Time string `json:"time"`
}

// LogWithPreviousLink
type LogWithPreviousLink struct {
	Logs     []*Log `json:"logs"`
	Previous string `json:"previous"` // 向上翻页链接
}

// Container 格式化的容器, 精简后的 v1.Container
type Container struct {
	Name string `json:"name"`
}

// parseLog 解析Log
func parseLog(rawLog string) (*Log, error) {
	item := strings.SplitN(rawLog, " ", 2)
	if len(item) != 2 {
		return nil, errors.Errorf("invalid log, %s", rawLog)
	}
	return &Log{Log: item[0], Time: item[1]}, nil
}

// GetPodContainers 获取 Pod 容器名称列表
func GetPodContainers(ctx context.Context, clusterId, namespace, podname string) ([]*Container, error) {
	client, err := GetK8SClientByClusterId(clusterId)
	if err != nil {
		return nil, err
	}
	pod, err := client.CoreV1().Pods(namespace).Get(ctx, podname, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := make([]*Container, 0, len(pod.Spec.Containers))
	for _, container := range pod.Spec.Containers {
		containers = append(containers, &Container{Name: container.Name})
	}
	return containers, nil
}

// GetPodLogByte 获取日志
func GetPodLogByte(ctx context.Context, clusterId, namespace, podname string, opt *LogQuery) ([]byte, error) {
	client, err := GetK8SClientByClusterId(clusterId)
	if err != nil {
		return nil, err
	}

	opts, err := opt.MakeOptions()
	if err != nil {
		return nil, err
	}

	result, err := client.CoreV1().Pods(namespace).GetLogs(podname, opts).DoRaw(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetPodLog 获取格式的日志列表
func GetPodLog(ctx context.Context, clusterId, namespace, podname string, opt *LogQuery) (*LogWithPreviousLink, error) {
	result, err := GetPodLogByte(ctx, clusterId, namespace, podname, opt)
	if err != nil {
		return nil, err
	}

	logs := strings.Split(string(result), "\n")
	logList := make([]*Log, 0, len(logs))
	for _, logStr := range logs {
		log, err := parseLog(logStr)
		if err != nil {
			continue
		}
		logList = append(logList, log)
	}

	logResult := &LogWithPreviousLink{Logs: logList}
	return logResult, nil
}

// GetPodLogStream 获取日志流
func GetPodLogStream(ctx context.Context, clusterId, namespace, podname string, opt *LogQuery) (<-chan *Log, error) {
	client, err := GetK8SClientByClusterId(clusterId)
	if err != nil {
		return nil, err
	}

	opts, err := opt.MakeOptions()
	if err != nil {
		return nil, err
	}
	opts.Follow = true

	reader, err := client.CoreV1().Pods(namespace).GetLogs(podname, opts).Stream(ctx)
	if err != nil {
		return nil, err
	}

	logChan := make(chan *Log)
	go func() {
		defer reader.Close()

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			log, err := parseLog(scanner.Text())
			if err != nil {
				continue
			}
			logChan <- log
		}
	}()

	return logChan, nil
}
