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

// Package nodecidr provides node CIDR related operations
package nodecidr

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	nodeName   string
	kubeconfig string
)

// NewNodeCidrCommand creates and returns the node-cidr cobra command
func NewNodeCidrCommand(scheme *runtime.Scheme) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node-cidr",
		Short: "Node CIDR operations",
		Long:  "Node CIDR related operations",
		Run: func(cmd *cobra.Command, args []string) {
			runNodeCidr(cmd, args, scheme)
		},
	}

	cmd.Flags().StringVar(&nodeName, "node-name", "", "Node name to get pod CIDR")
	_ = cmd.MarkFlagRequired("node-name")
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", "",
		"Path to kubeconfig file (default: use default kubeconfig or in-cluster config)")

	return cmd
}

// nolint
// runNodeCidr 获取指定节点的 pod CIDR 并输出
func runNodeCidr(cmd *cobra.Command, args []string, scheme *runtime.Scheme) {
	if nodeName == "" {
		fmt.Fprintf(os.Stderr, "Error: --node-name is required\n")
		os.Exit(1)
	}

	// 创建 k8s client
	var config *rest.Config
	var err error
	if kubeconfig != "" {
		// 使用指定的 kubeconfig 文件
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to build k8s config from kubeconfig file %s: %v\n", kubeconfig, err)
			os.Exit(1)
		}
	} else {
		// 使用默认的 kubeconfig 或 in-cluster config
		config, err = ctrl.GetConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get k8s config: %v\n", err)
			os.Exit(1)
		}
	}

	cl, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create k8s client: %v\n", err)
		os.Exit(1)
	}

	// 获取 node
	node := &corev1.Node{}
	ctx := context.Background()
	err = cl.Get(ctx, types.NamespacedName{Name: nodeName}, node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get node %s: %v\n", nodeName, err)
		os.Exit(1)
	}

	// 获取 pod CIDR
	// 优先使用 spec.podCIDR（单数，IPv4），如果没有则使用 spec.podCIDRs[0]（复数，支持多栈）
	var podCIDR string
	if node.Spec.PodCIDR != "" {
		podCIDR = node.Spec.PodCIDR
	} else if len(node.Spec.PodCIDRs) > 0 {
		podCIDR = node.Spec.PodCIDRs[0]
	} else {
		fmt.Fprintf(os.Stderr, "Error: node %s has no pod CIDR configured\n", nodeName)
		os.Exit(1)
	}

	// 直接输出字符串
	fmt.Print(podCIDR)
}
