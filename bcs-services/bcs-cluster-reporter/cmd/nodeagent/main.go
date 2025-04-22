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
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/cmd/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/configfilecheck"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/containercheck"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/diskcheck"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/dnscheck"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/hwcheck"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/netcheck"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/nodeinfocheck"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/processcheck"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/timecheck"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/uploader"
)

var (
	cmdOptions = options.NewNodeAgentOptions()

	rootCmd = &cobra.Command{
		Use:   "bcs-nodeagent",
		Short: "bcs-nodeagent",
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

func initConfig() {}

func init() {
	flags := rootCmd.PersistentFlags()
	cmdOptions.AddFlags(flags)

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

// Run main process
func Run() error {
	config := getConfig(cmdOptions.KubeConfigPath)
	if config == nil {
		klog.Fatalf("get kubeconfig failed.")
	}

	clientSet, err := k8s.GetClientsetByConfig(config)
	if err != nil {
		klog.Fatalf("Error: %s", err.Error())
	}

	nodeName := util.GetNodeName()
	node, err := clientSet.CoreV1().Nodes().Get(util.GetCtx(10*time.Second), nodeName, v1.GetOptions{ResourceVersion: "0"})
	if err != nil {

		klog.Fatalf("Error: %s", err.Error())
	}

	hostPath := cmdOptions.HostPath
	if hostPath == "/" {
		hostPath = util.GetHostPath()
	}

	kubernetesSvc, err := clientSet.CoreV1().Services("default").Get(util.GetCtx(10*time.Second), "kubernetes", v1.GetOptions{
		ResourceVersion: "0",
	})
	if err != nil {
		klog.Fatalf("get kubernetes svc failed: %s", err.Error())
	}

	pluginmanager.Pm.SetConfig(&pluginmanager.Config{
		NodeConfig: pluginmanager.NodeConfig{
			Config:        config,
			ClientSet:     clientSet,
			NodeName:      nodeName,
			Node:          node,
			HostPath:      hostPath,
			KubernetesSvc: kubernetesSvc.Spec.ClusterIP,
		},
	})

	// 读取配置文件
	go func() {
		err := pluginmanager.Pm.SetupPlugin(cmdOptions.Plugins, cmdOptions.ConfigPath, cmdOptions.RunMode)
		if err != nil {
			klog.Fatalf(err.Error())
		}
	}()

	// listening OS shutdown singal
	if cmdOptions.RunMode == pluginmanager.RunModeDaemon {
		r := gin.Default()
		pprof.Register(r)

		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
		go func() {
			if err := r.Run(cmdOptions.Addr); err != nil {
				klog.Fatalf(err.Error())
			}
		}()

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		<-signalChan
	} else {
		pluginmanager.Pm.Ready(cmdOptions.Plugins, "node")
		result := pluginmanager.Pm.GetNodeResult(cmdOptions.Plugins)
		// TODO 支持json格式输出
		checkItemList := make([]pluginmanager.CheckItem, 0, 0)
		for _, item := range result {
			checkItemList = append(checkItemList, item.Items...)
		}
		infoItemList := make([]pluginmanager.InfoItem, 0, 0)
		for _, item := range result {
			infoItemList = append(infoItemList, item.InfoItemList...)
		}

		data, _ := json.Marshal(checkItemList)
		fmt.Println(string(data))
		data, _ = json.Marshal(infoItemList)
		fmt.Println(string(data))
	}

	return nil
}

func getConfig(kubeconfigPath string) *rest.Config {
	configList := make([]*rest.Config, 0, 0)
	config, err := rest.InClusterConfig()
	if err == nil {
		configList = append(configList, config)
	}

	kubeconfigPathList := []string{kubeconfigPath, "/root/.kube/config", "/etc/kubernetes/kubelet-kubeconfig"}
	for _, configPath := range kubeconfigPathList {
		config, err = k8s.GetRestConfigByConfig(configPath)
		if err == nil {
			configList = append(configList, config)
		}
	}

	for _, config = range configList {
		clientSet, err := k8s.GetClientsetByConfig(config)
		if err != nil {
			klog.Errorf("Error: %s", err.Error())
		}

		nodeName := util.GetNodeName()
		_, err = clientSet.CoreV1().Nodes().Get(util.GetCtx(10*time.Second), nodeName, v1.GetOptions{ResourceVersion: "0"})
		if err != nil {
			klog.Errorf("Error: %s", err.Error())
			continue
		} else {
			return config
		}
	}

	return nil
}

// Execute rootCmd
func Execute() {
	defer func() {
		if r := recover(); r != nil {
			klog.Fatalf("nodeagent failed: %s, stack: %v\n", r, string(debug.Stack()))
		}
	}()

	err := rootCmd.Execute()
	if err != nil {
		klog.Fatalf(err.Error())
	}
}

// CheckErr deal with Complete error
func CheckErr(err error) {
	if err != nil {
		klog.Fatalf(err.Error())
	}
}

// Complete xxx
func Complete(cmd *cobra.Command, args []string) error {
	// 如果配置文件不存在则写入默认值
	_, err := os.Stat(cmdOptions.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(cmdOptions.ConfigPath), os.ModePerm)
			if err != nil {
				return err
			}

			err = util.WriteConfigIfNotExist(filepath.Join(cmdOptions.ConfigPath, "config"), `interval: 86400
pluginDir: /data/bcs/nodeagent`)
			if err != nil {
				return err
			}

		} else {
			return err
		}
	}

	return nil
}

func main() {
	Execute()
	defer klog.Flush()
}
