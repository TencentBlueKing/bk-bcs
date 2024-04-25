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
 */

package main

import (
	"os"
	"runtime"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"

	"github.com/TencentBlueKing/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-external-privilege/common"
	"github.com/TencentBlueKing/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-external-privilege/pkg"
)

const failRetryLimit = 40

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	option := common.LoadOption()
	blog.InitLogs(conf.LogConfig{ToStdErr: true, Verbosity: 3})

	var wg sync.WaitGroup
	var success = false
	for _, v := range option.DBPrivEnvList {
		wg.Add(1)
		go func(env common.DBPrivEnv) {
			blog.Infof("starting granting privilege to db: %s, dbname: %s", env.TargetDb, env.DbName)
			defer wg.Done()
			var doPriRetry, checkRetry = 0, 0
			client, err := pkg.InitClient(option, &env)
			if err != nil {
				blog.Errorf("failed to init client for external system, %v", err)
				return
			}

			for doPriRetry < failRetryLimit {
				err = client.DoPri(option, &env)
				if err == nil {
					break
				}
				blog.Errorf("error calling the privilege api: %s, db: %s, dbname: %s, retry %d",
					err.Error(), env.TargetDb, env.DbName, doPriRetry)
				doPriRetry++
			}
			if doPriRetry >= failRetryLimit {
				blog.Errorf("error calling the privilege api with db: %s, dbname: %s, max retry times reached",
					env.TargetDb, env.DbName)
				return
			}

			for checkRetry < failRetryLimit {
				common.WaitForSeveralSeconds()
				err = client.CheckFinalStatus()
				if err == nil {
					break
				}
				blog.Errorf("check operation status failed: %s, db: %s, dbname: %s, retry %d",
					err.Error(), env.TargetDb, env.DbName, checkRetry)
				checkRetry++
			}
			if checkRetry >= failRetryLimit {
				blog.Errorf("check operation status failed with db: %s, dbname: %s, max retry times reached",
					env.TargetDb, env.DbName)
				return
			}

			success = true
			blog.Infof("granting privilege to db: %s, dbname: %s succeeded", env.TargetDb, env.DbName)
		}(v)
	}
	wg.Wait()

	if !success {
		os.Exit(1)
	}

	os.Exit(0)
}
