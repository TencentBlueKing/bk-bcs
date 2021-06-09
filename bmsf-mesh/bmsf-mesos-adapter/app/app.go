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
	"strconv"
	"strings"
	"syscall"
	"time"

	pidfile "github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/queue"
	"github.com/Tencent/bk-bcs/bmsf-mesh/bmsf-mesos-adapter/controller"
	"github.com/Tencent/bk-bcs/bmsf-mesh/bmsf-mesos-adapter/discovery"
	"github.com/Tencent/bk-bcs/bmsf-mesh/bmsf-mesos-adapter/discovery/bcs"
	"github.com/Tencent/bk-bcs/bmsf-mesh/bmsf-mesos-adapter/rdiscover"
	apis "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/mesh/v1"

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
	if len(config.BCSZk) != 0 {
		config.BCSZk = strings.Replace(config.BCSZk, ";", ",", -1)
		bcsDiscover, bcsDiscoverEvent, err := rdiscover.NewAdapterDiscover(
			config.BCSZk, config.Address, config.Cluster, config.MetricPort)
		if err != nil {
			blog.Warnf("new bcs zookeeper %s Discover failed, err %s", config.BCSZk, err.Error())
		} else {
			go bcsDiscover.Start()
			go func() {
				for {
					select {
					case curEvent := <-bcsDiscoverEvent:
						blog.Infof("found bcs service discover event %s", curEvent)
					}
				}
			}()
		}

	}

	// create AdapterDiscover
	config.Zookeeper = strings.Replace(config.Zookeeper, ";", ",", -1)
	adapterDiscover, discoverEvent, err := rdiscover.NewAdapterDiscover(
		config.Zookeeper, config.Address, config.Cluster, config.MetricPort)
	if err != nil {
		blog.Errorf("new zookeeper %s Discover failed, err %s", config.Zookeeper, err.Error())
		return fmt.Errorf("new zookeeper %s Discover failed, err %s", config.Zookeeper, err.Error())
	}
	go adapterDiscover.Start()

	// create server
	server := NewServer(config)
	handleEvent(server, config, discoverEvent)
	return nil
}

func handleEvent(s *Server, config *Config, event <-chan rdiscover.RoleEvent) {
	signalChan := make(chan os.Signal, 5)
	signal.Notify(signalChan, syscall.SIGTRAP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	ticker := time.NewTicker(20 * time.Second)

	for {
		select {
		case curEvent := <-event:
			if curEvent == rdiscover.MasterToSlave {
				s.isMaster = false
				s.Stop()
			} else if curEvent == rdiscover.SlaveToMaster {
				s.isMaster = true
				s.Run()
			} else {
				blog.Errorf("invalid event %s", curEvent)
			}
		case <-ticker.C:
			if s.isMaster && !s.isRunning {
				blog.Infof("force run server")
				s.Run()
			}
		case sig := <-signalChan:
			blog.Warnf("bmsf-mesos-adaptor was killed, signal info: %s", sig.String())
			s.Stop()
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
	isRunning bool
	isMaster  bool
}

// NewServer create server object
func NewServer(config *Config) *Server {
	return &Server{
		config:    config,
		mgrStop:   make(chan struct{}),
		svcQueue:  queue.NewQueue(),
		nodeQueue: queue.NewQueue(),
	}
}

// createManager create manager
func (s *Server) createManager() error {

	s.mgrStop = make(chan struct{})
	s.mgr = settingManager(s.config.KubeConfig, int(s.config.MetricPort), s.svcQueue, s.nodeQueue)

	//starting manager
	go func() {
		if err := s.mgr.Start(s.mgrStop); err != nil {
			blog.Errorf("mesos-adaptor starting kubemanager failed, %s", err.Error())
			s.mgrStop <- struct{}{}
		}
	}()

	//wait for Cache ready
	blog.Infof("mesos-adaptor is waiting for kubemanager sync all cache datas.")
	caches := s.mgr.GetCache()
	if ok := caches.WaitForCacheSync(s.mgrStop); !ok {
		blog.Errorf("mesos-adaptor is waiting for cache synchronization failed, data synchronizing broken.")
		s.mgrStop <- struct{}{}
		return fmt.Errorf("data synchronization broken")
	}

	return nil
}

//Run running server loop
func (s *Server) Run() {
	// create controller manager
	if err := s.createManager(); err != nil {
		blog.Errorf("create controller manager failed, err %s", err.Error())
		return
	}

	//create cluster plugin
	clusterZkHosts := strings.Split(s.config.Zookeeper, ",")
	bkBcsCluster, err := bcs.NewCluster(s.config.Cluster, clusterZkHosts)
	if err != nil {
		blog.Errorf("init bk-bcs cluster failed, %s", err)
		return
	}
	s.cluster = bkBcsCluster
	//starting cluster
	s.cluster.AppSvcs().RegisterAppSvcQueue(s.svcQueue)
	s.cluster.AppNodes().RegisterAppNodeQueue(s.nodeQueue)
	s.cluster.Run()
	s.isRunning = true
}

//Stop stop server
func (s *Server) Stop() {
	s.cluster.Stop()
	s.mgrStop <- struct{}{}
	s.isRunning = false
	blog.Infof("Server is waiting 3 seconds for writing data back to kube-apiserver")
	time.Sleep(time.Second * 3)
}
