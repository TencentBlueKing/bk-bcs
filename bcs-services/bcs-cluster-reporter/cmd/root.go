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

// Package cmd xxx
package cmd

import (
	"context"
	"flag"
	"fmt"
	_ "net/http/pprof" // pprof
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/cmd/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/api/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

var (
	bcro = options.NewBcsClusterReporterOptions()

	rootCmd = &cobra.Command{
		Use:   "tcctl",
		Short: "bcs-cluster-reporter",
		Long: `
Basic Commands (Beginner):
	Get        Create a resource from a file or from stdin
`, // nolint
		Run: func(cmd *cobra.Command, args []string) {
			CheckErr(Complete(cmd, args))
			metric_manager.MM.RunPrometheusMetricsServer()

			err := Run()
			if err != nil {
				klog.Fatalf("bcs-cluster-reporter failed: %s", err.Error())
			}
		},
	}
)

// Run main process
func Run() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	configFileBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return err
	}

	id, err := os.Hostname() // os.Getenv("POD_NAME")
	if err != nil {
		return err
	}

	leaseName := os.Getenv("DEPLOY_NAME")
	if leaseName == "" {
		leaseName = "bcs-cluster-reporter"
	}
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      leaseName,
			Namespace: string(configFileBytes),
		},
		Client: client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: id,
		},
	}

	// 进行选举
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
				Lock:          lock,
				LeaseDuration: 30 * time.Second,
				RenewDeadline: 15 * time.Second,
				RetryPeriod:   5 * time.Second,
				Callbacks: leaderelection.LeaderCallbacks{
					OnStartedLeading: run,
					OnStoppedLeading: func() {
						klog.Infof("leader lost: %s", id)
						cancel()
						ctx, cancel = context.WithCancel(context.Background())
						// NOCC:vet/vet(忽略)
					},
				},
			})
		}
	}()

	// 退出清理
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	// stop plugins
	klog.Infof("start to shutdown bcs-cluster-reporter")
	pprof.Lookup("heap").WriteTo(os.Stdout, 1) // nolint
	cancel()

	return nil
}

func run(ctx context.Context) {
	getClusters()

	go func() {
		select {
		case <-ctx.Done():
			break
		default:
			for {
				time.Sleep(time.Minute * 1)
				getClusters()
			}
		}
	}()

	// start plugins
	err := plugin_manager.Pm.SetupPlugin(bcro.Plugins, bcro.PluginConfDir)
	if err != nil {
		klog.Fatalf("Setup plugin failed: %s", err.Error())
	}

	klog.Info("Setup plugins success")

	<-ctx.Done()
	// 停止模块的运行
	klog.Infof("start to stop plugins")
	err = plugin_manager.Pm.StopPlugin(bcro.Plugins)
	if err != nil {
		klog.Fatalf("Setup plugin failed: %s", err.Error())
	}
	klog.Infof("done stop plugins")
}

// Execute rootCmd
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		klog.Fatalf(err.Error())
	}
}

// CheckErr check err
func CheckErr(err error) {
	if err != nil {
		klog.Fatalf(err.Error())
	}
}

// Complete xxx
func Complete(cmd *cobra.Command, args []string) error {
	if bcro.BcsClusterManagerToken != "" || bcro.BcsClusterManagerApiserver != "" || bcro.BcsGatewayApiserver != "" ||
		bcro.BcsGatewayToken != "" {
		if bcro.BcsClusterManagerToken == "" || bcro.BcsClusterManagerApiserver == "" || bcro.BcsGatewayApiserver == "" ||
			bcro.BcsGatewayToken == "" {
			return fmt.Errorf(
				"bcs config missing, BcsClusterManagerToken, BcsClusterManagerApiserver, BcsGatewayApiserver, " +
					"BcsGatewayToken must be set")
		}
		bcro.BcsClusterManagerToken = util.Decode(bcro.BcsClusterManagerToken)
		bcro.BcsGatewayToken = util.Decode(bcro.BcsGatewayToken)
	}

	if (bcro.BcsGatewayApiserver != "" && bcro.BcsClusterManagerApiserver != "" && bcro.BcsGatewayToken != "" &&
		bcro.BcsClusterManagerToken != "") && bcro.InCluster {
		return fmt.Errorf("when run in in-cluster mode, no need to set bcs params")
	}

	if bcro.KubeConfigDir != "" && bcro.InCluster {
		return fmt.Errorf("when run in in-cluster mode, no need to set kubeConfigDir")
	}

	if bcro.InCluster && (bcro.ClusterID == "" || bcro.BizID == "") {
		return fmt.Errorf("when run in in-cluster mode, need to set clusterID and bizID")
	}
	return nil
}

func init() {
	flags := rootCmd.PersistentFlags()
	bcro.AddFlags(flags)

	cobra.OnInitialize(initConfig)

	// init klog flag
	fs := flag.NewFlagSet("", flag.PanicOnError)
	klog.InitFlags(fs)
	rootCmd.PersistentFlags().AddGoFlagSet(fs)

	err := viper.BindPFlags(flags)
	if err != nil {
		klog.Fatalf("Viper bindPFlags failed: %s", err.Error())
	}

}

