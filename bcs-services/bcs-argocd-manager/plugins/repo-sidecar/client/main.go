/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
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
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/plugins/repo-sidecar/client/pack"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/plugins/repo-sidecar/server/service"
)

func main() {
	data, err := pack.New().Pack(".")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "call plugin sidecar do pack failed, %s", err.Error())
		os.Exit(1)
	}

	message := service.Message{
		Env:     os.Environ(),
		Args:    os.Args,
		Content: data,
	}

	_, err = message.Request()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "call plugin sidecar failed, %s", err.Error())
		os.Exit(1)
	}

	os.Exit(1)
	return
}
