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

// Package calc xx
package calc

import (
	"context"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/cli/calchandler"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/cli/command"
)

var (
	// 用于接收传递过来的 kubeconfig 文件
	kubeConfig string

	template = `
总节点数: %d / %d          (变化: %d)
装箱率(CPU): %.2f / %.2f	（变化 %.2f）
装箱率(MEM): %.2f / %.2f	（变化 %.2f）
总核心: %.2f / %.2f	（变化：%.2f）
总内存: %.2f / %.2f （变化：%.2f）
`
)

// NewCalcCmd return calc command
func NewCalcCmd() *cobra.Command {
	calcCmd := &cobra.Command{
		Use:   "calc",
		Short: "calc the cluster resource with remote and local",
	}
	calcCmd.AddCommand(remoteCmd())
	return calcCmd
}

func remoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remote",
		Short: "calculator from remote",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if err := command.InitConfig(); err != nil {
				panic(err)
			}
			cacheManager, err := command.InitCacheManager(ctx, kubeConfig)
			if err != nil {
				panic(err)
			}
			handler := calchandler.NewCalculatorHandler(ctx, cacheManager)
			rate, err := handler.Calc()
			if err != nil {
				command.Exit("calc failed: %s", err.Error())
			}
			tw := tablewriter.NewWriter(os.Stdout)
			tw.SetHeader(func() []string {
				return []string{
					"名称", "优化前", "装箱后", "变化",
				}
			}())
			tw.Append(func() []string {
				return []string{
					"总节点数",
					fmt.Sprintf("%d", len(rate.OriginalRate.NodePackingRate)),
					fmt.Sprintf("%d", len(rate.OptimizedRate.NodePackingRate)),
					fmt.Sprintf("%d", len(rate.OriginalRate.NodePackingRate)-len(rate.OptimizedRate.NodePackingRate)),
				}
			}())
			tw.Append(func() []string {
				return []string{
					"装箱率(cpu)",
					fmt.Sprintf("%.2f", rate.OriginalRate.TotalRate.Cpu),
					fmt.Sprintf("%.2f", rate.OptimizedRate.TotalRate.Cpu),
					fmt.Sprintf("%.2f", rate.OptimizedRate.TotalRate.Cpu-rate.OriginalRate.TotalRate.Cpu),
				}
			}())
			tw.Append(func() []string {
				return []string{
					"装箱率(mem)",
					fmt.Sprintf("%.2f", rate.OriginalRate.TotalRate.Mem),
					fmt.Sprintf("%.2f", rate.OptimizedRate.TotalRate.Mem),
					fmt.Sprintf("%.2f", rate.OptimizedRate.TotalRate.Mem-rate.OriginalRate.TotalRate.Mem),
				}
			}())
			tw.Append(func() []string {
				return []string{
					"总核心",
					fmt.Sprintf("%.2f", rate.OriginalRate.TotalRate.CpuCapacity),
					fmt.Sprintf("%.2f", rate.OptimizedRate.TotalRate.CpuCapacity),
					fmt.Sprintf("%.2f", rate.OriginalRate.TotalRate.CpuCapacity-rate.OptimizedRate.TotalRate.CpuCapacity),
				}
			}())
			tw.Append(func() []string {
				return []string{
					"总内存",
					fmt.Sprintf("%.2f", rate.OriginalRate.TotalRate.MemCapacity/1024/1024/1024),
					fmt.Sprintf("%.2f", rate.OptimizedRate.TotalRate.MemCapacity/1024/1024/1024),
					fmt.Sprintf("%.2f", rate.OriginalRate.TotalRate.MemCapacity/1024/1024/1024-
						rate.OptimizedRate.TotalRate.MemCapacity/1024/1024/1024),
				}
			}())
			tw.Render()

			fmt.Println()
			fmt.Printf("(1) 优化节点列表\n")
			fmt.Printf("    %v\n", rate.OptimizedNodes)
			fmt.Println()
		},
	}
	cmd.PersistentFlags().StringVarP(&kubeConfig, "kubeconfig", "k", "",
		"the config of kubernetes cluster")
	cmd.MarkPersistentFlagRequired("kubeconfig")
	return cmd
}
