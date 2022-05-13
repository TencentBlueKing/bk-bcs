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

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LogQuery 日志查询参数
type LogQuery struct {
	ContainerName string `form:"container_name"`
	Previous      bool   `form:"previous"`
}

// Log 格式化的日志
type Log struct {
	Log  string `json:"log"`
	Time string `json:"time"`
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
	limitByte := int64(10 * 1024 * 1024)
	tailLines := int64(100)
	opts := &v1.PodLogOptions{
		Container:  opt.ContainerName,
		Previous:   opt.Previous,
		LimitBytes: &limitByte,
		TailLines:  &tailLines,
		Timestamps: true,
	}
	result, err := client.CoreV1().Pods(namespace).GetLogs(podname, opts).DoRaw(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetPodLog 获取格式的日志列表
func GetPodLog(ctx context.Context, clusterId, namespace, podname string, opt *LogQuery) ([]*Log, error) {
	result, err := GetPodLogByte(ctx, clusterId, namespace, podname, opt)
	if err != nil {
		return nil, err
	}

	logs := strings.Split(string(result), "\n")
	logResult := make([]*Log, 0, len(logs))
	for _, logStr := range logs {
		log, err := parseLog(logStr)
		if err != nil {
			continue
		}
		logResult = append(logResult, log)
	}

	return logResult, nil
}

// GetPodLogStream 获取日志流
func GetPodLogStream(ctx context.Context, clusterId, namespace, podname string, opt *LogQuery) (<-chan *Log, error) {
	client, err := GetK8SClientByClusterId(clusterId)
	if err != nil {
		return nil, err
	}
	limitByte := int64(10 * 1024 * 1024)
	tailLines := int64(100)
	opts := &v1.PodLogOptions{
		Container:  opt.ContainerName,
		LimitBytes: &limitByte,
		TailLines:  &tailLines,
		Timestamps: true,
		Follow:     true,
	}
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
