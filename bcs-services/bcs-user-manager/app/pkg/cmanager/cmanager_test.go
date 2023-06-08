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

package cmanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/cmanager/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/options"
)

// Test_GetProjectIDByClusterID GetProjectIDByClusterID get  projectID
func Test_GetProjectIDByClusterID(t *testing.T) {
	clusterManagerClient := ClusterManagerClient{
		cache: cache.New(time.Minute*5, time.Minute*60),
	}

	userMgrConfig := &config.UserMgrConfig{
		BcsAPI: &options.BcsAPI{
			Host:  "http://127.0.0.1:8080",
			Token: "8080",
		},
	}

	// 配置赋值
	config.SetGlobalConfig(userMgrConfig)

	// 模拟http请求
	go mockHttpRequest()
	// 保证协程先启动
	time.Sleep(time.Second)
	// 测试调用GetProjectIDByClusterID方法
	s, err := clusterManagerClient.GetProjectIDByClusterID("123455")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(s)
}

func mockHttpRequest() {
	// 定义一个测试用例
	tests := clustermanager.GetClusterResp{
		Code:    0,
		Message: "ok",
		Result:  true,
		Data: &clustermanager.Cluster{
			ClusterID: "123455",
			ProjectID: "78946132784",
		},
		Extra: &clustermanager.ExtraClusterInfo{
			ProviderType: "123",
		},
	}

	// 模拟http请求
	http.Handle("/clustermanager/v1/cluster/123455", http.HandlerFunc(func(w http.ResponseWriter,
		r *http.Request) {
		b, err := json.Marshal(tests)
		if err != nil {
			fmt.Println("error", err)
			return
		}
		w.Write(b)
	}))

	// 起一个端口
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("error", err)
		return
	}
}
