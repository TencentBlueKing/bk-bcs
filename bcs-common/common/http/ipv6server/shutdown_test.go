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

package ipv6server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

// shutdownAddresses 多个回环地址，保证有多个 listener 同时上报。端口用 "0"
// 让每个地址各取一个空闲端口，避免与并发测试抢占固定端口。
var shutdownAddresses = []string{IPv4LoopBack, "127.0.0.2", "127.0.0.3", IPv6LoopBack}

// Shutdown 会让全部 listener 同时返回 ErrServerClosed，而调用方通常在这之后
// 还要继续排空依赖。晚于 ListenAndServe 返回的上报不能让进程崩溃。
func TestShutdownReleasesEveryListener(t *testing.T) {
	certFile, keyFile := getTLSFiles()
	t.Cleanup(func() {
		_ = os.Remove(certFile)
		_ = os.Remove(keyFile)
	})

	serves := map[string]func(*IPv6Server) error{
		"ListenAndServe":    func(s *IPv6Server) error { return s.ListenAndServe() },
		"ListenAndServeTLS": func(s *IPv6Server) error { return s.ListenAndServeTLS(certFile, keyFile) },
	}
	for name, serve := range serves {
		serve := serve
		t.Run(name, func(t *testing.T) {
			srv := NewIPv6Server(shutdownAddresses, "0", types.TCP, http.NewServeMux())
			if listeners, err := srv.Listen(); err != nil || len(listeners) < 2 {
				for _, listener := range listeners {
					_ = listener.Close()
				}
				t.Skipf("need at least two loopback listeners to cover the race, got %d (%v)",
					len(listeners), err)
			} else {
				for _, listener := range listeners {
					_ = listener.Close()
				}
			}

			served := make(chan error, 1)
			go func() { served <- serve(srv) }()
			time.Sleep(200 * time.Millisecond)

			if err := srv.Shutdown(context.Background()); err != nil {
				t.Fatalf("Shutdown() = %v", err)
			}
			select {
			case err := <-served:
				if !errors.Is(err, http.ErrServerClosed) {
					t.Fatalf("serve returned %v, want %v", err, http.ErrServerClosed)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("serve did not return after Shutdown")
			}

			// 剩余 listener 在 serve 返回之后才上报。若上报通道已被关闭，
			// 这里会 panic 并直接终止整个测试进程。
			time.Sleep(500 * time.Millisecond)
		})
	}
}
