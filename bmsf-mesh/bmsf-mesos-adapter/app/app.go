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
	pidfile "bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/pkg/queue"
	"bk-bcs/bmsf-mesh/bmsf-mesos-adapter/controller"
	"bk-bcs/bmsf-mesh/bmsf-mesos-adapter/discovery"
	"bk-bcs/bmsf-mesh/bmsf-mesos-adapter/discovery/bcs"
	"bk-bcs/bmsf-mesh/bmsf-mesos-adapter/rdiscover"
	"bk-bcs/bmsf-mesh/pkg/apis"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

//Run entry for bmsf-mesos-adaptor
func Run(config *Config) error {
	//pid configuration
	if pidErr := pidfile.SavePid(config.ProcessConfig); pidErr != nil {
		blog.Errorf("bmsf-mesos-adaptor save pid file failed, %s", pidErr)
	}

	// register to bcs service layer, just for health check
	// no need to process discover event
	config.BCSZk = strings.ReplaceAll(config.BCSZk, ";", ",")
	bcsDiscover, bcsDiscoverEvent := rdiscover.NewAdapterDiscover(
		config.BCSZk, config.Address, config.Cluster, config.MetricPort)
	go bcsDiscover.Start()
	go func() {
		for {
			select {
			case curEvent := <-bcsDiscoverEvent:
				blog.Infof("found bcs service discover event %s", curEvent)
			}
		}
	}()

	// create AdapterDiscover
	config.Zookeeper = strings.ReplaceAll(config.Zookeeper, ";", ",")
	adapterDiscover, discoverEvent := rdiscover.NewAdapterDiscover(
		config.Zookeeper, config.Address, config.Cluster, config.MetricPort)
	go adapterDiscover.Start()
	handleEvent(config, discoverEvent)
	return nil
}

func handleEvent(config *Config, event <-chan rdiscover.RoleEvent) {
	var s *Server
	signalChan := make(chan os.Signal, 5)
	signal.Notify(signalChan, syscall.SIGTRAP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	for {
		select {
		case curEvent := <-event:
			if curEvent == rdiscover.MasterToSlave {
				if s != nil {
					s.Stop()
				}
			} else if curEvent == rdiscover.SlaveToMaster {
				s := &Server{
					config:    config,
					mgrStop:   make(chan struct{}),
					svcQueue:  queue.NewQueue(),
					nodeQueue: queue.NewQueue(),
				}
				//todo: init event messge Queue for Reconciler & bk-bcs cluster
				s.mgr = settingManager(config.KubeConfig, int(config.MetricPort), s.svcQueue, s.nodeQueue)
				//create cluster plugin
				clusterZkHosts := strings.Split(config.Zookeeper, ",")
				bkBcsCluster, err := bcs.NewCluster(config.Cluster, clusterZkHosts)
				if err != nil {
					blog.Errorf("init bk-bcs cluster failed, %s", err)
					fmt.Printf("init bk-bcs cluster failed, %s", err.Error())
					os.Exit(-1)
				}
				s.cluster = bkBcsCluster
				go func() {
					err = s.Run()
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(-1)
					}
				}()

			} else {
				blog.Errorf("invalid event %s", curEvent)
			}
		case sig := <-signalChan:
			blog.Warnf("bmsf-mesos-adaptor was killed, signal info: %s", sig.String())
			if s != nil {
				s.Stop()
			}
			return
		}
	}
}

func settingManager(kubeconfig string, port int, svc queue.Queue, node queue.Queue) manager.Manager {
	//init manager
	blog.Infof("setting up manager...")
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		blog.Errorf("create BcsManager with kubeconfig %s failed, %s", kubeconfig, err.Error())
		os.Exit(1)
	}
	metricsAddr := ":" + strconv.Itoa(port)
	mgr, err := manager.New(restConfig, manager.Options{MetricsBindAddress: metricsAddr})
	if err != nil {
		blog.Errorf("unable to set up overall controller manager, %s", err)
		os.Exit(1)
	}
	// Setup Scheme for all resources
	blog.Infof("setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		blog.Errorf("unable add APIs to scheme, %s", err)
		os.Exit(1)
	}
	// Setup all Controllers
	blog.Infof("Setting up controller")
	if err := controller.AddToManager(mgr, svc, node); err != nil {
		blog.Errorf("unable to add speficied AppSvc & AppNode Reconciler to Manager, %s", err.Error())
		os.Exit(1)
	}
	return mgr
}

//Server discovery server, holding all resources
type Server struct {
	config    *Config           //server config
	mgrStop   chan struct{}     //stop signal chan for manager
	mgr       manager.Manager   //manager of all Reconciler
	cluster   discovery.Cluster //cluster instance
	svcQueue  queue.Queue       //queue for AppSvc
	nodeQueue queue.Queue       //queue for AppNode
}

//Run running server loop
func (s *Server) Run() error {
	//starting manager
	go func() {
		if err := s.mgr.Start(s.mgrStop); err != nil {
			fmt.Printf("mesos-adaptor starting kubemanager failed, %s", err.Error())
			os.Exit(1)
		}
	}()
	time.Sleep(time.Second * 2)
	//wait for Cache ready
	blog.Infof("mesos-adaptor is waiting for kubemanager sync all cache datas.")
	caches := s.mgr.GetCache()
	if ok := caches.WaitForCacheSync(s.mgrStop); !ok {
		blog.Errorf("mesos-adaptor is waiting for cache synchronization failed, data synchronizing broken.")
		return fmt.Errorf("data synchronization broken")
	}
	//starting cluster
	s.cluster.AppSvcs().RegisterAppSvcQueue(s.svcQueue)
	s.cluster.AppNodes().RegisterAppNodeQueue(s.nodeQueue)
	s.cluster.Run()
	<-s.mgrStop
	return nil
}

//Stop stop server
func (s *Server) Stop() {
	s.cluster.Stop()
	blog.Infof("Server is waiting 3 seconds for writing data back to kube-apiserver")
	time.Sleep(time.Second * 3)
	close(s.mgrStop)
}
