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
	"runtime"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/openapi"

	_ "github.com/go-openapi/loads"
	_ "go.uber.org/automaxprocs"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog"
	"sigs.k8s.io/apiserver-builder-alpha/pkg/cmd/server"
)

func main() {

	klog.Info("real GOMAXPROCS", runtime.GOMAXPROCS(-1))
	version := "v0.1.0"

	err := server.StartApiServerWithOptions(&server.StartOptions{
		EtcdPath:    "/registry/federated.bkbcs.tencent.com",
		Apis:        apis.GetAllApiBuilders(),
		Openapidefs: openapi.GetOpenAPIDefinitions,
		Title:       "Api",
		Version:     version,
	})
	if err != nil {
		panic(err)
	}
}
