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

// Package types xxx
package types

import (
	"io"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/k8sclient"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// PodExec pod exec context
type PodExec struct {
	K8sClient     *kubernetes.Clientset
	RESTConfig    *rest.Config
	Namespace     string
	PodName       string
	ContainerName string
	Command       []string
	Stdin         io.Reader
	Stdout        io.Writer
	Stderr        io.Writer
	Tty           bool
	NoPreserve    bool
}

// NewPodExec 创建 PodExec
func (p *PodContext) NewPodExec() (*PodExec, error) {
	clientset, err := k8sclient.GetK8SClientByClusterId(p.AdminClusterId)
	if err != nil {
		return nil, err
	}
	restConfig := k8sclient.GetK8SConfigByClusterId(p.AdminClusterId)
	return &PodExec{
		Namespace:     p.Namespace,
		PodName:       p.PodName,
		ContainerName: p.ContainerName,
		K8sClient:     clientset,
		RESTConfig:    restConfig,
	}, nil
}

// Exec 在给定容器中执行命令
func (p *PodExec) Exec() error {
	req := p.K8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(p.PodName).
		Namespace(p.Namespace).
		SubResource("exec").
		VersionedParams(&coreV1.PodExecOptions{
			Command:   p.Command,
			Container: p.ContainerName,
			Stdin:     p.Stdin != nil,
			Stdout:    p.Stdout != nil,
			Stderr:    p.Stderr != nil,
			TTY:       p.Tty,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(p.RESTConfig, "POST", req.URL())
	if err != nil {
		return err
	}
	var sizeQueue remotecommand.TerminalSizeQueue
	return exec.Stream(remotecommand.StreamOptions{
		Stdin:             p.Stdin,
		Stdout:            p.Stdout,
		Stderr:            p.Stderr,
		Tty:               p.Tty,
		TerminalSizeQueue: sizeQueue,
	})
}
