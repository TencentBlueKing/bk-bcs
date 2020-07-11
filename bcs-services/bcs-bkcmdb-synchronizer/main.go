/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/esb/apigateway/paascc"
	cmdb "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/cmdbv3"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/controller"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/taskinformer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/taskmanager"
)

func main() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
		debug.PrintStack()
	}()

	ops := new(config.SyncOption)
	ops.Load()

	isValid, msg := ops.Validate()
	if !isValid {
		fmt.Printf("validate options failed, msg %s\n", msg)
		os.Exit(-1)
	}

	blog.InitLogs(ops.LogConfig)

	var err error
	var tlsConfig *tls.Config
	if len(ops.StorageCa) != 0 || len(ops.StorageCert) != 0 || len(ops.StorageKey) != 0 {
		tlsConfig, err = ssl.ClientTslConfVerity(
			ops.StorageCa,
			ops.StorageCert,
			ops.StorageKey,
			static.ClientCertPwd,
		)
		if err != nil {
			blog.Errorf("load storage client tls config failed, err %s", err.Error())
			os.Exit(-1)
		}
	}

	blog.Infof("create bcs storage client")
	storageClient, err := storage.NewStorageClient(ops.StorageZk)
	if err != nil {
		blog.Errorf("create new storage client failed, err %s", err.Error())
		os.Exit(-1)
	}
	if tlsConfig != nil {
		storageClient.SetTLSConfig(tlsConfig)
	}
	storageClient.Start(context.Background())

	serverInfo := new(types.ServerInfo)
	serverInfo.IP = ops.Address
	serverInfo.Port = ops.Port
	if len(ops.ServerCertFile) != 0 || len(ops.ServerKeyFile) != 0 || len(ops.CAFile) != 0 {
		serverInfo.Scheme = "https"
	} else {
		serverInfo.Scheme = "http"
	}

	blog.Infof("create discovery")
	zkAddr := strings.Replace(ops.ZkAddr, ";", ",", -1)
	disc := discovery.New(zkAddr, "", serverInfo)
	go disc.Run()

	zkAddrs := strings.Split(zkAddr, ",")
	zkcli := zkclient.NewZkClient(zkAddrs)
	err = zkcli.Connect()
	if err != nil {
		blog.Errorf("failed to connect zk server, err %s", err.Error())
		os.Exit(-1)
	}

	paasccClient := paascc.NewClientInterface(ops.PaasAddr, ops.PaasAppCode, ops.PaasAppSecret, nil)
	blog.Infof("create task manager")
	manager, err := taskmanager.NewManager(
		ops.PaasEnv,
		ops.PaasClusterEnv,
		int(ops.ClusterPullInterval),
		disc,
		zkcli,
		paasccClient,
	)
	if err != nil {
		blog.Infof("failed to create task manager, err %s", err.Error())
	}

	blog.Infof("create task informer")
	informer := taskinformer.NewInformer(serverInfo, zkcli)

	blog.Infof("create cmdb client")
	cmdbClient := cmdb.NewClientInterface(ops.CmdbAddr, nil)
	cmdbClient.SetDefaultHeader(http.Header{
		"Host":                      []string{ops.CmdbAddr},
		"HTTP_BLUEKING_SUPPLIER_ID": []string{ops.CmdbSupplierID},
		"BK_User":                   []string{ops.CmdbUser},
	})
	blog.Infof("create controller")
	controller := controller.NewController(
		ops,
		serverInfo,
		disc,
		storageClient,
		cmdbClient,
		informer,
		manager,
	)

	blog.Infof("start controller")
	ctx := context.Background()
	controller.Run(ctx)
}
