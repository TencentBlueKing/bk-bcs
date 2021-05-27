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

package manager

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/remote/alertmanager"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/schedcontext"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store/etcd"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store/zk"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/pluginManager"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"
)

// Mananger main body of scheduler
type Manager struct {
	sched        *sched.Sched
	schedContext *schedcontext.SchedContext
	config       util.SchedConfig
}

// New create Manager according config item
func New(config util.SchedConfig) (*Manager, error) {
	manager := &Manager{
		config: config,
	}

	var s store.Store
	var err error
	var pm *pluginManager.PluginManager
	if config.Scheduler.Plugins != "" {
		blog.Infof("start init plugin manager")
		plugins := strings.Split(config.Scheduler.Plugins, ",")

		pm, err = pluginManager.NewPluginManager(plugins, config.Scheduler.PluginDir)
		if err != nil {
			blog.Errorf("NewPluginManager error %s", err.Error())
		}
	}

	if config.Scheduler.StoreDriver == "etcd" {
		s, err = etcd.NewEtcdStore(config.Scheduler.Kubeconfig, pm, config.Scheduler.Cluster)
		if err != nil {
			blog.Errorf("new etcd store failed: %s", err.Error())
			return nil, err
		}
		config.Scheduler.UseCache = true
	} else {
		dbzk := zk.NewDbZk(strings.Split(config.ZkHost, ","))
		err = dbzk.Connect()
		if err != nil {
			blog.Errorf("connect zookeeper %s failed: %s", config.ZkHost, err.Error())
			return nil, err
		}
		s = zk.NewManagerStore(dbzk, pm, config.Scheduler.Cluster)
		config.Scheduler.UseCache = false
	}

	manager.schedContext = &schedcontext.SchedContext{
		Config: config,
		Store:  s,
	}

	listener := &manager.config.HttpListener
	blog.Info("NewHttpServer: %s %s, SSL(%s)", listener.TCPAddr, listener.UnixAddr, strconv.FormatBool(listener.IsSSL))

	splitID := strings.Split(listener.TCPAddr, ":")
	if len(splitID) < 2 {
		return nil, fmt.Errorf("listen address %s format error", listener.TCPAddr)
	}
	ip := splitID[0]
	port, err := strconv.Atoi(splitID[1])
	if err != nil {
		blog.Error("get port from %s error: %s", listener.TCPAddr, err.Error())
		return nil, fmt.Errorf("listen address %s format error", listener.TCPAddr)
	}

	manager.schedContext.ApiServer2 = httpserver.NewHttpServer(uint(port), ip, listener.UnixAddr)
	if manager.schedContext.ApiServer2 == nil {
		blog.Error("NewHttpServer: %s:%d %s fail", ip, port, listener.UnixAddr)
	} else {
		blog.Info("NewHttpServer: %s:%d %s succ", ip, port, listener.UnixAddr)
	}

	if listener.IsSSL {
		blog.Info("Set SSL for HttpServer: CA(%s) Cert(%s) Key(%s)", listener.CAFile, listener.CertFile, listener.KeyFile)
		manager.schedContext.ApiServer2.SetSsl(listener.CAFile, listener.CertFile, listener.KeyFile, listener.CertPasswd)
	}

	if len(config.AlertManager.Server) != 0 {
		alertClient, err := alertmanager.NewAlertManager(alertmanager.Options{
			Server:     config.AlertManager.Server,
			ClientAuth: config.AlertManager.ClientAuth,
			Debug:      config.AlertManager.Debug,
			Token:      config.AlertManager.Token,
		})
		if err != nil {
			blog.Errorf("NewAlertManager failed: %v", err)
			return nil, err
		}
		blog.Infof("alertmanager init successful")
		manager.schedContext.AlertManager = alertClient
	} else {
		blog.Warnf("alertmanager server address is empty, alertmanager is disabled")
	}
	

	manager.config.Scheduler.Address = listener.TCPAddr
	manager.config.Scheduler.ZK = config.ZkHost
	manager.sched = sched.New(manager.config.Scheduler, manager.schedContext)

	blog.Info("New manager finish")

	return manager, nil
}

// Stop stop manager
func (manager *Manager) Stop() error {
	return nil
}

// Start entry point of bcs-scheduler
func (manager *Manager) Start() error {

	err := manager.sched.Start()
	if err != nil {
		os.Exit(1)
	}

	blog.Info("to run http server")
	err = manager.schedContext.ApiServer2.ListenAndServe()
	if err != nil {
		blog.Info("run http server err:%s", err.Error())
		os.Exit(1)
	}

	return nil
}
