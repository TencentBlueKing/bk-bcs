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
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	grpcserver "github.com/go-micro/plugins/v4/server/grpc"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/server"
	"google.golang.org/grpc/test/bufconn"
)

type blockingRegistry struct {
	registry.Registry
	entered chan struct{}
	release chan struct{}
	err     error
}

func (r *blockingRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	close(r.entered)
	<-r.release
	if r.err != nil {
		return r.err
	}
	return r.Registry.Register(s, opts...)
}

func TestStopAndDeregisterWaitsInFlight(t *testing.T) {
	for _, registerErr := range []error{nil, errors.New("registration failed")} {
		name := "success"
		if registerErr != nil {
			name = "failure"
		}
		t.Run(name, func(t *testing.T) {
			backend := &blockingRegistry{
				Registry: registry.NewMemoryRegistry(),
				entered:  make(chan struct{}),
				release:  make(chan struct{}),
				err:      registerErr,
			}
			var releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(backend.release) }) }
			defer release()
			r := &shutdownRegistry{Registry: backend}
			svc := &registry.Service{
				Name: "shutdown-test", Version: "1",
				Nodes: []*registry.Node{{Id: "node-1", Address: "127.0.0.1:18081"}},
			}
			registered := make(chan error, 1)
			go func() { registered <- r.Register(svc) }()
			select {
			case <-backend.entered:
			case <-time.After(time.Second):
				t.Fatal("registration did not start")
			}

			stopping := make(chan struct{})
			stopped := make(chan error, 1)
			go func() {
				close(stopping)
				stopped <- r.StopAndDeregister()
			}()
			<-stopping
			select {
			case <-stopped:
				t.Fatal("shutdown did not wait for in-flight registration")
			case <-time.After(20 * time.Millisecond):
			}
			release()
			select {
			case err := <-registered:
				if !errors.Is(err, registerErr) {
					t.Fatalf("Register() = %v, want %v", err, registerErr)
				}
			case <-time.After(time.Second):
				t.Fatal("registration did not finish")
			}
			select {
			case err := <-stopped:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("shutdown did not finish")
			}

			assertServiceAbsent(t, backend, svc.Name)
			if err := r.StopAndDeregister(); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 3; i++ {
				if err := r.Register(svc); err != nil {
					t.Fatal(err)
				}
			}
			assertServiceAbsent(t, backend, svc.Name)
		})
	}
}

type testListener struct{ net.Listener }

func (testListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 18081}
}

func TestShutdownBlocksGRPCRenewal(t *testing.T) {
	backend := registry.NewMemoryRegistry()
	r := &shutdownRegistry{Registry: backend}
	s := grpcserver.NewServer(
		server.Name("shutdown-grpc-test"),
		server.Registry(r),
		server.RegisterTTL(30*time.Second),
		server.RegisterInterval(5*time.Millisecond),
		grpcserver.Listener(testListener{bufconn.Listen(1024)}),
	)
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Stop(); err != nil {
			t.Error(err)
		}
	})
	if services, err := backend.GetService("shutdown-grpc-test"); err != nil || len(services) == 0 {
		t.Fatalf("initial registration missing: %v", err)
	}
	if err := r.StopAndDeregister(); err != nil {
		t.Fatal(err)
	}
	assertServiceAbsent(t, backend, "shutdown-grpc-test")
	for i := 0; i < 32; i++ {
		time.Sleep(5 * time.Millisecond)
		assertServiceAbsent(t, backend, "shutdown-grpc-test")
	}
}

func assertServiceAbsent(t *testing.T, r registry.Registry, name string) {
	t.Helper()
	services, err := r.GetService(name)
	if err != nil && !errors.Is(err, registry.ErrNotFound) {
		t.Fatal(err)
	}
	if len(services) != 0 {
		t.Fatal("service reappeared after deregistration")
	}
}

func TestStopAndDeregisterBeforeStart(t *testing.T) {
	r := &shutdownRegistry{Registry: registry.NewMemoryRegistry()}
	if err := r.StopAndDeregister(); err != nil {
		t.Fatal(err)
	}
	svc := &registry.Service{Name: "never-started", Nodes: []*registry.Node{{Id: "self"}}}
	if err := r.Register(svc); err != nil {
		t.Fatal(err)
	}
	assertServiceAbsent(t, r, svc.Name)
}

