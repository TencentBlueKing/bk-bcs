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

// Package app xxx
package app

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	global "github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/k8s/resources"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/output"
)

var globalStopChan = make(chan struct{})

func signalListen() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM) // nolint
	for {
		sig := <-ch
		fmt.Printf(fmt.Sprintf("get sig=%v\n", sig)) // nolint

		switch sig {
		case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM:
			fmt.Println("get exit signal, exit process")
			// stop signal
			close(globalStopChan)
			// exit process
			os.Exit(0)
		}
	}
}

func validateConfigFilePath(configFilePath string) error {

	file, err := os.Stat(configFilePath)
	if nil != err {
		return fmt.Errorf("check config file %s failed. error: %v", configFilePath, err)
	}
	if file.IsDir() {
		return fmt.Errorf("config file path %s is a directory", configFilePath)
	}
	return nil

}

func savePID(pidFilePath string) error {

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("check current path failed. Error:%v", err)
	}

	var pidFileSavePath string
	pidFileName := fmt.Sprintf("%s.pid", filepath.Base(os.Args[0]))
	if pidFilePath == "" {
		pidFileSavePath = currentDir + "/" + pidFileName
	} else {
		pidFileSavePath = strings.TrimRight(pidFilePath, "/") + "/" + pidFileName
	}
	global.SetPidfilePath(pidFileSavePath)

	if err := global.WritePid(); nil != err {
		return fmt.Errorf("write pid file failed. Error: %v", err)
	}
	return nil
}

// PrepareRun checks configuration for running
func PrepareRun(configFilePath string, pidFilePath string) error {
	err := savePID(pidFilePath)
	if err != nil {
		return err
	}

	err = validateConfigFilePath(configFilePath)
	return err
}

// Run entrypoint for k8s-watch
func Run(watchConfig *options.WatchConfig) error {
	if len(watchConfig.BCS.NetServiceZKHosts) == 0 {
		watchConfig.BCS.NetServiceZKHosts = watchConfig.BCS.ZkHosts
	}

	if err := global.SavePid(watchConfig.ProcessConfig); err != nil {
		glog.Warn("fail to save pid. err:%s", err.Error())
	} else {
		glog.Infof("save pid successful")
	}

	glog.Info("Get ClusterID DONE! ClusterID=%s", watchConfig.Default.ClusterID)

	glog.Info("I'm leader.")
	stopChan := make(chan struct{})
	go RunAsLeader(stopChan, watchConfig, watchConfig.Default.ClusterID) // nolint
	go signalListen()

	// Stop RunAsLeader by forward the stop event
	select { // nolint
	// Closed by global signal handler
	case <-globalStopChan:
		close(stopChan)
	}
	return nil
}

// RunAsLeader do the leader stuff
func RunAsLeader(stopChan <-chan struct{}, config *options.WatchConfig, clusterID string) error { // nolint
	bcsTLSConfig := config.BCS.TLS

	glog.Info("getting storage service now...")
	storageService, _, err := bcs.GetStorageService(config.BCS.ZkHosts, bcsTLSConfig, config.BCS.CustomStorageEndpoints,
		config.BCS.IsExternal)
	if err != nil {
		panic(err)
	}
	glog.Info("get storage service done")

	var (
		netservice     *bcs.InnerService
		netserviceZKRD *RegisterDiscover.RegDiscover
	)
	if !config.WatchResource.DisableNetservice {
		glog.Info("getting netservice now...")
		netservice, netserviceZKRD, err = bcs.GetNetService(config.BCS.NetServiceZKHosts, bcsTLSConfig,
			config.BCS.CustomNetServiceEndpoints, false)
		if err != nil {
			panic(err)
		}
		glog.Info("get netservice done")

		// waiting for netservice discovery
		time.Sleep(5 * time.Second)
		if len(netservice.Servers()) == 0 {
			glog.Infof("got non netservice address this moment")
		}
	}

	// init server actions && register web server && register metrics server
	glog.Info("start http server")
	certConfig := bcs.CertConfig{
		CAFile:   config.CAFile,
		CertFile: config.ServerCertFile,
		KeyFile:  config.ServerKeyFile,
		CertPwd:  static.ServerCertPwd,
	}
	httpServer := bcs.GetHTTPServer(config, bcs.WithCertConfig(certConfig), bcs.WithDebug(config.DebugMode))
	go func() {
		err = httpServer.ListenAndServe()
		if err != nil {
			glog.Errorf("http listen and serve failed: %v", err)
			close(globalStopChan)
		}
	}()

	bcs.RunPrometheusMetricsServer(config)
	glog.Info("start http server successful")

	filterConfig := config.ParseFilter()
	// init resourceList to watch
	err = resources.InitResourceList(&config.K8s, filterConfig, &config.WatchResource)
	if err != nil {
		panic(err)
	}

	// create writer.
	glog.Info("creating writer now...")
	writer, err := output.NewWriter(clusterID, storageService, config.BCS)
	if err != nil {
		panic(err)
	}
	glog.Info("create writer success")

	glog.Info("starting writer now...")
	// NOCC:vetshadow/shadow(设计如此:这里err可以被覆盖)
	if err = writer.Run(stopChan); err != nil {
		panic(err)
	}
	glog.Info("start writer success")

	// create watcher manager.
	glog.Info("creating watcher manager now...")
	watcherMgr, err := k8s.NewWatcherManager(
		&k8s.WatcherManagerOptions{
			ClusterID:      clusterID,
			WatchResource:  &config.WatchResource,
			FilterConfig:   filterConfig,
			Writer:         writer,
			K8sConfig:      &config.K8s,
			StorageService: storageService,
			Netservice:     netservice,
			StopChan:       stopChan,
		})
	if err != nil {
		panic(err)
	}
	glog.Info("create watcher manager success")

	glog.Info("start watcher manager now...")
	watcherMgr.Run(stopChan)
	glog.Info("start watcher manager success")

	<-stopChan

	// stop all kubefed watchers
	watcherMgr.StopCrdWatchers()

	if !config.WatchResource.DisableNetservice {
		// stop service discovery.
		err = netserviceZKRD.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}
