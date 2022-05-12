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
	"context"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetContainerNames 获取 Pod 容器名称列表
func GetContainerNames(ctx context.Context, clusterId, namespace, podname string) ([]string, error) {
	client, err := GetK8SClientByClusterId(clusterId)
	if err != nil {
		return nil, err
	}
	pod, err := client.CoreV1().Pods(namespace).Get(ctx, podname, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := make([]string, 0, len(pod.Status.ContainerStatuses))
	for _, container := range pod.Status.ContainerStatuses {
		containers = append(containers, container.Name)
	}
	return containers, nil
}

// LogQuery 日志查询参数
type LogQuery struct {
	ContainerName string `form:"container_name"`
	Previous      bool   `form:"previous"`
}

type Log struct {
	Log  string `json:"log"`
	Time string `json:"time"`
}

// GetContainerNames 获取 Pod 容器名称列表
func GetContainerLog(ctx context.Context, clusterId, namespace, podname string, opt *LogQuery) ([]*Log, error) {
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

	logs := strings.Split(string(result), "\n")
	logResult := make([]*Log, 0, len(logs))
	for _, logStr := range logs {
		item := strings.SplitN(logStr, " ", 2)
		if len(item) != 2 {
			continue
		}
		logResult = append(logResult, &Log{Log: item[0], Time: item[1]})
	}

	return logResult, nil
}
