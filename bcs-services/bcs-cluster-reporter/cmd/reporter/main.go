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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	apiv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"

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
			if bcro.BcsGatewayToken == "" {
				bcro.BcsGatewayToken = os.Getenv("gatewayToken")
			}

			if bcro.BcsClusterManagerToken == "" {
				bcro.BcsClusterManagerToken = os.Getenv("gatewayToken")
			}

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
				time.Sleep(time.Minute * 10)
				getClusters()
			}
		}
	}()

	// start plugins
	err := pluginmanager.Pm.SetupPlugin(bcro.Plugins, bcro.PluginConfDir, bcro.RunMode)
	if err != nil {
		klog.Fatalf("Setup plugin failed: %s", err.Error())
	}

	klog.Info("Setup plugins success")

	// start webserver
	if bcro.RunMode == pluginmanager.RunModeDaemon {
		r.GET("cluster/:clusterID/pdf", func(c *gin.Context) {
			clusterID := c.Param("clusterID")
			pdf, reportErr := pluginmanager.GetClusterReport(clusterID, bcro.Plugins)
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

			pdf, reportErr := pluginmanager.GetBizReport(bizID, bcro.Plugins)
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

			html, htmlErr := pluginmanager.GetClusterReportHtml(clusterID, bcro.Plugins)
			if htmlErr != nil {
				c.String(404, fmt.Sprintf("cluster %s not found", clusterID))
				return
			}

			c.String(200, html)
			return
		})

		r.GET("biz/:bizID/html", func(c *gin.Context) {
			bizID := c.Param("bizID")

			html, htmlErr := pluginmanager.GetBizReportHtml(bizID, bcro.Plugins)
			if htmlErr != nil {
				c.String(404, fmt.Sprintf("biz %s not found", bizID))
				return
			}

			c.String(200, html)
			return
		})

		r.GET("/metrics", gin.WrapH(promhttp.Handler()))

		// config mm
		metricmanager.MM.SetEngine(r)
	} else if bcro.RunMode == pluginmanager.RunModeOnce {
		for _, clusterConfig := range pluginmanager.Pm.GetConfig().ClusterConfigs {
			result := pluginmanager.Pm.GetClusterResult(bcro.Plugins, clusterConfig.ClusterID)
			data, _ := yaml.Marshal(result)
			fmt.Println(string(data))
		}
		return
	}

	<-ctx.Done()
	// 停止模块的运行
	klog.Infof("start to stop plugins")
	err = pluginmanager.Pm.StopPlugin(bcro.Plugins)
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

