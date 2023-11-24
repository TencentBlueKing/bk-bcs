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
	"bytes"
	"context"
	"io"
	"time"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/repository"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

// HistoryMgr .bash_history mgr
type HistoryMgr struct {
	stor repository.Provider
}

var historyMgr *HistoryMgr

// InitHistoryMgr init when config ready
func InitHistoryMgr() {
	// 使用系统对象存储
	stor, err := repository.NewProvider(config.G.Repository.StorageType)
	if err != nil {
		logger.Warnf("init historyMgr fail: %s, operator will silence ignore", err.Error())
	}

	historyMgr = &HistoryMgr{stor: stor}
}

// createBashHistory 创建.bash_history文件
func (h *HistoryMgr) createBashHistory(podCtx *types.PodContext) error {
	// 未配置 repo, 忽略
	if h.stor == nil {
		return nil
	}

	// 往容器写文件
	pe, err := podCtx.NewPodExec()
	if err != nil {
		return err
	}
	pe.Command = []string{"cp", "/dev/stdin", "/root/.bash_history"}
	stdin := &bytes.Buffer{}
	pe.Stderr = &bytes.Buffer{}
	// 读取保存的.bash_history文件
	historyFileByte, err := h.getBashHistory(podCtx.PodName)
	if err != nil {
		return err
	}
	_, err = stdin.Write(historyFileByte)
	if err != nil {
		return err
	}
	pe.Stdin = stdin
	err = pe.Exec()
	if err != nil {
		return err
	}
	return nil
}

// getBashHistory直接读取存储在远程repo中的文件
func (h *HistoryMgr) getBashHistory(podName string) ([]byte, error) {
	filename := podName + historyFileName

	// 5秒下载时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// 下载对象储存文件
	fileInput, err := h.stor.DownloadFile(ctx, historyRepoDir+filename)
	if err != nil {
		return nil, err
	}

	// 读取.bash_history内容byte格式
	output, err := io.ReadAll(fileInput)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// persistenceHistory 命令行持久化保存
func (h *HistoryMgr) persistenceBashHistory(
	k8sClient *kubernetes.Clientset, podName, namespace, containerName, clusterId string) error {
	// 未配置 repo, 忽略
	if h.stor == nil {
		return nil
	}

	req := k8sClient.CoreV1().RESTClient().Post().Resource("pods").Name(podName).Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Command:   []string{"/bin/bash", "-c", "cat /root/.bash_history"},
		Container: containerName,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true, // kubectl 默认 stderr 未设置, virtual-kubelet 节点 stderr 和 tty 不能同时为 true
		TTY:       false,
	}, scheme.ParameterCodec)

	k8sConfig := k8sclient.GetK8SConfigByClusterId(clusterId)
	executor, err := remotecommand.NewSPDYExecutor(k8sConfig, "POST", req.URL())
	if err != nil {
		return err
	}

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return err
	}

	// 文件名
	filename := podName + historyFileName

	// 10秒上传超时时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 推送至远程repo对象存储
	err = h.stor.UploadFileByReader(ctx, stdout, historyRepoDir+filename)
	if err != nil {
		return err
	}
	return nil
}
