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

// Package main xx
package main

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/internal/utils"
	gitopsv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/apis/gitopsworkflow/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/controller"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/option"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(gitopsv1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {
	option.Parse()
	op := option.GlobalOption()
	blog.InitLogs(op.LogConfig)
	defer blog.CloseLogs()
	blog.Infof("option: %v", utils.ToJsonString(op))

	ctx := ctrl.SetupSignalHandler()
	mgr, err := buildControllerManager(op)
	if err != nil {
		blog.Fatalf("build controller manager failed: %s", err.Error())
	}
	if err = mgr.Start(ctx); err != nil {
		blog.Errorf("controller manager exit: %s", err.Error())
	} else {
		blog.Infof("controller manager is stopped")
	}
}

func buildControllerManager(op *option.ControllerOption) (manager.Manager, error) {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Port:                   op.Port,
		MetricsBindAddress:     fmt.Sprintf("0.0.0.0:%d", op.MetricPort),
		HealthProbeBindAddress: fmt.Sprintf("0.0.0.0:%d", op.HealthPort),
		LeaderElection:         op.EnableLeaderElection,
		LeaderElectionID:       "gitops-workflow.bkbcs.tencent.com",
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
		NewClient: func(cache cache.Cache, config *rest.Config, options client.Options,
			uncachedObjects ...client.Object) (client.Client, error) {
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

	if err = controller.NewWorkflowController(mgr.GetClient(), mgr.GetScheme()).SetupWithManager(mgr); err != nil {
		return nil, errors.Wrapf(err, "setup controller manager failed")
	}
	if err = controller.NewHistoryController(mgr.GetClient(), mgr.GetScheme()).SetupWithManager(mgr); err != nil {
		return nil, errors.Wrapf(err, "setup histroy controller failed")
	}
	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return nil, errors.Wrapf(err, "unable to set up health check")
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return nil, errors.Wrapf(err, "unable to set up ready check")
	}
	return mgr, nil
}