// Complete check for cmd args
func Complete(cmd *cobra.Command, args []string) error {
	if (bcro.BcsClusterManagerToken != "" || bcro.BcsClusterManagerApiserver != "" || bcro.BcsGatewayApiserver != "" ||
		bcro.BcsGatewayToken != "") && (bcro.BcsClusterManagerToken == "" || bcro.BcsClusterManagerApiserver == "" || bcro.BcsGatewayApiserver == "" ||
		bcro.BcsGatewayToken == "") {
		return fmt.Errorf(
			"bcs config missing, BcsClusterManagerToken, BcsClusterManagerApiserver, BcsGatewayApiserver, BcsGatewayToken must be set")

	}

	if (bcro.BcsGatewayApiserver != "" || bcro.BcsClusterManagerApiserver != "" || bcro.BcsGatewayToken != "" ||
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

// initConfig configure viper to read config
func initConfig() {}

func getClusters() {
	clusterConfigList := make(map[string]*pluginmanager.ClusterConfig)
	if pluginmanager.Pm.GetConfig() != nil {
		clusterConfigList = pluginmanager.Pm.GetConfig().ClusterConfigs
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
		clusterConfigList[bcro.ClusterID] = &pluginmanager.ClusterConfig{
			BusinessID: bcro.BizID,
			ClusterID:  bcro.ClusterID,
			Config:     config}
		pluginmanager.Pm.SetConfig(&pluginmanager.Config{
			ClusterConfigs: clusterConfigList,
			InClusterConfig: pluginmanager.ClusterConfig{
				BusinessID: bcro.BizID,
				ClusterID:  bcro.ClusterID,
				Config:     config},
		})
	} else {
		// 集中化模式
		pluginmanager.Pm.SetConfig(&pluginmanager.Config{
			ClusterConfigs: clusterConfigList,
		})
	}
}

// GetClusterConfigFromBCS get clusterconfig from bcs api
func GetClusterConfigFromBCS(bcsClusterManagerToken, bcsClusterManagerApiserver, bcsGatewayApiserver, bcsGatewayToken string, existClusterConfigList map[string]*pluginmanager.ClusterConfig) (map[string]*pluginmanager.ClusterConfig, error) {
	clusterConfigList := make(map[string]*pluginmanager.ClusterConfig)
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
	clusterloop:
		for _, cluster := range clusterList {
			if cluster.IsShared == true || cluster.Status != "RUNNING" || cluster.EngineType != "k8s" || (cluster.Environment == bcro.BcsClusterType && bcro.BcsClusterType != "") {
				continue // 跳过公共集群的记录 跳过未就绪集群 跳过非K8S集群 以及匹配对应参数的集群
			}

			// 跳过master ip不正常的集群
			for masterName, _ := range cluster.Master {
				if strings.Contains(masterName, "127.0.0") {
					// 跳过算力集群
					continue clusterloop
				}
			}

			if strings.Contains(cluster.ClusterName, "联邦") {
				continue
			}

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

			// 已存在的集群信息则直接复用
			if existClusterConfig, ok := existClusterConfigList[cluster.ClusterID]; ok {
				existClusterConfig.Config = config
				existClusterConfig.ClientSet = clusterConfig.ClientSet
				existClusterConfig.MetricSet = clusterConfig.MetricSet
				clusterConfig = existClusterConfig
			}

			clusterConfig.ClusterID = cluster.ClusterID
			clusterConfig.BusinessID = cluster.BusinessID
			clusterConfig.BCSCluster = cluster

			if strings.HasPrefix(cluster.SystemID, "cls") {
				clusterConfig.ClusterType = pluginmanager.TKECluster
			} else {
				clusterConfig.ClusterType = cluster.ClusterType
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
func GetClusterInfo(kubeConfigDir string, existClusterConfigList map[string]*pluginmanager.ClusterConfig) (map[string]*pluginmanager.ClusterConfig, error) {
	clusterConfigList := make(map[string]*pluginmanager.ClusterConfig)
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
			existClusterConfig.MetricSet = clusterConfig.MetricSet
			clusterConfig = existClusterConfig
		}

		clusterConfig.ClusterID = filename
		clusterConfig.BusinessID = "0"

		//  读取配置文件时没配置bizid
		clusterConfigList[filename] = clusterConfig

		klog.Infof("load kubeconfig success, clusterID: %s", filename)
	}

	return clusterConfigList, nil
}

func checkMetricAPI(cluster *pluginmanager.ClusterConfig) error {
	dynamicConfig, err := dynamic.NewForConfig(cluster.Config)
	if err != nil {
		return fmt.Errorf("%s get metric failed: %s", cluster.ClusterID, err.Error())
	}

	result, err := dynamicConfig.Resource(schema.GroupVersionResource{
		Group:    "apiregistration.k8s.io",
		Version:  "v1",
		Resource: "apiservices",
	}).Get(util.GetCtx(10*time.Second), "v1beta1.metrics.k8s.io", v1.GetOptions{ResourceVersion: "0"})

	if err != nil {
		return fmt.Errorf("%s get metric failed: %s", cluster.ClusterID, err.Error())
	}

	apiService := &apiv1.APIService{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(result.Object, apiService); err != nil {
		return fmt.Errorf("%s get metric failed: %s", cluster.ClusterID, err.Error())
	}

	for _, condition := range apiService.Status.Conditions {
		if condition.Status == apiv1.ConditionFalse {
			return fmt.Errorf("%s get metric failed: %s %s", cluster.ClusterID, condition.Reason, condition.Message)
		}
	}
	return nil
}

// GetClusterConfig return ClusterConfig by clusterID and rest config
func GetClusterConfig(clusterID string, config *rest.Config) (*pluginmanager.ClusterConfig, error) {
	clusterConfig := &pluginmanager.ClusterConfig{}

	clientSet, err := k8s.GetClientsetByConfig(config)
	if err != nil {
		return nil, fmt.Errorf("get clientset failed: %s, skip", err.Error())
	}
	metricsClient, err := metricsclientset.NewForConfig(config)
	if err != nil {
		klog.Errorf("%s Get metric set failed: %s", clusterID, err.Error())
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

	nodeNum := 0
	masterList := make([]string, 0, 0)
	nodeList, err := clientSet.CoreV1().Nodes().List(util.GetCtx(time.Second*10), v1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		// 获取节点失败可能由于集群已经出问题了，所以需要继续将此集群加入集群列表，以进行检查
		klog.Errorf("get %s node failed: %s", clusterID, err.Error())
	}

	for _, node := range nodeList.Items {
		for key, _ := range node.Labels {
			if key == "node-role.kubernetes.io/master" {
				masterList = append(masterList, getIP(&node))
			}
		}

		if checkNodeReady(node) {
			nodeNum = nodeNum + 1
		}
	}

	nodeNum = nodeNum - len(masterList)
	// 排除没有任何正常工作节点的集群
	if nodeNum <= 0 && strings.Contains(clusterID, "BCS-K8S-4") {
		return nil, fmt.Errorf("a cluster without any work nodes, skip")
	}

	nodeInfo := make(map[string]plugin.NodeInfo)

	clusterConfig = &pluginmanager.ClusterConfig{
		ClusterID: clusterID,
		Config:    config,
		ClientSet: clientSet,
		MetricSet: metricsClient,
		Master:    masterList,
		NodeInfo:  nodeInfo,
	}

	// 检测集群得metricAPI是否可用，不可用置为nil
	err = checkMetricAPI(clusterConfig)
	if err != nil {
		clusterConfig.MetricSet = nil
		klog.Errorf(err.Error())
	}

	return clusterConfig, nil
}

func checkNodeReady(node v12.Node) bool {
	//if node.Status.Phase == v12.NodeRunning {
	for _, condition := range node.Status.Conditions {
		if condition.Type == v12.NodeReady && condition.Status == v12.ConditionTrue {
			return true
		}
	}
	//}
	return false
}

func getIP(node *v12.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type != v12.NodeInternalIP {
			continue
		}
		return address.Address
	}
	return ""
}

func main() {
	Execute()
	defer klog.Flush()
}
