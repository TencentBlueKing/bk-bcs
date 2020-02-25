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
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"bk-bcs/bcs-common/common/types"
	bcsVersion "bk-bcs/bcs-common/common/version"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/bcs"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/k8s"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/options"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output/action"

	global "bk-bcs/bcs-common/common"
	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/metric"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/k8s/resources"
	disbcs "bk-bcs/bcs-k8s/bcs-k8s-watch/pkg/discovery/bcs"
	disreg "bk-bcs/bcs-k8s/bcs-k8s-watch/pkg/discovery/register"
)

var globalStopChan = make(chan struct{})

func signalListen() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)
	for {
		sig := <-ch
		fmt.Printf(fmt.Sprintf("get sig=%v\n", sig))

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

func PrepareRun(configFilePath string, pidFilePath string) error {
	err := savePID(pidFilePath)
	if err != nil {
		return err
	}

	err = validateConfigFilePath(configFilePath)
	return err
}

func Run(configFilePath string) error {
	// 1. init configTODO
	glog.Info("Init config begin......")
	watchConfig, err := options.ParseConfigFile(configFilePath)
	if err != nil {
		panic(err.Error())
	}

	glog.Info("Init config DONE!")

	zkHosts := strings.Join(watchConfig.BCS.ZkHosts, ",")
	if watchConfig.BCS.NetServiceZKHosts == nil || len(watchConfig.BCS.NetServiceZKHosts) == 0 {
		watchConfig.BCS.NetServiceZKHosts = watchConfig.BCS.ZkHosts
	}

	hostIP := watchConfig.Default.HostIP
	bcsTLSConfig := watchConfig.BCS.TLS

	var clusterID string
	// 1.1 get clusterID
	// Get ClusterID via ClusterKeeper
	glog.Info("Get ClusterID begin......")
	if watchConfig.Default.ClusterIDSource == options.ClusterIDSourceClusterKeeper {
		clusterID, err = bcs.GetClusterID(zkHosts, hostIP, bcsTLSConfig)
		if err != nil {
			panic(err.Error())
		}
	} else {
		// Read ClusterID from config file
		clusterID = watchConfig.Default.ClusterID
	}
	glog.Info("Get ClusterID DONE! ClusterID=%s", clusterID)

	// Get current node info to register it to zookeeper
	hostname, _ := os.Hostname()
	serverInfo := types.ServerInfo{
		IP:       hostIP,
		Port:     0,
		HostName: hostname,
		Scheme:   "",
		Cluster:  clusterID,
		Version:  bcsVersion.GetVersion(),
		Pid:      os.Getpid(),
	}
	node := disbcs.NewServiceNode(serverInfo)

	// Register current node info to zookeeper and start discovering
	basePath := fmt.Sprintf("%s/%s/%s",
		types.BCS_SERV_BASEPATH,
		types.BCS_MODULE_KUBEDATAWATCH,
		clusterID,
	)
	reg := disreg.NewNodeRegister(zkHosts, basePath, &node)
	if err := reg.DoRegister(); err != nil {
		return fmt.Errorf("unable to register data-watch: %s", err)
	}
	go reg.StartDiscover(0)

	// Start leadereletor
	callbacks := &disbcs.LeaderCallbacks{
		OnStartedLeading: func(stop <-chan struct{}) {
			glog.Info("I'm leader.")
			stopChan := make(chan struct{})
			go RunAsLeader(stopChan, watchConfig, clusterID)

			// Stop RunAsLeader by forward the stop event
			select {
			case <-stop:
				// Closed by leaderElector
				close(stopChan)
				// Closed by global signal handler
			case <-globalStopChan:
				close(stopChan)
			}
		},
		OnStoppedLeading: func() {
			glog.Info("I'm slave now.")
		},
	}
	elector := disbcs.NewLeaderElector(reg, callbacks)

	go signalListen()
	go elector.Run()

	<-globalStopChan
	return nil
}

func startMetricForMaster(moduleIP, clusterID string) error {
	// NOTE: will use the IP:MetricPort as the listen addr
	c := metric.Config{
		ModuleName:          "k8s-watch",
		IP:                  moduleIP,
		MetricPort:          9089,
		DisableGolangMetric: true,
		ClusterID:           clusterID,
	}
	healthz := func() metric.HealthMeta {
		return metric.HealthMeta{
			CurrentRole: "Master",
			IsHealthy:   true,
		}
	}
	if err := metric.NewMetricController(c, healthz); err != nil {
		fmt.Printf("new metric collector failed. err: %v\n", err)
		return err
	}
	return nil
}

// RunAsLeader do the leader stuff
func RunAsLeader(stopChan <-chan struct{}, config *options.WatchConfig, clusterID string) error {
	zkHosts := strings.Join(config.BCS.ZkHosts, ",")
	netServiceZKHosts := strings.Join(config.BCS.NetServiceZKHosts, ",")
	bcsTLSConfig := config.BCS.TLS

	glog.Info("getting storage service now...")
	storageService, storageServiceZKRD, err := bcs.GetStorageService(zkHosts, bcsTLSConfig, config.BCS.CustomStorageEndpoints, config.BCS.IsExternal)
	if err != nil {
		panic(err)
	}
	glog.Info("get storage service done")

	glog.Info("getting netservice now...")
	netservice, netserviceZKRD, err := bcs.GetNetService(netServiceZKHosts, bcsTLSConfig, config.BCS.CustomNetServiceEndpoints, false)
	if err != nil {
		panic(err)
	}
	glog.Info("get netservice done")

	// sleep for 5 seconds, wait.
	glog.Infof("sleep for 5 seconds, wait for fetching storage/netservice address from ZK")
	time.Sleep(5 * time.Second)

	if len(storageService.Servers()) == 0 && config.Default.Environment != "development" {
		glog.Infof("got non storage service address, sleep for another again")
		time.Sleep(5 * time.Second)

		if len(storageService.Servers()) == 0 {
			panic("can't get storage service address from ZK after 10 seconds")
		}
	}
	if len(netservice.Servers()) == 0 {
		glog.Infof("got non netservice address this moment")
	}

	// init alertor with bcs-health.
	moduleIP := config.Default.HostIP
	alertor, err := action.NewAlertor(clusterID, moduleIP, zkHosts, config.BCS.TLS)
	if err != nil {
		glog.Warnf("Init Alertor fail, no alarm will be sent!")
	}

	// init resourceList to watch
	err = resources.InitResourceList(&config.K8s)
	if err != nil {
		panic(err)
	}

	// create writer.
	glog.Info("creating writer now...")
	writer, err := output.NewWriter(clusterID, storageService, alertor)
	if err != nil {
		panic(err)
	}
	glog.Info("create writer success")

	glog.Info("starting writer now...")
	if err := writer.Run(stopChan); err != nil {
		panic(err)
	}
	glog.Info("start writer success")

	// create watcher manager.
	glog.Info("creating watcher manager now...")
	watcherMgr, err := k8s.NewWatcherManager(clusterID, writer, &config.K8s, storageService, netservice, stopChan)
	if err != nil {
		panic(err)
	}
	glog.Info("create watcher manager success")

	glog.Info("start watcher manager now...")
	watcherMgr.Run(stopChan)
	glog.Info("start watcher manager success")

	// finally, start metric, allow fail
	glog.Info("start metric......")
	err = startMetricForMaster(moduleIP, clusterID)
	if err != nil {
		glog.Errorf("Init metric fail, the metric and health will not be ok!")
	}

	// glog.Infof("start health checker......")
	// h := HealthChecker{}
	// go h.Run(stopChan)

	<-stopChan

	// stop all kubefed watchers
	watcherMgr.StopCrdWatchers()

	// stop service discovery.
	storageServiceZKRD.Stop()
	netserviceZKRD.Stop()

	return nil
}