type failingRegistry struct {
	registry.Registry
	registerErr     error
	deregisterErr   error
	deregistered    *registry.Service
	deregisterCalls int
}

func (r *failingRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	if r.registerErr != nil {
		return r.registerErr
	}
	return r.Registry.Register(s, opts...)
}

func (r *failingRegistry) Deregister(s *registry.Service, opts ...registry.DeregisterOption) error {
	r.deregisterCalls++
	r.deregistered = s
	if r.deregisterErr != nil {
		return r.deregisterErr
	}
	return r.Registry.Deregister(s, opts...)
}

func TestStopAndDeregisterKeepsIdentity(t *testing.T) {
	backend := &failingRegistry{Registry: registry.NewMemoryRegistry()}
	r := &shutdownRegistry{Registry: backend}
	svc := &registry.Service{
		Name: "identity-test", Version: "1",
		Nodes: []*registry.Node{{Id: "self", Address: "127.0.0.1:18081"}},
	}
	if err := r.Register(svc); err != nil {
		t.Fatal(err)
	}
	peer := &registry.Service{
		Name: svc.Name, Version: svc.Version,
		Nodes: []*registry.Node{{Id: "peer", Address: "127.0.0.1:18082"}},
	}
	if err := backend.Register(peer); err != nil {
		t.Fatal(err)
	}
	// Caller mutations and a failed renewal must not replace the last successful identity.
	svc.Name, svc.Version = "changed", "2"
	svc.Nodes[0].Id, svc.Nodes[0].Address = "changed", "127.0.0.1:18083"
	backend.registerErr = errors.New("renewal failed")
	if err := r.Register(svc); !errors.Is(err, backend.registerErr) {
		t.Fatalf("Register() = %v, want %v", err, backend.registerErr)
	}
	if err := r.StopAndDeregister(); err != nil {
		t.Fatal(err)
	}
	got := backend.deregistered
	if got.Name != "identity-test" || got.Version != "1" || len(got.Nodes) != 1 ||
		got.Nodes[0].Id != "self" || got.Nodes[0].Address != "127.0.0.1:18081" {
		t.Fatalf("unexpected deregistration identity: %+v", got)
	}
	services, err := backend.GetService(peer.Name)
	if err != nil || len(services) != 1 || len(services[0].Nodes) != 1 || services[0].Nodes[0].Id != "peer" {
		t.Fatalf("deregistration must remove only this instance: services=%+v, err=%v", services, err)
	}
}

func TestStopAndDeregisterFailureRetry(t *testing.T) {
	backend := &failingRegistry{Registry: registry.NewMemoryRegistry()}
	r := &shutdownRegistry{Registry: backend}
	svc := &registry.Service{Name: "retry-test", Nodes: []*registry.Node{{Id: "self"}}}
	if err := r.Register(svc); err != nil {
		t.Fatal(err)
	}
	backend.deregisterErr = errors.New("deregistration failed")
	if err := r.StopAndDeregister(); !errors.Is(err, backend.deregisterErr) {
		t.Fatalf("StopAndDeregister() = %v, want %v", err, backend.deregisterErr)
	}
	backend.registerErr = errors.New("must not renew during shutdown")
	if err := r.Register(svc); err != nil {
		t.Fatal(err)
	}
	backend.deregisterErr = nil
	if err := r.StopAndDeregister(); err != nil {
		t.Fatal(err)
	}
	assertServiceAbsent(t, backend, svc.Name)
	if err := r.StopAndDeregister(); err != nil {
		t.Fatal(err)
	}
	if backend.deregisterCalls != 2 {
		t.Fatalf("Deregister called %d times, want 2", backend.deregisterCalls)
	}
	// Framework cleanup can still deregister through the standard interface.
	if err := r.Deregister(svc); err != nil {
		t.Fatal(err)
	}
	assertServiceAbsent(t, backend, svc.Name)
}