// initConfig
// configure viper to read config
func initConfig() {}

// nolint funlen
func getClusters() {
	clusterConfigList := make([]plugin_manager.ClusterConfig, 0, 0)

	// 获取BCS集群配置
	if bcro.BcsGatewayApiserver != "" && bcro.BcsClusterManagerApiserver != "" && bcro.BcsGatewayToken != "" &&
		bcro.BcsClusterManagerToken != "" {
		bcsClusterManager, err := bcs.NewClusterManager(bcro.BcsClusterManagerToken, bcro.BcsClusterManagerApiserver,
			bcro.BcsGatewayApiserver, bcro.BcsGatewayToken)
		if err != nil {
			klog.Fatalf("NewClusterManager failed: %s", err.Error())
		}

		clusterList, err := bcsClusterManager.GetClusters([]string{})
		if err != nil {
			klog.Errorf("NewClusterManager failed: %s", err.Error())
			return
		}

		filteredClusterList := make([]cmproto.Cluster, 0, 0)
		if len(bcro.BcsClusterList) != 0 {
			for _, clusterId := range bcro.BcsClusterList {
				for _, cluster := range clusterList {
					if clusterId == cluster.ClusterID {
						filteredClusterList = append(filteredClusterList, cluster)
					}
					break
				}
			}
		} else {
			for _, cluster := range clusterList {
				if cluster.IsShared == true {
					continue // 跳过公共集群的记录
				} else if cluster.Status != "RUNNING" {
					continue // 跳过未就绪集群
				} else if cluster.EngineType != "k8s" {
					continue
				} else {
					if len(cluster.Master) > 0 {
						continueFlag := false
						for masterName := range cluster.Master {
							if strings.Contains(masterName, "127.0.0") {
								// 跳过算力集群
								continueFlag = true
								break
							}
						}
						if continueFlag {
							continue
						}
					}

					// 选取对应类型的集群
					if (cluster.Environment == bcro.BcsClusterType && bcro.BcsClusterType != "") || bcro.BcsClusterType == "" {
						if cluster.CreateTime != "" {
							createTime, err := time.Parse(time.RFC3339, cluster.CreateTime)
							if err != nil {
								klog.Errorf("parse cluster %s createtime failed %s", cluster.ClusterID, err.Error())
								continue
							}
							// 创建时间超过60分钟才进行巡检
							if (time.Now().Unix() - createTime.Unix()) > 60*60 {
								filteredClusterList = append(filteredClusterList, cluster)
							}
						}
					}
				}
			}
		}

		for _, cluster := range filteredClusterList {
			clusterConfigList = append(clusterConfigList, plugin_manager.ClusterConfig{
				ClusterID:  cluster.ClusterID,
				Config:     bcsClusterManager.GetKubeconfig(cluster.ClusterID),
				BusinessID: cluster.BusinessID,
			})
		}
	}

	// 获取kubeconfig的配置
	if bcro.KubeConfigDir != "" {
		var filePathList []string
		err := filepath.Walk(bcro.KubeConfigDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.Mode()&os.ModeSymlink != 0 {
				return nil
			}

			if !info.IsDir() && strings.HasSuffix(info.Name(), "config") {
				filePathList = append(filePathList, path)
			}
			return nil
		})
		if err != nil {
			klog.Fatalf("Error: %s", err.Error())
			return
		}

		for _, filePath := range filePathList {
			config, err := k8s.GetRestConfigByConfig(filePath)
			if err != nil {
				klog.Fatalf("Error: %s", err.Error())
				return
			}

			config.TLSClientConfig.Insecure = true

			filenameWithExt := filepath.Base(filePath) // 获取文件名（包括后缀）
			ext := filepath.Ext(filenameWithExt)       // 获取文件后缀

			filename := strings.TrimSuffix(filenameWithExt, ext) // 移除后缀
			//  读取配置文件时没配置bizid
			clusterConfigList = append(clusterConfigList, plugin_manager.ClusterConfig{BusinessID: "0", ClusterID: filename,
				Config: config})
			klog.Infof("load kubeconfig success, clusterID: %s", filename)
		}
	}

	// Incluster模式
	if bcro.InCluster {
		config, err := rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("Error: %s", err.Error())
			return
		}
		clusterConfigList = append(clusterConfigList, plugin_manager.ClusterConfig{BusinessID: bcro.BizID,
			ClusterID: bcro.ClusterID, Config: config})
		plugin_manager.Pm.SetConfig(&plugin_manager.Config{
			ClusterConfigs:  clusterConfigList,
			InClusterConfig: plugin_manager.ClusterConfig{BusinessID: bcro.BizID, ClusterID: bcro.ClusterID, Config: config},
		})
	} else {
		// 集中化模式
		plugin_manager.Pm.SetConfig(&plugin_manager.Config{
			ClusterConfigs: clusterConfigList,
		})
	}
}
