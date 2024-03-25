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

package httpserver

import (
	"context"
	"fmt"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/cli/calchandler"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/cli/command"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/cachemanager"
)

var (
	// 用于 http server 启动的端口
	httpPort int
	// 用于接收传递过来的 kubeconfig 文件
	kubeConfig string
)

func NewHTTPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "this will run the http server",
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
			server := &httpServer{
				ctx:          ctx,
				httpPort:     httpPort,
				cacheManager: cacheManager,
			}
			if err := server.run(); err != nil {
				command.Exit("http server closed with err: %s", err.Error())
			}
		},
	}
	cmd.PersistentFlags().StringVarP(&kubeConfig, "kubeconfig", "k", "",
		"the config of kubernetes cluster")
	cmd.PersistentFlags().IntVarP(&httpPort, "port", "p", 8080,
		"the port of http server")
	return cmd
}

type httpServer struct {
	ctx          context.Context
	httpPort     int
	cacheManager cachemanager.CacheInterface
}

func (s *httpServer) run() error {
	r := gin.Default()
	handler := calchandler.NewCalculatorHandler(s.ctx, s.cacheManager)
	rate, err := handler.Calc()
	if err != nil {
		panic(err)
	}
	result := &HttpResult{
		TargetPackingRate: 80,
		NodeNum: []*NodeNumObj{
			{
				Kind: "可优化节点",
				Num:  len(rate.OriginalRate.NodePackingRate) - len(rate.OptimizedRate.NodePackingRate),
			}, {
				Kind: "剩余节点数",
				Num:  len(rate.OptimizedRate.NodePackingRate),
			},
		},
		CPUPackingRate: []*PackingRateObj{
			{
				Kind: "优化前",
				Rate: rate.OriginalRate.TotalRate.Cpu,
			}, {
				Kind: "优化后",
				Rate: rate.OptimizedRate.TotalRate.Cpu,
			},
		},
		MEMPackingRate: []*PackingRateObj{
			{
				Kind: "优化前",
				Rate: rate.OriginalRate.TotalRate.Mem,
			}, {
				Kind: "优化后",
				Rate: rate.OptimizedRate.TotalRate.Mem,
			},
		},
		CPUCapacity: []*CapacityObj{
			{
				Kind:     "优化前",
				Capacity: rate.OriginalRate.TotalRate.CpuCapacity,
			}, {
				Kind:     "优化后",
				Capacity: rate.OptimizedRate.TotalRate.CpuCapacity,
			},
		},
		MEMCapacity: []*CapacityObj{
			{
				Kind:     "优化前",
				Capacity: rate.OriginalRate.TotalRate.MemCapacity / 1024 / 1024 / 1024,
			}, {
				Kind:     "优化后",
				Capacity: rate.OptimizedRate.TotalRate.MemCapacity / 1024 / 1024 / 1024,
			},
		},
		OptimizedNode: make([]NodeInfo, 0, len(rate.OptimizedNodes)),
	}
	result.OptimizePrice = []PriceObj{
		{
			Kind:  "优化前",
			Value: fmt.Sprintf("%.2f", (rate.OriginalRate.TotalRate.CpuCapacity)*24*12),
		},
		{
			Kind:  "优化后",
			Value: fmt.Sprintf("%.2f", (rate.OptimizedRate.TotalRate.CpuCapacity)*24*12),
		},
		{
			Kind: "预计节省",
			Value: fmt.Sprintf("%.2f", (rate.OriginalRate.TotalRate.CpuCapacity-
				rate.OptimizedRate.TotalRate.CpuCapacity)*24*12),
		},
	}
	for _, n := range rate.OptimizedNodes {
		rt, ok := rate.OriginalRate.NodePackingRate[n]
		if !ok {
			continue
		}
		np, ok := rate.OriginalRate.NodePods[n]
		if !ok {
			continue
		}
		result.OptimizedNode = append(result.OptimizedNode, NodeInfo{
			Name:           n,
			CPUPackingRate: fmt.Sprintf("%.2f", rt.Cpu),
			MEMPackingRate: fmt.Sprintf("%.2f", rt.Mem),
			CPUCapacity:    fmt.Sprintf("%.2f", rt.CpuCapacity),
			MEMCapacity:    fmt.Sprintf("%.2f", rt.MemCapacity/1024/1024/1024),
			PodNum:         fmt.Sprintf("%d", len(np)),
		})
	}
	sort.Sort(NodeInfoList(result.OptimizedNode))
	r.GET("/api/info", func(c *gin.Context) {
		c.JSON(200, result)
	})
	return r.Run(fmt.Sprintf("0.0.0.0:%d", s.httpPort))
}
