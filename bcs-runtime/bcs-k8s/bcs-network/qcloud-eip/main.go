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
	"fmt"
	"runtime"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/qcloud-eip/eip"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/qcloud-eip/option"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"
)

var CmdOption *option.Option
var app eip.Interface

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
	CmdOption = &option.Option{}
	conf.Parse(CmdOption)
	app = &eip.EIP{}
}

func cmdAdd(args *skel.CmdArgs) error {
	return app.CNIAdd(args)
}

func cmdDel(args *skel.CmdArgs) error {
	return app.CNIDel(args)
}

func main() {
	if CmdOption.WorkMode == option.WorkModeCNI {
		CmdOption.LogDir = option.DefaultConfigLogDir
		blog.InitLogs(CmdOption.LogConfig)
		defer blog.CloseLogs()
		skel.PluginMain(cmdAdd, cmdDel, version.All)
		return
	}
	if len(CmdOption.ConfigFile) == 0 {
		CmdOption.ConfigFile = option.DefaultConfigFilePath
	}
	if CmdOption.LogDir == "./logs" {
		CmdOption.LogDir = option.DefaultConfigLogDir
	}
	blog.InitLogs(CmdOption.LogConfig)
	defer blog.CloseLogs()
	switch CmdOption.WorkMode {
	case option.WorkModeInit:
		app.Init(CmdOption.ConfigFile, CmdOption.EniNum, CmdOption.IPNum)
	case option.WorkModeRecover:
		app.Recover(CmdOption.ConfigFile, CmdOption.EniNum)
	case option.WorkModeRelease:
		app.Release(CmdOption.ConfigFile)
	case option.WorkModeClean:
		app.Clean(CmdOption.ConfigFile)
	default:
		fmt.Printf("unknown work mode %s", CmdOption.WorkMode)
	}

}
