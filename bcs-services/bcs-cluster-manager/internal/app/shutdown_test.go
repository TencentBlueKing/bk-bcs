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

package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/http/ipv6server"
	commonutil "github.com/Tencent/bk-bcs/bcs-common/common/util"
	grpcserver "github.com/go-micro/plugins/v4/server/grpc"
	micro "go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"google.golang.org/grpc/test/bufconn"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
)

// Exercise the real micro.Run hooks. A stop notification must remove registration
// while HTTP remains usable, then wait for active HTTP requests before AfterStop.
func TestShutdownHooksDrainBeforeCleanup(t *testing.T) {
	cm := NewClusterManager(&options.ClusterManagerOptions{})
	defer cm.ctxCancelFunc()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cm.stopMicro = cancel
	backend := registry.NewMemoryRegistry()
	cm.serviceRegistry = &shutdownRegistry{Registry: backend}
	cm.initSignalHandler()

	entered, release, drained := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/slow" {
			close(entered)
			<-release
			defer close(drained)
		}
		select {
		case <-cm.ctx.Done():
			http.Error(w, "dependencies closed", http.StatusServiceUnavailable)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(ts.Close)
	t.Cleanup(unblock)
	cm.httpServer = &ipv6server.IPv6Server{Server: ts.Config}
	httpStopping := make(chan struct{})
	cm.httpServer.RegisterOnShutdown(func() { close(httpStopping) })
	started, cleaned := make(chan struct{}), make(chan struct{})
	svc := micro.NewService(
		micro.Cmd(commonutil.NewDummyMicroCmd()),
		micro.Server(grpcserver.NewServer(grpcserver.Listener(testListener{bufconn.Listen(1024)}))),
		micro.Registry(cm.serviceRegistry), micro.Name("shutdown-hooks-test"),
		micro.RegisterInterval(5*time.Millisecond), micro.Context(ctx), micro.HandleSignal(false),
		micro.AfterStart(func() error { close(started); return nil }),
		micro.BeforeStop(cm.beforeStop),
		micro.AfterStop(func() error {
			select {
			case <-drained:
			default:
				t.Error("cleanup ran before the active HTTP request drained")
			}
			cm.ctxCancelFunc()
			close(cleaned)
			return nil
		}),
	)
	svc.Init()
	cm.microService = svc
	finished := make(chan error, 1)
	go func() { finished <- svc.Run() }()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("micro service did not start")
	}
	request := func(path string) error {
		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Get(ts.URL + path)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP status %d", resp.StatusCode)
		}
		return nil
	}
	response := make(chan error, 1)
	go func() { response <- request("/slow") }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}
	cm.stopCh <- struct{}{}
	deadline := time.Now().Add(3 * time.Second)
	for {
		services, err := backend.GetService("shutdown-hooks-test")
		if err == registry.ErrNotFound || (err == nil && len(services) == 0) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("registration was not removed before the propagation wait")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err := request("/health"); err != nil {
		t.Fatalf("HTTP unavailable during route propagation: %v", err)
	}
	select {
	case <-httpStopping:
	case <-time.After(15 * time.Second):
		t.Fatal("HTTP drain did not start after notification receiver exited")
	}
	select {
	case <-cleaned:
		t.Fatal("cleanup skipped active request")
	default:
	}
	unblock()
	if err := <-response; err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("micro.Run did not finish shutdown")
	}
	assertServiceAbsent(t, backend, "shutdown-hooks-test")
}
