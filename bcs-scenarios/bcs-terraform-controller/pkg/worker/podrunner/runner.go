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

// Package podrunner 定义 terraform-worker 的接口
package podrunner

import (
	"context"
	"time"

	"github.com/pkg/errors"
	appv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/option"
)

const (
	// EnvPodIndex 环境变量中的 Pod 索引
	EnvPodIndex = "POD_INDEX"
)

// Runner 定义 terraform-worker 的接口
type Runner interface {
	Init(ctx context.Context) error
}

type workerRunner struct {
	op        *option.ControllerOption
	k8sClient *kubernetes.Clientset
}

// NewRunner 创建 Runner 实例
func NewRunner() Runner {
	return &workerRunner{}
}

// Init 检测 terraform worker 的启动是否完成
func (r *workerRunner) Init(ctx context.Context) error {
	r.op = option.GlobalOption()
	config, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrapf(err, "get in-cluster config failed")
	}
	r.k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "create in-cluster client failed")
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, time.Duration(300)*time.Second)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			var state *appv1.StatefulSet
			state, err = r.k8sClient.AppsV1().StatefulSets(r.op.WorkerNamespace).Get(ctx, r.op.WorkerName,
				metav1.GetOptions{})
			if err != nil && !k8serrors.IsNotFound(err) {
				return errors.Wrapf(err, "get statefulset '%s/%s' failed", state.Namespace, state.Name)
			}
			if k8serrors.IsNotFound(err) {
				logctx.Infof(ctx, "wait for worker statefulset '%s/%s' create...",
					state.Namespace, state.Name)
				continue
			}
			if state.Status.ReadyReplicas != *state.Spec.Replicas {
				logctx.Infof(ctx, "worker is still starting...(%d/%d)",
					state.Status.ReadyReplicas, *state.Spec.Replicas)
				continue
			}
			return nil
		case <-ctx.Done():
			return errors.Errorf("worker start with context timeout")
		}
	}
}
