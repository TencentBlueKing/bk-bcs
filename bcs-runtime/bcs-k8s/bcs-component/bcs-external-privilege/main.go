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
	"os"
	"runtime"
	"sync"

	"github.com/TencentBlueKing/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-external-privilege/common"
	"github.com/TencentBlueKing/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-external-privilege/gcs"
	"github.com/TencentBlueKing/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-external-privilege/scr"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

const failRetryLimit = 40

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	option := common.LoadOption()
	blog.InitLogs(conf.LogConfig{ToStdErr: true, Verbosity: 3})

	var wg sync.WaitGroup
	var succ = true
	for _, v := range option.DBPrivEnvList {
		wg.Add(1)

		go func(env common.DBPrivEnv) {
			blog.Infof("Starting granting privilege to db: %s, dbname: %s", env.TargetDb, env.DbName)
			defer wg.Done()
			var failRetry = 0
			var jobId string
			client, err := env.InitClient(option)
			if err != nil {
				blog.Errorf("failed to init client for external system, %v", err)
				os.Exit(1)
			}
			for failRetry < failRetryLimit {
				err = client.DoPri(option, &env)
				if !env.IsSCR() {
					jobId, err = gcs.PrivilegeRequest(option, env)
				} else {
					jobId, err = scr.PrivilegeRequest(option, env)
				}
				if err == nil && jobId != "" {
					break
				}
				blog.Errorf("error calling the privilege api: %s, db: %s, dbname: %s, retry %d", err.Error(), env.TargetDb, env.DbName, failRetry)
				failRetry++
				continue
			}
			if failRetry >= failRetryLimit {
				blog.Errorf("error calling the privilege api with db: %s, dbname: %s", env.TargetDb, env.DbName)
				succ = false
				return
			}
			if !env.IsSCR() {
				err = gcs.CheckFinalStatus(option, jobId, failRetryLimit)
			} else {
				err = scr.CheckFinalStatus(option, jobId, failRetryLimit)
			}
			if err != nil {
				blog.Errorf("error to check whether the privilege succeed: %s, db: %s, dbname: %s", err.Error(), env.TargetDb, env.DbName)
				succ = false
				return
			}
			blog.Infof("Granting privilege to db: %s, dbname: %s succ", env.TargetDb, env.DbName)
		}(v)
	}
	wg.Wait()
	if !succ {
		os.Exit(1)
	}
	os.Exit(0)
}
