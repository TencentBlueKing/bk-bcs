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

// Package app xxx
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

// Run run agent
func Run() error {
	kubeconfig := viper.GetString("agent.kubeconfig")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kubeCtx, err := NewKubeClientContext(kubeconfig)
	if err != nil {
		return fmt.Errorf("kubeAgent init kubeClient context failed: %s", err.Error())
	}
	go kubeCtx.Run(ctx)

	useWebsocket := viper.GetBool("agent.use-websocket")
	if useWebsocket {
		err := buildWebsocketToBke(kubeCtx.GetRestConfig())
		if err != nil {
			return err
		}
	} else {
		go reportToBke(kubeCtx)
	}

	// to run in the container, should not trap SIGTERM
	interrupt := make(chan os.Signal, 10)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case e := <-interrupt:
			blog.Infof("receive interrupt %s, do close", e.String())
			cancel()
		default:
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	listenAddr := viper.GetString("agent.listenAddr")
	return http.ListenAndServe(listenAddr, nil)
}
