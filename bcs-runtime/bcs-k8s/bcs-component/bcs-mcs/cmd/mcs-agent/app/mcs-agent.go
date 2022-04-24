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

package app

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/cmd/mcs-agent/app/options"
	bcsmcsv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/pkg/apis/mcs/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/pkg/controllers"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/pkg/utils"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/pkg/version"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/pkg/version/sharedcommand"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	mcsv1alpha1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"
)

var schema = runtime.NewScheme()

func init() {
	var _ = clientgoscheme.AddToScheme(schema)
	var _ = mcsv1alpha1.AddToScheme(schema)
	var _ = bcsmcsv1alpha1.AddToScheme(schema)
}

// NewAgentCommand creates a *cobra.Command object with default parameters
func NewAgentCommand(ctx context.Context) *cobra.Command {
	opts := options.NewOptions()
	cmd := &cobra.Command{
		Use:  "mcs-agent",
		Long: `mcs-agent is a tool for managing mcs-agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// validate options
			if errs := opts.Validate(); len(errs) != 0 {
				return errs.ToAggregate()
			}
			if err := run(ctx, opts); err != nil {
				return err
			}
			return nil
		},
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
	}

	opts.AddFlags(cmd.Flags())
	cmd.AddCommand(sharedcommand.NewCmdVersion(os.Stdout, "mcs-agent"))
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	return cmd
}

func run(ctx context.Context, opts *options.Options) error {
	klog.Infof("mcs-agent version: %s", version.Get())

	restConfig := controllerruntime.GetConfigOrDie()

	controllerManager, err := controllerruntime.NewManager(restConfig, controllerruntime.Options{
		Scheme:                     schema,
		LeaderElection:             opts.LeaderElection.LeaderElect,
		LeaderElectionID:           fmt.Sprintf("bcs-mcs-agent-%s", opts.AgentID),
		LeaderElectionNamespace:    opts.LeaderElection.ResourceNamespace,
		LeaderElectionResourceLock: opts.LeaderElection.ResourceLock,
		HealthProbeBindAddress:     net.JoinHostPort(opts.BindAddress, strconv.Itoa(opts.HealthCheckPort)),
		MetricsBindAddress:         net.JoinHostPort(opts.BindAddress, strconv.Itoa(opts.MetricsPort)),
	})
	if err != nil {
		klog.Errorf("failed to build controller manager: %v", err)
		return err
	}

	if err := controllerManager.AddHealthzCheck("ping", healthz.Ping); err != nil {
		klog.Errorf("failed to add health check endpoint: %v", err)
		return err
	}
	if err := controllerManager.AddReadyzCheck("ping", healthz.Ping); err != nil {
		klog.Errorf("failed to add ready check endpoint: %v", err)
		return err
	}

	kubeClient := kubernetes.NewForConfigOrDie(restConfig)
	var parentKubeClient kubernetes.Interface
	var parentCluster cluster.Cluster
	if opts.ParentKubeconfigPath != "" {
		parentRestConfig, err := clientcmd.BuildConfigFromFlags("", opts.ParentKubeconfigPath)
		if err != nil {
			klog.Fatalf("build rest config error %+v", err)
		}
		parentCluster, err = cluster.New(parentRestConfig, func(o *cluster.Options) {
			o.Scheme = schema
		})
		if err != nil {
			klog.Errorf("failed to create parentCluster, err=%v", err)
			return err
		}
		if err := controllerManager.Add(parentCluster); err != nil {
			klog.Errorf("failed to add cluster: %v", err)
			return err
		}
		parentKubeClient = kubernetes.NewForConfigOrDie(parentRestConfig)
	} else {
		// 当不指定父级集群时，默认使用当前集群作为父集群
		parentCluster = controllerManager
		parentKubeClient = kubeClient
	}
	err = createParentClusterNamespace(ctx, parentKubeClient, opts.AgentID)
	if err != nil {
		klog.ErrorS(err, "create parent cluster namespace error")
		return err
	}

	setupControllers(controllerManager, opts, parentCluster)

	if err := controllerManager.Start(ctx); err != nil {
		klog.Errorf("controller manager exits unexpectedly: %v", err)
		return err
	}
	return nil
}

func setupControllers(mgr controllerruntime.Manager, opts *options.Options, parentCluster cluster.Cluster) {
	if err := (&controllers.ServiceExportController{
		Client:              mgr.GetClient(),
		AgentID:             opts.AgentID,
		ParentClusterClient: parentCluster.GetClient(),
		EventRecorder:       mgr.GetEventRecorderFor(controllers.ServiceExportControllerName),
	}).SetupWithManager(mgr); err != nil {
		klog.Fatalf("unable to create %s controller, err=%+v", controllers.ServiceExportControllerName, err)
	}

	if err := (&controllers.ServiceImportController{
		Client:              mgr.GetClient(),
		AgentID:             opts.AgentID,
		ParentClusterClient: parentCluster.GetClient(),
		EventRecorder:       mgr.GetEventRecorderFor(controllers.ServiceImportControllerName),
	}).SetupWithManager(mgr, parentCluster); err != nil {
		klog.Fatalf("unable to create %s controller, err=%+v", controllers.ServiceImportControllerName, err)
	}

	if err := (&controllers.ServiceController{
		Client: mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		klog.Fatalf("unable to create %s controller, err=%+v", controllers.ServiceControllerName, err)
	}
}

func createParentClusterNamespace(ctx context.Context, client kubernetes.Interface, agentID string) error {
	namespaceName := utils.GenManifestNamespace(agentID)
	// 创建父级集群的namespace
	parentClusterNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	_, err := client.CoreV1().Namespaces().Get(ctx, namespaceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = client.CoreV1().Namespaces().Create(ctx, parentClusterNamespace, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			klog.Infof("create parent cluster namespace %s success", namespaceName)
			return nil
		}
		return err
	}
	return nil
}
