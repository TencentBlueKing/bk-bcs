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
	"flag"
	"time"

	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	bcselection "github.com/Tencent/bk-bcs/bcs-network/pkg/leaderelection"
)

func main() {
	var kubeconfig string
	flag.StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig for kube-apiserver")
	flag.Parse()

	blog.InitLogs(conf.LogConfig{
		LogDir:          "",
		LogMaxSize:      500,
		LogMaxNum:       10,
		ToStdErr:        true,
		AlsoToStdErr:    true,
		Verbosity:       0,
		StdErrThreshold: "2",
	})

	electionClient, err := bcselection.New(resourcelock.LeasesResourceLock, "test", "electiontest", kubeconfig,
		20*time.Second, 15*time.Second, 2*time.Second)
	if err != nil {
		blog.Fatalf("create election failed, err %s", err.Error())
	}
	electionClient.RunOrDie()
}
