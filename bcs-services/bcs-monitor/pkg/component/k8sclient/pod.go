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

package k8sclient

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 需要取指针, 不能用常量
var (
	// 最多返回 10W 条日志
	MAX_TAIL_LINES = 100000

	// 默认返回 100 条日志
	DEFAULT_TAIL_LINES = int64(100)
)

// LogQuery 日志查询参数， 精简后的 v1.PodLogOptions
type LogQuery struct {
	ContainerName string `form:"container_name" binding:"required"` // 必填参数
	Previous      bool   `form:"previous"`
	StartedAt     string `form:"started_at"`
	FinishedAt    string `form:"finished_at"`
}

func (q *LogQuery) makeOptions() (*v1.PodLogOptions, error) {
	opt := &v1.PodLogOptions{
		Container:  q.ContainerName,
		Previous:   q.Previous,
		Timestamps: true,
	}

	if q.StartedAt == "" || q.FinishedAt == "" {
		opt.TailLines = &DEFAULT_TAIL_LINES
	} else {

		// 开始时间, 需要用做查询
		t, err := time.Parse(time.RFC3339Nano, q.StartedAt)
		if err != nil {
			return nil, err
		}

		// 结束时间, 只做校验
		if _, err := time.Parse(time.RFC3339Nano, q.FinishedAt); err != nil {
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

// MakePreviousLink 计算向上翻页链接
func (l *LogWithPreviousLink) MakePreviousLink(projectId, clusterId, namespace, podname string, opt *LogQuery) error {
	if len(l.Logs) <= 2 {
		return nil
	}

	startTime := l.Logs[0].Time
	endTime := l.Logs[len(l.Logs)-1].Time
	sinceTime, err := calcSinceTime(startTime, endTime)
	if err != nil {
		return err
	}

	u := *config.G.Web.BaseURL
	u.Path = path.Join(u.Path, fmt.Sprintf("/projects/%s/clusters/%s/namespaces/%s/pods/%s/logs", projectId, clusterId, namespace, podname))

	query := url.Values{}
	query.Set("started_at", sinceTime.Format(time.RFC3339Nano))
	query.Set("finished_at", startTime) // 本次一次的开始时间做上一页的结束时间
	query.Set("container_name", opt.ContainerName)
	query.Set("previous", strconv.FormatBool(opt.Previous))

	u.RawQuery = query.Encode()

	l.Previous = u.String()

	return nil
}

// calcSinceTime 计算下一次的开始时间
func calcSinceTime(startTime string, endTime string) (*time.Time, error) {
	// 简单场景, 认为日志打印量是均衡的，通过计算时间差获取
	start, err := time.Parse(time.RFC3339Nano, startTime)
	if err != nil {
		return nil, errors.Wrapf(err, "startTime: %s", startTime)
	}
	end, err := time.Parse(time.RFC3339Nano, endTime)
	if err != nil {
		return nil, errors.Wrapf(err, "endTime: %s", endTime)
	}
	duration := end.Sub(start)
	sinceTime := start.Add(-duration)
	return &sinceTime, nil
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
	return &Log{Time: item[0], Log: item[1]}, nil
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

	opts, err := opt.makeOptions()
	if err != nil {
		return nil, err
	}

	result := client.CoreV1().Pods(namespace).GetLogs(podname, opts).Do(ctx)
	if result.Error() != nil {
		return nil, result.Error()
	}

	body, err := result.Raw()
	if err != nil {
		return nil, err
	}

	return body, nil
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

		// 只返回当前历史数据
		if opt.FinishedAt != "" && log.Log == opt.FinishedAt {
			break
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

	opts, err := opt.makeOptions()
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
