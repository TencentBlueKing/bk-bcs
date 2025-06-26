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

package utils

import (
	"testing"
)

func TestParseOpenTelemetryEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    string
		expectHost  string
		expectPort  int32
		expectPath  string
		expectError bool
	}{
		// 正常情况
		{
			name:        "完整的endpoint带路径",
			endpoint:    "bkm-collector.bkmonitor-operator.svc.cluster.local:443/v1/traces",
			expectHost:  "bkm-collector.bkmonitor-operator.svc.cluster.local",
			expectPort:  443,
			expectPath:  "/v1/traces",
			expectError: false,
		},
		{
			name:        "endpoint不带路径",
			endpoint:    "bkm-collector.bkmonitor-operator.svc.cluster.local:4318",
			expectHost:  "bkm-collector.bkmonitor-operator.svc.cluster.local",
			expectPort:  4318,
			expectPath:  "",
			expectError: false,
		},
		{
			name:        "带http协议前缀",
			endpoint:    "http://collector.example.com:8080/api/traces",
			expectHost:  "collector.example.com",
			expectPort:  8080,
			expectPath:  "/api/traces",
			expectError: false,
		},
		{
			name:        "带https协议前缀",
			endpoint:    "https://secure-collector.example.com:8443/v1/traces",
			expectHost:  "secure-collector.example.com",
			expectPort:  8443,
			expectPath:  "/v1/traces",
			expectError: false,
		},
		{
			name:        "localhost地址",
			endpoint:    "localhost:14268/api/traces",
			expectHost:  "localhost",
			expectPort:  14268,
			expectPath:  "/api/traces",
			expectError: false,
		},
		{
			name:        "IP地址",
			endpoint:    "192.168.1.100:9411/api/v2/spans",
			expectHost:  "192.168.1.100",
			expectPort:  9411,
			expectPath:  "/api/v2/spans",
			expectError: false,
		},
		{
			name:        "IPv6地址",
			endpoint:    "[::1]:8080/traces",
			expectHost:  "[::1]",
			expectPort:  8080,
			expectPath:  "/traces",
			expectError: false,
		},
		{
			name:        "复杂路径",
			endpoint:    "collector.example.com:8080/api/v1/traces?format=json",
			expectHost:  "collector.example.com",
			expectPort:  8080,
			expectPath:  "/api/v1/traces?format=json",
			expectError: false,
		},
		{
			name:        "端口1",
			endpoint:    "example.com:1/path",
			expectHost:  "example.com",
			expectPort:  1,
			expectPath:  "/path",
			expectError: false,
		},
		{
			name:        "端口65535",
			endpoint:    "example.com:65535",
			expectHost:  "example.com",
			expectPort:  65535,
			expectPath:  "",
			expectError: false,
		},

		// 错误情况
		{
			name:        "空字符串",
			endpoint:    "",
			expectError: true,
		},
		{
			name:        "缺少端口",
			endpoint:    "example.com",
			expectError: true,
		},
		{
			name:        "缺少主机名",
			endpoint:    ":8080/path",
			expectError: true,
		},
		{
			name:        "端口为0",
			endpoint:    "example.com:0",
			expectError: true,
		},
		{
			name:        "端口为负数",
			endpoint:    "example.com:-1",
			expectError: true,
		},
		{
			name:        "端口超出范围",
			endpoint:    "example.com:65536",
			expectError: true,
		},
		{
			name:        "端口不是数字",
			endpoint:    "example.com:abc",
			expectError: true,
		},
		{
			name:        "多个冒号",
			endpoint:    "example.com:8080:9090",
			expectError: true,
		},
		{
			name:        "只有协议前缀",
			endpoint:    "http://",
			expectError: true,
		},
		{
			name:        "只有协议前缀和主机",
			endpoint:    "https://example.com",
			expectError: true,
		},
		{
			name:        "端口为浮点数",
			endpoint:    "example.com:8080.5",
			expectError: true,
		},
		{
			name:        "端口包含字母",
			endpoint:    "example.com:80a0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, path, err := ParseOpenTelemetryEndpoint(tt.endpoint)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseOpenTelemetryEndpoint() expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseOpenTelemetryEndpoint() unexpected error: %v", err)
				return
			}

			if host != tt.expectHost {
				t.Errorf("ParseOpenTelemetryEndpoint() host = %v, want %v", host, tt.expectHost)
			}

			if port != tt.expectPort {
				t.Errorf("ParseOpenTelemetryEndpoint() port = %v, want %v", port, tt.expectPort)
			}

			if path != tt.expectPath {
				t.Errorf("ParseOpenTelemetryEndpoint() path = %v, want %v", path, tt.expectPath)
			}
		})
	}
}

// TestParseOpenTelemetryEndpoint_Examples 测试注释中的示例
func TestParseOpenTelemetryEndpoint_Examples(t *testing.T) {
	// 测试注释中的第一个示例
	host, port, path, err := ParseOpenTelemetryEndpoint("bkm-collector.bkmonitor-operator.svc.cluster.local:443/v1/traces")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if host != "bkm-collector.bkmonitor-operator.svc.cluster.local" {
		t.Errorf("host = %v, want bkm-collector.bkmonitor-operator.svc.cluster.local", host)
	}
	if port != 443 {
		t.Errorf("port = %v, want 443", port)
	}
	if path != "/v1/traces" {
		t.Errorf("path = %v, want /v1/traces", path)
	}

	// 测试注释中的第二个示例
	host, port, path, err = ParseOpenTelemetryEndpoint("bkm-collector.bkmonitor-operator.svc.cluster.local:4318")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if host != "bkm-collector.bkmonitor-operator.svc.cluster.local" {
		t.Errorf("host = %v, want bkm-collector.bkmonitor-operator.svc.cluster.local", host)
	}
	if port != 4318 {
		t.Errorf("port = %v, want 4318", port)
	}
	if path != "" {
		t.Errorf("path = %v, want empty string", path)
	}
}

// BenchmarkParseOpenTelemetryEndpoint 性能测试
func BenchmarkParseOpenTelemetryEndpoint(b *testing.B) {
	endpoint := "bkm-collector.bkmonitor-operator.svc.cluster.local:443/v1/traces"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = ParseOpenTelemetryEndpoint(endpoint)
	}
}
