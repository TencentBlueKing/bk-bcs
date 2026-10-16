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
	"net"
	"net/url"
	"testing"
)

func TestBuildAPIServerAddress(t *testing.T) {
	tests := []struct {
		name, host, port, want string
	}{
		{"IPv4", "10.96.0.1", "443", "https://10.96.0.1:443"},
		{"IPv4 custom port", "192.168.1.10", "6443", "https://192.168.1.10:6443"},
		{"compressed IPv6", "fd11:1111:1111:15::1", "443", "https://[fd11:1111:1111:15::1]:443"},
		{"full IPv6", "fd11:1111:1111:0015:0000:0000:0000:0001", "6443", "https://[fd11:1111:1111:0015:0000:0000:0000:0001]:6443"},
		{"IPv6 loopback", "::1", "443", "https://[::1]:443"},
		{"hostname", "kubernetes.default.svc", "443", "https://kubernetes.default.svc:443"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildAPIServerAddress(tt.host, tt.port)
			if got != tt.want {
				t.Fatalf("address = %q, want %q", got, tt.want)
			}
			u, err := url.Parse(got)
			if err != nil {
				t.Fatal(err)
			}
			host, port, err := net.SplitHostPort(u.Host)
			if err != nil {
				t.Fatalf("registered address cannot be dialed: %v", err)
			}
			if u.Scheme != "https" || host != tt.host || port != tt.port {
				t.Fatalf("parsed address changed: scheme=%q host=%q port=%q", u.Scheme, host, port)
			}
		})
	}
}
