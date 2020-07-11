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

package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/bcs-health/api"
	"github.com/Tencent/bk-bcs/bcs-common/common/bcs-health/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"os"
)

func main() {
	var zkAddr, clusterID string
	var tls api.TLSConfig
	flag.StringVar(&zkAddr, "zkaddr", "127.0.0.1:2181", "bcs zk address")
	flag.StringVar(&clusterID, "clusterid", "demo-clusterid", "cluster id value.")
	flag.StringVar(&tls.CaFile, "cafile", "", "ca file path")
	flag.StringVar(&tls.CertFile, "certfile", "", "cert file path")
	flag.StringVar(&tls.KeyFile, "keyfile", "", "key file path")
	flag.Parse()

	dir, _ := os.Getwd()

	c := conf.LogConfig{
		LogDir:       dir + "/blog",
		LogMaxSize:   500,
		ToStdErr:     true,
		AlsoToStdErr: true,
		Verbosity:    5,
	}
	blog.InitLogs(c)

	if err := api.NewBcsHealth(zkAddr, tls); nil != err {
		fmt.Printf("new bcs health instance failed. err: %v\n", err)
		return
	}
	fmt.Println("new health demo success.")

	for {
		time.Sleep(3 * time.Second)
		health := api.HealthInfo{
			Module:    "fake_health_client",
			Kind:      api.InfoKind,
			IP:        "999.999.999.999",
			ClusterID: clusterID,
			Version:   "demo-version",

			Namespace:     "demo_namespace",
			Message:       fmt.Sprintf("now is %s", time.Now().String()),
			AlarmName:     "bcs-health-time-report-demo",
			Affiliation:   types.Both,
			AppAlarmLevel: "important",
		}
		if err := api.SendHealthInfo(&health); nil != err {
			fmt.Printf("send health info failed. err: %v \n", err)
			continue
		}
		fmt.Println("send alarm success.")

	}

}
