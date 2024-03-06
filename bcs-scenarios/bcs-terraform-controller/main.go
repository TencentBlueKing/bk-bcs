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

package main

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/go-git/go-git/v5/plumbing"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/controllers"
	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/terraformextensions/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/repository"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/tfhandler"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/utils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/worker"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

// 端口变更！参考deployment.yaml
func init() {
	utilruntime.Must(tfv1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	// git.Repository的ResolveRevision()方法描述与实际实现不同，缺少origin的情况的实现，故在这里加上origin条件
	plumbing.RefRevParseRules = append(plumbing.RefRevParseRules, "refs/remotes/origin/%s")
}

func main() {
	if err := option.Parse(); err != nil {
		panic(err)
	}
	op := option.GlobalOption()
	blog.InitLogs(op.LogConfig)
	defer blog.CloseLogs()
	blog.Infof("option: %v", utils.ToJsonString(op))

	ctx := ctrl.SetupSignalHandler()
	if !op.IsWorker {
		startController(ctx, op)
	} else {
		startWorker(ctx, op)
	}
}

func startWorker(ctx context.Context, op *option.ControllerOption) {
	tfWorker := &worker.TerraformWorker{}
	if err := tfWorker.Init(ctx); err != nil {
		blog.Fatalf("init terraform worker failed: %s", err.Error())
	}
	tfWorker.Start(ctx)
	blog.Warnf("terraform worker is finished")
}

// startController 运行 controller:
// - ControllerManager: 接收 terraform cr 事件，并做对应处理
// - TerraformServer: 运行 grpc server, 负责分配处理 cr 的队列
func startController(ctx context.Context, op *option.ControllerOption) {
	tfServer := worker.NewTerraformServer()
	if err := tfServer.Init(ctx); err != nil {
		blog.Fatalf("init terraform server failed: %s", err.Error())
	}
	defer tfServer.Stop()
	mgr, err := buildControllerManager(ctx, op, tfServer)
	if err != nil {
		blog.Fatalf("build controller manager failed: %s", err.Error())
	}

	closeCh := make(chan struct{})
	go runControllerManager(ctx, mgr, closeCh)
	go runTerraformServer(ctx, tfServer, closeCh)
	select {
	case <-ctx.Done():
		blog.Warnf("received shutdown signal")
	case _, ok := <-closeCh:
		if !ok {
			blog.Errorf("close channel is closed")
			break
		}
		blog.Infof("received from close channel")
	}
}

func runControllerManager(ctx context.Context, mgr manager.Manager, closeCh chan struct{}) {
	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		blog.Errorf("controller manager running occurred an err: %s", err.Error())
	}
	blog.Infof("controller manager is stopped")
	closeCh <- struct{}{}
}

func runTerraformServer(ctx context.Context, tfServer *worker.TerraformServer, closeCh chan struct{}) {
	if err := tfServer.Start(ctx); err != nil {
		blog.Errorf("terraform rpc server start failed: %s", err.Error())
	}
	blog.Infof("tfWorker is stopped")
	closeCh <- struct{}{}
}

func buildControllerManager(ctx context.Context, op *option.ControllerOption,
	tfServer *worker.TerraformServer) (manager.Manager, error) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// Metrics:
		// Port:                   9443,
		HealthProbeBindAddress: "0.0.0.0:8083",
		LeaderElection:         op.EnableLeaderElection,
		LeaderElectionID:       "gitops-terraform.bkbcs.tencent.com",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
		NewClient: func(cache cache.Cache, config *rest.Config, options client.Options, uncachedObjects ...client.Object) (client.Client, error) {
			config.QPS = float32(op.KubernetesQPS)
			config.Burst = op.KubernetesBurst
			// Create the Client for Write operations.
			c, err := client.New(config, options)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "create controller manager failed")
	}

	tfHandler, err := buildTerraformHandler(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "build terraform handler failed")
	}
	if err = (&controllers.TerraformReconciler{
		Config:    op,
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Queue:     tfServer,
		TFHandler: tfHandler,
	}).SetupWithManager(mgr); err != nil {
		return nil, errors.Wrapf(err, "setup controller manager failed")
	}
	// +kubebuilder:scaffold:builder

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return nil, errors.Wrapf(err, "unable to set up health check")
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return nil, errors.Wrapf(err, "unable to set up ready check")
	}
	return mgr, nil
}

func buildTerraformHandler(ctx context.Context) (tfhandler.TerraformHandler, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "get in-cluster config failed")
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "create in-cluster client failed")
	}
	argoDB, _, err := store.NewArgoDB(ctx, option.GlobalOption().ArgoAdminNamespace)
	if err != nil {
		return nil, errors.Wrapf(err, "create argo db failed")
	}
	repoHandler := repository.NewRepositoryHandler(argoDB)
	tfHandler := tfhandler.NewTerraformHandler(repoHandler, k8sClient)
	return tfHandler, nil
}
