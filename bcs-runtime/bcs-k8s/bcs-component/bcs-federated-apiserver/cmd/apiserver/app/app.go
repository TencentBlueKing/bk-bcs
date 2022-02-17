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
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-federated-apiserver/pkg/apiserver"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-federated-apiserver/pkg/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-federated-apiserver/pkg/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	cmdName = "apiserver"
)

// NewAggregationCommand creates a new command for running the aggregation apiserver.
func NewAggregationCommand() *cobra.Command {
	opts := options.NewAggregationOptions()

	cmd := &cobra.Command{
		Use:  cmdName,
		Long: `Running in parent cluster, responsible for multiple cluster managements`,
		Run: func(cmd *cobra.Command, args []string) {

			if err := opts.Complete(); err != nil {
				klog.Exit(err)
			}
			if err := opts.Validate(args); err != nil {
				klog.Exit(err)
			}

			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				klog.V(1).Infof("FLAG: --%s=%q", flag.Name, flag.Value)
			})

			apiserverConfig, err := opts.APIServerConfig()
			if err != nil {
				klog.Fatalf("Failed to create config: %v", err)
			}
			completedConfig := apiserverConfig.Complete()

			//genericAPIServer
			genericServer, err := completedConfig.GenericConfig.New("bcs-federated-apiserver", genericapiserver.NewEmptyDelegate())
			if err != nil {
				klog.Fatalf("Error in initializing configuration: %v", err)
			}
			c := make(chan struct{})
			completedConfig.GenericConfig.SharedInformerFactory.Start(c)

			//bcsStorage
			bcsStorage := storage.NewBcsStorage(opts.GetConfig().BcsStorageAddress, opts.GetConfig().BcsStorageToken, opts.GetConfig().BcsStorageURLPrefix)
			genericServer.AddPostStartHookOrDie("start-federated-aggregationapis", func(context genericapiserver.PostStartHookContext) error {
				if genericServer != nil {
					cfg, err := clientcmd.BuildConfigFromFlags("", "")
					if err != nil {
						return err
					}
					kubeclient, err := kubernetes.NewForConfig(cfg)
					ss := apiserver.NewAggregationAPIServer(genericServer,
						opts.GetConfig(),
						int64(3*1024*1024),
						1800,
						completedConfig.GenericConfig.AdmissionControl,
						kubeclient.RESTClient(),
						bcsStorage,
					)
					return ss.InstallShadowAPIGroups(c, kubeclient.DiscoveryClient)

				}

				select {
				case <-context.StopCh:
				}

				return nil
			})

			err = genericServer.PrepareRun().Run(c)
			if err != nil {
				klog.Fatalf("Error in running generic api server: %v", err)
			}
			// TODO: add logic
		},
	}

	flags := cmd.Flags()
	opts.AddFlags(flags)

	return cmd
}
