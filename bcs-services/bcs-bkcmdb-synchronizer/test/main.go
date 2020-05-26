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
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"bk-bcs/bcs-common/common/ssl"
	"bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/storage"
)

func init() {
	pflag.String("zk", "127.0.0.1:2181", "zk address")
	pflag.String("command", "list", "command")
	pflag.String("ctype", "mesos", "")
	pflag.String("cid", "", "")
	pflag.String("rtype", "taskgroup", "")
	pflag.String("ca", "", "")
	pflag.String("cert", "", "")
	pflag.String("key", "", "")
	pflag.String("certpwd", "", "")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}
func main() {
	cli, err := storage.NewStorageClient(viper.GetString("zk"))
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	command := viper.GetString("command")
	ca := viper.GetString("ca")
	cert := viper.GetString("cert")
	key := viper.GetString("key")
	certpwd := viper.GetString("certpwd")

	tlsConf, err := ssl.ClientTslConfVerity(ca, cert, key, certpwd)
	if err != nil {
		panic(err)
	}

	cli.SetTLSConfig(tlsConf)

	switch command {
	case "list":
		retList, err := cli.ListResources(viper.GetString("ctype"), viper.GetString("cid"), viper.GetString("rtype"))
		if err != nil {
			panic(err)
		}
		retData, _ := json.Marshal(retList)
		fmt.Printf("%s", retData)
	case "watch":
		ch, err := cli.WatchClusterResources(viper.GetString("cid"), viper.GetString("rtype"))
		if err != nil {
			panic(err)
		}
		for {
			select {
			case tmp := <-ch:
				fmt.Printf("%#v\n", tmp)
			}
		}
	default:
		panic(command)
	}
}
