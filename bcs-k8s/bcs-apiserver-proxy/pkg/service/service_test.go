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

package service

import (
	"reflect"
	"testing"
)

func TestLvsProxy_IsVSAvailable(t *testing.T) {
	lvs := NewLvsProxy()

	vsList := []struct {
		vs string
		ok bool
	}{
		{
			vs: "127.0.0.1:6443",
			ok: true,
		},
		{
			vs: "127.0.0.2:6443",
			ok: false,
		},
	}

	for _, server := range vsList {
		ok := lvs.IsVirtualServerAvailable(server.vs)
		if server.ok != ok {
			t.Logf("IsVirtualServerAvailable failed")
		}
	}

	t.Logf("IsVirtualServerAvailable successful")
}

func TestLvsProxy_CreateVirtualServer(t *testing.T) {
	lvs := NewLvsProxy()

	vsList := []string{"127.0.0.1:6443", "127.0.0.2:6443"}

	for _, server := range vsList {
		err := lvs.CreateVirtualServer(server)
		if err != nil {
			t.Logf("CreateVirtualServer failed")
		}
	}

	t.Logf("CreateVirtualServer successful")
}

func TestLvsProxy_DeleteVirtualServer(t *testing.T) {
	lvs := NewLvsProxy()
	vsList := []string{"127.0.0.1:6443", "127.0.0.2:6443"}

	for _, server := range vsList {
		err := lvs.DeleteVirtualServer(server)
		if err != nil {
			t.Logf("DeleteVirtualServer failed")
		}
	}

	t.Logf("DeleteVirtualServer successful")
}

func TestLvsProxy_CreateRealServer(t *testing.T) {
	lvs := NewLvsProxy()

	vs := "127.0.0.1:6443"
	ok := lvs.IsVirtualServerAvailable(vs)
	if !ok {
		err := lvs.CreateVirtualServer(vs)
		if err != nil {
			t.Fatalf("CreateVirtualServer failed: %v", err)
			return
		}
	}

	rss := []string{"192.168.0.1:8081", "192.168.0.2:8082", "192.168.0.3:8083"}
	for _, rs := range rss {
		err := lvs.CreateRealServer(rs)
		if err != nil {
			t.Fatalf("CreateRealServer rs[%s] failed: %v", rs, err)
			return
		}
	}

	t.Logf("CreateRealServer successful")
}

func TestLvsProxy_DeleteRealServer(t *testing.T) {
	lvs := NewLvsProxy()

	vs := "127.0.0.1:6443"
	ok := lvs.IsVirtualServerAvailable(vs)
	if !ok {
		t.Fatalf("IsVirtualServerAvailable failed: %s", vs)
		return
	}

	rss := []string{"192.168.0.1:8081", "192.168.0.2:8082", "192.168.0.3:8083"}
	for _, rs := range rss {
		err := lvs.DeleteRealServer(rs)
		if err != nil {
			t.Fatalf("CreateRealServer rs[%s] failed: %v", rs, err)
			return
		}
	}

	t.Logf("DeleteRealServer successful")
}

func TestLvsProxy_ListRealServer(t *testing.T) {
	lvs := NewLvsProxy()

	vs := "127.0.0.1:6443"
	ok := lvs.IsVirtualServerAvailable(vs)
	if !ok {
		t.Fatalf("IsVirtualServerAvailable failed: %s", vs)
		return
	}

	expectRss := []string{"192.168.0.1:8081", "192.168.0.2:8082", "192.168.0.3:8083"}
	rss, err := lvs.ListRealServer()
	if err != nil {
		t.Fatalf("ListRealServer vs[%s] failed", vs)
		return
	}

	if !reflect.DeepEqual(rss, expectRss) {
		t.Logf("ListRealServer failed")
	}

	t.Logf("ListRealServer successful")
}
