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

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller"
)

func main() {
	cfgHandler := options.GlobalConfigHandler()
	if err := cfgHandler.Parse(); err != nil {
		blog.Fatalf("parse options failed: %s", err.Error())
	}
	defer blog.CloseLogs()

	mgr := controller.NewControllerManager()
	if err := mgr.Init(); err != nil {
		panic(err)
	}
	ctx := ctrl.SetupSignalHandler()
	if err := mgr.Run(ctx); err != nil {
		os.Exit(1)
	}
}
