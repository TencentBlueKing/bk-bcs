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
	"strconv"
	"strings"
	//"sync"
	"os"
	//"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/apiserver"
	//"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/ipam"
	//"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/ns"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/schedcontext"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/util"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/http/httpserver"
)

type Manager struct {
	store        *store.Store
	sched        *sched.Sched
	schedContext *schedcontext.SchedContext
	config       util.SchedConfig
}

func New(config util.SchedConfig) (*Manager, error) {
	manager := &Manager{
		config: config,
	}

	dbzk := store.NewDbZk(strings.Split(config.ZkHost, ","))
	dbzk.Connect()

	zkStore := store.NewManagerStore(dbzk)

	manager.schedContext = &schedcontext.SchedContext{
		Config: config,
		Store:  zkStore,
	}

	listener := &manager.config.HttpListener
	blog.Info("NewHttpServer: %s %s, SSL(%s)", listener.TCPAddr, listener.UnixAddr, strconv.FormatBool(listener.IsSSL))

	splitID := strings.Split(listener.TCPAddr, ":")
	if len(splitID) < 2 {
		return nil, fmt.Errorf("listen adress %s format error", listener.TCPAddr)
	}
	ip := splitID[0]
	port, err := strconv.Atoi(splitID[1])
	if err != nil {
		blog.Error("get port from %s error: %s", listener.TCPAddr, err.Error())
		return nil, fmt.Errorf("listen adress %s format error", listener.TCPAddr)
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

	manager.config.Scheduler.Address = listener.TCPAddr
	manager.config.Scheduler.ZK = config.ZkHost
	manager.sched = sched.New(manager.config.Scheduler, manager.schedContext)

	blog.Info("New manager finish")

	return manager, nil
}

func (manager *Manager) Stop() error {
	return nil
}

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
