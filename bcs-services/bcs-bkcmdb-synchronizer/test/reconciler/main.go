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
	"net/http"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/conf"
	"bk-bcs/bcs-common/common/ssl"
	"bk-bcs/bcs-common/pkg/esb/cmdbv3"
	"bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/common"
	"bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/reconciler"
	"bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/storage"
)

func init() {
	pflag.String("host", "", "")
	pflag.Int64("biz", 0, "")
	pflag.Int64("moduleid", 0, "")
	pflag.String("cluster", "", "")
	pflag.String("interval", "", "")

	pflag.String("zk", "127.0.0.1:2181", "")
	pflag.String("ca", "", "")
	pflag.String("cert", "", "")
	pflag.String("key", "", "")

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
	cli, err := storage.NewStorageClient(viper.GetString("zk"))
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	ca := viper.GetString("ca")
	cert := viper.GetString("cert")
	key := viper.GetString("key")
	certpwd := viper.GetString("certpwd")

	tlsConf, err := ssl.ClientTslConfVerity(ca, cert, key, certpwd)
	if err != nil {
		panic(err)
	}

	cli.SetTLSConfig(tlsConf)

	clusterInfo := common.Cluster{
		ClusterID:       viper.GetString("cluster"),
		BizID:           viper.GetInt64("biz"),
		DefaultModuleID: viper.GetInt64("moduleid"),
	}

	client := cmdbv3.NewClientInterface(viper.GetString("host"), nil)
	client.SetDefaultHeader(http.Header{
		"Host":                      []string{"cmdb.test.com"},
		"HTTP_BLUEKING_SUPPLIER_ID": []string{"0"},
		"BK_User":                   []string{"admin"},
	})

	rl, err := reconciler.NewReconciler(clusterInfo, cli, client, 10*time.Second)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	rl.Run(ctx)

}
