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
	"context"
	"fmt"
	"log"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/cli-util/validate-listener-name/pkg"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = networkextensionv1.AddToScheme(scheme)
}

func main() {

	opts := &pkg.ControllerOption{}
	opts.BindFromCommandLine()

	ctx := context.Background()
	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0", // "0"表示禁用默认的Metric Service， 需要使用自己的实现支持IPV6
		LeaderElection:     false,
	})
	if err != nil {
		panic(err)
	}

	go func() {
		if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			log.Fatalf("start manager failed, err: %s", err.Error())
		}
	}()
	log.Println("wait for cache sync...")
	if !mgr.GetCache().WaitForCacheSync(ctx.Done()) {
		log.Fatalf("WaitForCacheSync failed")
	}
	log.Println("cache sync success")

	handler, err := pkg.NewHandler(ctx, mgr.GetClient(), opts)
	if err != nil {
		log.Fatalf(err.Error())
	}

	lbListenerMap, notReadyListeners, err := handler.LoadListener()
	if err != nil {
		log.Fatalf(err.Error())
	}

	if len(notReadyListeners) != 0 {
		for _, li := range notReadyListeners {
			log.Printf("%s/%s \n", li.GetNamespace(), li.GetName())
		}
		log.Fatalf("以上监听器状态异常/没有云监听器ID， 请根据监听器status确认异常")
	}

	invalidListenerMap := make(map[string][]*tclb.Listener)
	for lbID, liList := range lbListenerMap {
		invalidListenerList, err1 := handler.CheckListenerName(lbID, liList)
		if err1 != nil {
			log.Fatalf("校验监听器名称失败, err: %s", err1.Error())
		}

		for _, li := range invalidListenerList {
			log.Printf("监听器[%s-%s-%d]名称[%s]不符合规范...", lbID, *li.Protocol, *li.Port, *li.ListenerName)
		}

		invalidListenerMap[lbID] = invalidListenerList
	}

	log.Println("更新以上监听器名称？(监听器数量较多时可能处理时间较长)[y/n] ")
	userFlag := ""
	if _, err = fmt.Scanln(&userFlag); err != nil {
		log.Fatalf("read user flag failed, err: %s", err.Error())
	}
	if userFlag != "y" && userFlag != "Y" {
		return
	}

	if err1 := handler.BatchUpdateListenerName(invalidListenerMap); err1 != nil {
		log.Fatalf("批量更新监听器名称失败, err: %s", err1.Error())
	}

	log.Println("批量更新监听器名称成功")
}
