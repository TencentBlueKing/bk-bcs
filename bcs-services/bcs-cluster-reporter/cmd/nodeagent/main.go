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
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/cmd/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

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
	config, err := rest.InClusterConfig()
	if err != nil {
		if cmdOptions.KubeConfigPath != "" {
			config, err = k8s.GetRestConfigByConfig(cmdOptions.KubeConfigPath)
			if err != nil {
				klog.Fatalf("Error: %s", err.Error())
			}
		}
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
	plugin_manager.Pm.SetConfig(&plugin_manager.Config{
		NodeConfig: plugin_manager.NodeConfig{
			Config:    config,
			ClientSet: clientSet,
			NodeName:  nodeName,
			Node:      node,
			HostPath:  hostPath,
		},
	})

	// 读取配置文件
	go func() {
		err := plugin_manager.Pm.SetupPlugin(cmdOptions.Plugins, cmdOptions.ConfigPath, cmdOptions.RunMode)
		if err != nil {
			klog.Fatalf(err.Error())
		}
	}()

	// listening OS shutdown singal
	if cmdOptions.RunMode == plugin_manager.RunModeDaemon {
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
		plugin_manager.Pm.Ready(cmdOptions.Plugins, "node")
		result := plugin_manager.Pm.GetNodeResult(cmdOptions.Plugins)
		// TODO 支持json格式输出
		checkItemList := make([]plugin_manager.CheckItem, 0, 0)
		for _, item := range result {
			checkItemList = append(checkItemList, item.Items...)
		}
		infoItemList := make([]plugin_manager.InfoItem, 0, 0)
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
