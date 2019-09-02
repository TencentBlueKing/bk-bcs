/*
Copyright (C) 2019 The BlueKing Authors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package app

import (
	"bk-bcs/bmsf-mesh/bmsf-mesos-adapter/controller"
	"bk-bcs/bmsf-mesh/bmsf-mesos-adapter/discovery"
	"bk-bcs/bmsf-mesh/bmsf-mesos-adapter/discovery/bcs"
	"bk-bcs/bmsf-mesh/bmsf-mesos-adapter/pkg/queue"
	"bk-bcs/bmsf-mesh/pkg/apis"
	"bk-bcs/bcs-common/common/blog"
	pidfile "bk-bcs/bcs-common/common"
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
	s := &Server{
		config:    config,
		mgrStop:   make(chan struct{}),
		svcQueue:  queue.NewQueue(),
		nodeQueue: queue.NewQueue(),
	}
	//todo: init event messge Queue for Reconciler & bk-bcs cluster
	s.mgr = settingManager(config.KubeConfig, int(config.MetricPort), s.svcQueue, s.nodeQueue)
	//create cluster plugin
	zkHosts := strings.Split(config.BCSZk, ",")
	bkBcsCluster, err := bcs.NewCluster(config.Cluster, zkHosts)
	if err != nil {
		blog.Errorf("init bk-bcs cluster failed, %s", err)
		return err
	}
	s.cluster = bkBcsCluster
	//ready to run
	s.HandleSignal()
	return s.Run()
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

//HandleSignal handle system exit signals
func (s *Server) HandleSignal() {
	signalChan := make(chan os.Signal, 5)
	signal.Notify(signalChan, syscall.SIGTRAP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	go func() {
		select {
		case sig := <-signalChan:
			blog.Warnf("bmsf-mesos-adaptor was killed, signal info: %s", sig.String())
			s.Stop()
		}
	}()
}
