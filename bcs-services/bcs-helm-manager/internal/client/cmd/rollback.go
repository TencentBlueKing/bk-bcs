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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"os"
	"strconv"

	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"

	"github.com/spf13/cobra"
)

var rollbackCMD = &cobra.Command{
	Use:   "rollback",
	Short: "rollback",
	Long:  "rollback chart release",
	Run:   Rollback,
}

// Rollback provide the actions to do rollbackCMD
func Rollback(cmd *cobra.Command, args []string) {
	req := &helmmanager.RollbackReleaseReq{}

	if len(args) < 2 {
		fmt.Printf("rollback args need at least 2, install [name] [revision]\n")
		os.Exit(1)
	}
	revision, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("rollback get invalid revision %s, %s\n", args[1], err.Error())
		os.Exit(1)
	}
	if revision <= 0 {
		fmt.Printf("rollback get invalid revision %s, revision should be positive\n", args[1])
		os.Exit(1)
	}

	req.Name = common.GetStringP(args[0])
	req.Namespace = &flagNamespace
	req.ClusterID = &flagCluster
	req.Revision = common.GetUint32P(uint32(revision))

	c := newClientWithConfiguration()
	if err := c.Release().Rollback(cmd.Context(), req); err != nil {
		fmt.Printf("rollback release failed, %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("success to rollback release %s namespace %s cluster %s to revision %d\n",
		req.GetName(), req.GetNamespace(), req.GetClusterID(), revision)
}

func init() {
	rollbackCMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace for operation")
	rollbackCMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "", "", "release cluster id for operation")
}
