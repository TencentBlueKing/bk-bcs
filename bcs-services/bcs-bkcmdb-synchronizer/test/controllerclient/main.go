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
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/conf"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/common/version"
	"bk-bcs/bcs-common/common/zkclient"
	"bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/controller"
	"bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/discovery"
	"bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/taskinformer"
	"bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/taskmanager"
)

func init() {
	pflag.String("zk", "", "")
	pflag.String("env", "", "")
	pflag.String("clusterEnv", "", "")
	pflag.String("paascc", "", "")
	pflag.String("appcode", "", "")
	pflag.String("appsecret", "", "")
	pflag.Int64("interval", 15, "")
	pflag.String("ip", "127.0.0.1", "")
	pflag.Int("port", 8089, "")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	blog.InitLogs(conf.LogConfig{
		Verbosity:       5,
		LogDir:          "./logs",
		LogMaxSize:      500,
		LogMaxNum:       20,
		StdErrThreshold: "2",
		AlsoToStdErr:    true,
	})
}

func main() {
	serverInfo := &types.ServerInfo{
		IP:      viper.GetString("ip"),
		Port:    uint(viper.GetInt("port")),
		Scheme:  "http",
		Version: version.GetVersion(),
	}

	zkAddr := strings.Replace(viper.GetString("zk"), ";", ",", -1)
	zkAddrs := strings.Split(zkAddr, ",")
	zkcli := zkclient.NewZkClient(zkAddrs)
	err := zkcli.Connect()
	if err != nil {
		panic(err)
	}

	disc := discovery.New(viper.GetString("zk"), "", serverInfo)
	go disc.Run()

	m, err := taskmanager.NewManager(
		viper.GetString("env"),
		viper.GetString("clusterEnv"),
		viper.GetString("paascc"),
		viper.GetString("appcode"),
		viper.GetString("appsecret"),
		viper.GetInt("interval"),
		disc,
		zkcli,
	)
	if err != nil {
		panic(err)
	}

	informer := taskinformer.NewInformer(serverInfo, zkcli)

	ctrl := controller.NewController(disc, informer, m)

	ctx := context.Background()
	ctrl.Run(ctx)
}
