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

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var availableCMD = &cobra.Command{
	Use:   "available",
	Short: "check available",
	Long:  "check if bcs-helm-manager service is available",
	Run:   Available,
}

// Available provide the actions to do availableCMD
func Available(cmd *cobra.Command, _ []string) {
	c := newClientWithConfiguration()
	if err := c.Available(cmd.Context()); err != nil {
		fmt.Printf("check bcs-helm-manager available failed, %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("bcs-helm-manager is available\n")
}
