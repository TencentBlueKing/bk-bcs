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
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/healthz"
	"fmt"
)

func main() {
	c := conf.CertConfig{}
	cli, err := healthz.NewHealthzClient([]string{"127.0.0.1:2181"}, c)
	if err != nil {
		fmt.Printf("new healthz client failed, err:%v\n", err)
		return
	}
	hCtl, _ := healthz.NewHealthCtrl(cli)
	_, err = hCtl.PackageHealthResult()
	if err != nil {
		fmt.Printf("\n%v\n", err)
	}
}
