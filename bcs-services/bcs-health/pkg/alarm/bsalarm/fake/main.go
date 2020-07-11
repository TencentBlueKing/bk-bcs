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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/master/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/alarm/bsalarm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/alarm/utils"
	"github.com/pborman/uuid"
)

func main() {
	var path, configfile string
	flag.StringVar(&path, "path", "", "plugin path")
	flag.StringVar(&configfile, "configfile", "", "kafka config file path.")
	flag.Parse()
	c := config.Config{
		KafkaConf: config.KafkaConf{
			DataID:     "9748",
			PluginPath: path,
			ConfigFile: configfile,
		},
		LogConfig: conf.LogConfig{
			LogDir:       "./glog",
			LogMaxSize:   500,
			ToStdErr:     true,
			AlsoToStdErr: true,
			Verbosity:    5,
		},
	}
	blog.InitLogs(c.LogConfig)
	bsAlarm, err := bsalarm.NewBlueShieldAlarm(c)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("new bs alarm success.")

	for {
		time.Sleep(2 * time.Second)
		op := &utils.AlarmOptions{
			AlarmID:       "demo_alarm_id",
			AlarmName:     "demo_name",
			ClusterID:     "demo_cluster_id",
			Namespace:     "demo_namespace",
			Module:        "demo_module",
			EventMessage:  "this is a test alarm event.",
			ModuleIP:      "888.888.888.888",
			ModuleVersion: "v1.0.0",
			AtTime:        time.Now().Unix(),
			UUID:          uuid.NewUUID().String(),
		}
		fmt.Println("start to send alarm.")
		if err := bsAlarm.SendAlarm(op, "888.888.888.888"); err != nil {
			fmt.Printf("send alarm failed, err: %v\n", err)
			continue
		}
		fmt.Printf("send alarm success, uuid: %s, time: %d \n", op.UUID, op.AtTime)
	}
}
