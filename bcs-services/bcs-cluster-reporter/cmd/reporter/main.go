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

// Package main xxx
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/cmd/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/api/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/capacitycheck"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/clustercheck"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodecheck"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/systemappcheck"
)

var (
	bcro = options.NewBcsClusterReporterOptions()

	rootCmd = &cobra.Command{
		Use:   "bcs-cluster-reporter",
		Short: "bcs-cluster-reporter",
		Long: `
Basic Commands (Beginner):
	bcs-cluster-reporter
`,
		Run: func(cmd *cobra.Command, args []string) {
			CheckErr(Complete(cmd, args))

			err := Run()
			if err != nil {
				klog.Fatalf("bcs-cluster-reporter failed: %s", err.Error())
			}
		},
	}
)

// Run main process
func Run() error {
	run(context.Background())

	return nil
}

func run(ctx context.Context) {
	r := gin.Default()
	pprof.Register(r)
	go func() {
		if err := r.Run(":6216"); err != nil {
			klog.Fatalf(err.Error())
		}
	}()

	getClusters()

	go func() {
		select {
		case <-ctx.Done():
			break
		default:
			for {
				time.Sleep(time.Minute * 30)
				getClusters()
			}
		}
	}()

	// start plugins
	err := plugin_manager.Pm.SetupPlugin(bcro.Plugins, bcro.PluginConfDir, bcro.RunMode)
	if err != nil {
		klog.Fatalf("Setup plugin failed: %s", err.Error())
	}

	klog.Info("Setup plugins success")

	// start webserver
	if bcro.RunMode == plugin_manager.RunModeDaemon {
		r.GET("cluster/:clusterID/pdf", func(c *gin.Context) {
			clusterID := c.Param("clusterID")
			pdf, reportErr := plugin_manager.GetClusterReport(clusterID, bcro.Plugins)
			if reportErr != nil {
				c.String(404, fmt.Sprintf("cluster %s not found", clusterID))
				return
			}
			err = pdf.Output(c.Writer)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to generate PDF")
				klog.Errorf(err.Error())
				return
			}
			c.Header("Content-Type", "application/pdf")
			c.Header("Content-Disposition", "attachment; filename=output.pdf")
			return
			// 将PDF内容写入HTTP响应
		})

		r.GET("biz/:bizID/pdf", func(c *gin.Context) {
			bizID := c.Param("bizID")

			pdf, reportErr := plugin_manager.GetBizReport(bizID, bcro.Plugins)
			if reportErr != nil {
				c.String(404, fmt.Sprintf("biz %s not found", bizID))
				return
			}

			err = pdf.Output(c.Writer)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to generate PDF")
				klog.Errorf(err.Error())
				return
			}
			c.Header("Content-Type", "application/pdf")
			c.Header("Content-Disposition", "attachment; filename=output.pdf")
			return
		})

		r.GET("cluster/:clusterID/html", func(c *gin.Context) {
			clusterID := c.Param("clusterID")

			html, htmlErr := plugin_manager.GetClusterReportHtml(clusterID, bcro.Plugins)
			if htmlErr != nil {
				c.String(404, fmt.Sprintf("cluster %s not found", clusterID))
				return
			}

			c.String(200, html)
			return
		})

		r.GET("biz/:bizID/html", func(c *gin.Context) {
			bizID := c.Param("bizID")

			html, htmlErr := plugin_manager.GetBizReportHtml(bizID, bcro.Plugins)
			if htmlErr != nil {
				c.String(404, fmt.Sprintf("biz %s not found", bizID))
				return
			}

			c.String(200, html)
			return
		})

		r.GET("/metrics", gin.WrapH(promhttp.Handler()))

		// config mm
		metric_manager.MM.SetEngine(r)
	} else if bcro.RunMode == plugin_manager.RunModeOnce {
		for _, clusterConfig := range plugin_manager.Pm.GetConfig().ClusterConfigs {
			result := plugin_manager.Pm.GetClusterResult(bcro.Plugins, clusterConfig.ClusterID)
			data, _ := yaml.Marshal(result)
			fmt.Println(string(data))
		}
		return
	}

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
	if bcro.BcsGatewayToken == "" {
		bcro.BcsGatewayToken = os.Getenv("gatewayToken")
	}

	if bcro.BcsClusterManagerToken == "" {
		bcro.BcsClusterManagerToken = os.Getenv("gatewayToken")
	}

	if bcro.BcsClusterManagerToken != "" || bcro.BcsClusterManagerApiserver != "" || bcro.BcsGatewayApiserver != "" ||
		bcro.BcsGatewayToken != "" {
		if bcro.BcsClusterManagerToken == "" || bcro.BcsClusterManagerApiserver == "" || bcro.BcsGatewayApiserver == "" ||
			bcro.BcsGatewayToken == "" {
			return fmt.Errorf(
				"bcs config missing, BcsClusterManagerToken, BcsClusterManagerApiserver, BcsGatewayApiserver, BcsGatewayToken must be set")
		}
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

func getClusters() {
	clusterConfigList := make(map[string]*plugin_manager.ClusterConfig)
	if plugin_manager.Pm.GetConfig() != nil {
		clusterConfigList = plugin_manager.Pm.GetConfig().ClusterConfigs
	}

	// 从bcs获取BCS集群配置
	if bcro.BcsGatewayApiserver != "" && bcro.BcsClusterManagerApiserver != "" && bcro.BcsGatewayToken != "" &&
		bcro.BcsClusterManagerToken != "" {
		bcsClusterConfigList, err := GetClusterConfigFromBCS(bcro.BcsClusterManagerToken,
			bcro.BcsClusterManagerApiserver, bcro.BcsGatewayApiserver, bcro.BcsGatewayToken, clusterConfigList)
		if err != nil {
			klog.Fatalf(err.Error())
		}
		clusterConfigList = bcsClusterConfigList
	}

	// 从文件夹获取kubeconfig的配置
	if bcro.KubeConfigDir != "" {
		klog.Infof(bcro.KubeConfigDir)
		fileClusterConfigList, err := GetClusterInfo(bcro.KubeConfigDir, clusterConfigList)
		if err != nil {
			klog.Fatalf(err.Error())
		}
		for key, value := range fileClusterConfigList {
			clusterConfigList[key] = value
		}
	}

	// Incluster模式
	if bcro.InCluster {
		config, err := rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("Error: %s", err.Error())
			return
		}
		clusterConfigList[bcro.ClusterID] = &plugin_manager.ClusterConfig{BusinessID: bcro.BizID,
			ClusterID: bcro.ClusterID, Config: config}
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

// GetClusterConfigFromBCS get clusterconfig from bcs api
func GetClusterConfigFromBCS(bcsClusterManagerToken, bcsClusterManagerApiserver, bcsGatewayApiserver, bcsGatewayToken string, existClusterConfigList map[string]*plugin_manager.ClusterConfig) (map[string]*plugin_manager.ClusterConfig, error) {
	clusterConfigList := make(map[string]*plugin_manager.ClusterConfig)
	bcsClusterManager, err := bcs.NewClusterManager(bcsClusterManagerToken, bcsClusterManagerApiserver,
		bcsGatewayApiserver, bcsGatewayToken)
	if err != nil {
		klog.Fatalf("NewClusterManager failed: %s", err.Error())
		return nil, err
	}

	clusterList, err := bcsClusterManager.GetClusters([]string{})
	if err != nil {
		klog.Fatalf("GetClusters failed: %s", err.Error())
		return nil, err
	}

	filteredClusterList := make([]cmproto.Cluster, 0, 0)
	if len(bcro.BcsClusterList) != 0 {
		for _, clusterId := range bcro.BcsClusterList {
			for _, cluster := range clusterList {
				if clusterId == cluster.ClusterID {
					filteredClusterList = append(filteredClusterList, cluster)
					break
				}
			}
		}
	} else {
		for _, cluster := range clusterList {
			if cluster.IsShared == true {
				continue // 跳过公共集群的记录
			} else if cluster.Status != "RUNNING" {
				continue // 跳过未就绪集群
			} else if cluster.EngineType != "k8s" {
				continue // 跳过非K8S集群
			} else {
				// 跳过master ip不正常的集群
				if len(cluster.Master) > 0 {
					continueFlag := false
					for masterName, _ := range cluster.Master {
						if strings.Contains(masterName, "127.0.0") {
							// 跳过算力集群
							continueFlag = true
							break
						}
					}
					if continueFlag {
						klog.Infof("skip %s , master ip starts with 127.0.0", cluster.ClusterID)
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
						// 创建时间超过10分钟才进行巡检
						if (time.Now().Unix() - createTime.Unix()) > 60*30 {
							filteredClusterList = append(filteredClusterList, cluster)
						}
					}
				}
			}

		}
	}

	var mapLock sync.Mutex
	var wg sync.WaitGroup
	routinePool := util.NewRoutinePool(50)
	for _, cluster := range filteredClusterList {
		wg.Add(1)
		routinePool.Add(1)
		go func(cluster cmproto.Cluster) {
			defer func() {
				wg.Done()
				routinePool.Done()
			}()

			config := bcsClusterManager.GetKubeconfig(cluster.ClusterID)
			clusterConfig, err := GetClusterConfig(cluster.ClusterID, config)
			if err != nil {
				klog.Errorf("GetClusterConfig %s failed: %s", cluster.ClusterID, err.Error())
				return
			}

			if existClusterConfig, ok := existClusterConfigList[cluster.ClusterID]; ok {
				existClusterConfig.Config = config
				existClusterConfig.ClientSet = clusterConfig.ClientSet
				clusterConfig = existClusterConfig
			}

			clusterConfig.ClusterID = cluster.ClusterID
			clusterConfig.BusinessID = cluster.BusinessID
			clusterConfig.BCSCluster = cluster
			clusterConfig.NodeInfo = make(map[string]plugin.NodeInfo)

			if strings.HasPrefix(cluster.SystemID, "cls") {
				clusterConfig.ClusterType = plugin_manager.TKECluster
			}

			mapLock.Lock()
			clusterConfigList[clusterConfig.ClusterID] = clusterConfig
			mapLock.Unlock()
		}(cluster)
	}
	wg.Wait()

	return clusterConfigList, nil
}

// GetClusterInfo return ClusterConfig by parsing kubeconfig file
func GetClusterInfo(kubeConfigDir string, existClusterConfigList map[string]*plugin_manager.ClusterConfig) (map[string]*plugin_manager.ClusterConfig, error) {
	clusterConfigList := make(map[string]*plugin_manager.ClusterConfig)
	var filePathList []string
	err := filepath.Walk(kubeConfigDir, func(path string, info os.FileInfo, err error) error {
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
		return nil, err
	}

	for _, filePath := range filePathList {
		filenameWithExt := filepath.Base(filePath) // 获取文件名（包括后缀）
		ext := filepath.Ext(filenameWithExt)       // 获取文件后缀

		filename := strings.TrimSuffix(filenameWithExt, ext) // 移除后缀

		config, err := k8s.GetRestConfigByConfig(filePath)
		if err != nil {
			return nil, err
		}

		if config.CAData == nil {
			config.TLSClientConfig.Insecure = true
		}

		clusterConfig, err := GetClusterConfig(filename, config)
		if err != nil {
			klog.Errorf("GetClusterConfig %s failed: %s", filename, err.Error())
			continue
		}

		if existClusterConfig, ok := existClusterConfigList[filename]; ok {
			existClusterConfig.Config = config
			existClusterConfig.ClientSet = clusterConfig.ClientSet
			clusterConfig = existClusterConfig
		}

		clusterConfig.ClusterID = filename
		clusterConfig.BusinessID = "0"
		clusterConfig.NodeInfo = make(map[string]plugin.NodeInfo)

		//  读取配置文件时没配置bizid
		clusterConfigList[filename] = clusterConfig

		klog.Infof("load kubeconfig success, clusterID: %s", filename)
	}

	return clusterConfigList, nil
}

// GetClusterConfig return ClusterConfig by clusterID and rest config
func GetClusterConfig(clusterID string, config *rest.Config) (*plugin_manager.ClusterConfig, error) {
	clusterConfig := &plugin_manager.ClusterConfig{}

	clientSet, err := k8s.GetClientsetByConfig(config)
	if err != nil {
		return nil, fmt.Errorf("get clientset failed: %s, skip", err.Error())
	}

	// 跳过算力集群
	apiResources, err := k8s.GetK8sApi(clientSet)
	if err != nil {
		klog.Errorf("get %s apiresourcelist failed: %s", clusterID, err.Error())
	} else {
		for _, group := range apiResources {
			if group.GroupVersion == "cluster.karmada.io/v1alpha1" {
				return nil, fmt.Errorf("has karmada resource, skip")
			}
		}
	}

	// 跳过work node为0的集群
	nodeList, err := clientSet.CoreV1().Nodes().List(util.GetCtx(time.Second*10), v1.ListOptions{ResourceVersion: "0"})
	masterList := make([]string, 0, 0)
	if err != nil {
		klog.Errorf("get %s node failed: %s", clusterID, err.Error())
	} else {
		nodeNum := len(nodeList.Items)
		for _, node := range nodeList.Items {
			for key, val := range node.Labels {
				if key == "node-role.kubernetes.io/master" && (val == "true" || val == "") {
					nodeNum = nodeNum - 1
					for _, address := range node.Status.Addresses {
						if address.Type == v12.NodeInternalIP {
							masterList = append(masterList, address.Address)
						}
					}

				}
			}
		}

		if nodeNum == 0 && strings.Contains(clusterID, "BCS-K8S-4") {
			return nil, fmt.Errorf("a cluster without any work nodes, skip")
		}
	}

	clusterConfig = &plugin_manager.ClusterConfig{
		Config:    config,
		ClientSet: clientSet,
		Master:    masterList,
	}

	clusterConfig.ClusterID = "incluster"
	clusterConfig.BusinessID = "0"
	clusterConfig.NodeInfo = make(map[string]plugin.NodeInfo)
	return clusterConfig, nil
}

func main() {
	Execute()
	defer klog.Flush()
}
