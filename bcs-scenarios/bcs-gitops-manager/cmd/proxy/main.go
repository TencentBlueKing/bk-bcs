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
	goflag "flag"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/proxy/server"
)

func main() {
	command := server.NewCommand()
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "gitops proxy starting failed %s\n", err.Error())
		os.Exit(1)
	}
}
