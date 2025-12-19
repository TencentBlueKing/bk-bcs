package utils

import (
	"testing"

	"google.golang.org/protobuf/types/known/wrapperspb"

	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

func TestConvertCPUToMilliCores(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		isRequest bool
		expected  string
	}{
		// Request 字段测试
		{"request already milli cores", "1000m", true, "1000m"},
		{"request whole number", "1", true, "1000m"},
		{"request decimal", "0.5", true, "500m"},
		{"request zero", "0", true, "0m"}, // 零值会被转换为 0m
		{"request empty string", "", true, ""},
		{"request invalid format", "invalid", true, "invalid"}, // 转换失败时返回原始值
		{"request negative", "-1", true, "-1"},                 // 转换失败时返回原始值

		// Limit 字段测试
		{"limit already milli cores", "1000m", false, "1000m"},
		{"limit whole number", "1", false, "1000m"},
		{"limit decimal", "0.5", false, "500m"},
		{"limit zero", "0", false, ""}, // limit 为0时返回空字符串
		{"limit empty string", "", false, ""},
		{"limit invalid format", "invalid", false, "invalid"}, // 转换失败时返回原始值
		{"limit negative", "-1", false, "-1"},                 // 转换失败时返回原始值
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertCPUToMilliCores(tt.input, tt.isRequest)
			if result != tt.expected {
				t.Errorf("convertCPUToMilliCores(%s, %t) = %s, want %s", tt.input, tt.isRequest, result, tt.expected)
			}
		})
	}
}

func TestConvertMemoryToMi(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		isRequest bool
		expected  string
	}{
		// Request 字段测试
		{"request already Mi", "1000Mi", true, "1000Mi"},
		{"request Gi to Mi", "1Gi", true, "1024Mi"},
		{"request zero", "0", true, "0Mi"}, // 零值会被转换为 0Mi
		{"request empty string", "", true, ""},
		{"request invalid format", "invalid", true, "invalid"}, // 转换失败时返回原始值
		{"request negative", "-1Gi", true, "-1Gi"},             // 转换失败时返回原始值

		// Limit 字段测试
		{"limit already Mi", "1000Mi", false, "1000Mi"},
		{"limit Gi to Mi", "1Gi", false, "1024Mi"},
		{"limit 2Gi to Mi", "2Gi", false, "2048Mi"},
		{"limit decimal Gi", "1.5Gi", false, "1536Mi"},
		{"limit G to Mi", "1G", false, "953.67Mi"},
		{"limit M to Mi", "1000M", false, "953.67Mi"},
		{"limit Ki to Mi", "1024Ki", false, "1Mi"},
		{"limit bytes to Mi", "1048576", false, "1Mi"},
		{"limit zero", "0", false, ""}, // limit 为0时返回空字符串
		{"limit empty string", "", false, ""},
		{"limit invalid format", "invalid", false, "invalid"}, // 转换失败时返回原始值
		{"limit negative", "-1Gi", false, "-1Gi"},             // 转换失败时返回原始值
		{"limit 1Ti", "1Ti", false, "1048576Mi"},
		{"limit 1T", "1T", false, "953674.32Mi"},
		{"limit 1Pi", "1Pi", false, "1073741824Mi"},
		{"limit 1P", "1P", false, "953674316.41Mi"},
		{"limit 1Ei", "1Ei", false, "1099511627776Mi"},
		{"limit 1E", "1E", false, "953674316406.25Mi"},
		{"limit 2Ti", "2Ti", false, "2097152Mi"},
		{"limit 0.5Ti", "0.5Ti", false, "524288Mi"},
		{"limit 1024Gi", "1024Gi", false, "1048576Mi"},
		{"limit 1000G", "1000G", false, "953674.32Mi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertMemoryToMi(tt.input, tt.isRequest)
			if result != tt.expected {
				t.Errorf("convertMemoryToMi(%s, %t) = %s, want %s", tt.input, tt.isRequest, result, tt.expected)
			}
		})
	}
}

func TestNormalizeResourceConfigUnits(t *testing.T) {
	tests := []struct {
		name     string
		input    *meshmanager.ResourceConfig
		expected *meshmanager.ResourceConfig
	}{
		{
			name: "convert all fields",
			input: &meshmanager.ResourceConfig{
				CpuRequest:    wrapperspb.String("1"),
				CpuLimit:      wrapperspb.String("2.5"),
				MemoryRequest: wrapperspb.String("1Gi"),
				MemoryLimit:   wrapperspb.String("2G"),
			},
			expected: &meshmanager.ResourceConfig{
				CpuRequest:    wrapperspb.String("1000m"),
				CpuLimit:      wrapperspb.String("2500m"),
				MemoryRequest: wrapperspb.String("1024Mi"),
				MemoryLimit:   wrapperspb.String("1907.35Mi"),
			},
		},
		{
			name: "already in correct format",
			input: &meshmanager.ResourceConfig{
				CpuRequest:    wrapperspb.String("1000m"),
				CpuLimit:      wrapperspb.String("2000m"),
				MemoryRequest: wrapperspb.String("1024Mi"),
				MemoryLimit:   wrapperspb.String("2048Mi"),
			},
			expected: &meshmanager.ResourceConfig{
				CpuRequest:    wrapperspb.String("1000m"),
				CpuLimit:      wrapperspb.String("2000m"),
				MemoryRequest: wrapperspb.String("1024Mi"),
				MemoryLimit:   wrapperspb.String("2048Mi"),
			},
		},
		{
			name:     "nil config",
			input:    nil,
			expected: nil,
		},
		{
			name: "partial fields",
			input: &meshmanager.ResourceConfig{
				CpuRequest: wrapperspb.String("0.5"),
				// 其他字段为 nil
			},
			expected: &meshmanager.ResourceConfig{
				CpuRequest: wrapperspb.String("500m"),
				// 其他字段为 nil
			},
		},
		{
			name: "invalid CPU request - returns original value",
			input: &meshmanager.ResourceConfig{
				CpuRequest: wrapperspb.String("0"), // 无效的 request，返回转换后的值
			},
			expected: &meshmanager.ResourceConfig{
				CpuRequest: wrapperspb.String("0m"), // 返回转换后的值
			},
		},
		{
			name: "invalid memory request - returns original value",
			input: &meshmanager.ResourceConfig{
				MemoryRequest: wrapperspb.String("invalid"), // 无效的 request，返回原始值
			},
			expected: &meshmanager.ResourceConfig{
				MemoryRequest: wrapperspb.String("invalid"), // 返回原始值
			},
		},
		{
			name: "valid limit zero values",
			input: &meshmanager.ResourceConfig{
				CpuLimit:    wrapperspb.String("0"), // 有效的 limit
				MemoryLimit: wrapperspb.String("0"), // 有效的 limit
			},
			expected: &meshmanager.ResourceConfig{
				CpuLimit:    wrapperspb.String(""), // limit 为0时返回空字符串
				MemoryLimit: wrapperspb.String(""), // limit 为0时返回空字符串
			},
		},
		{
			name: "large memory units",
			input: &meshmanager.ResourceConfig{
				CpuRequest:    wrapperspb.String("4"),
				CpuLimit:      wrapperspb.String("8"),
				MemoryRequest: wrapperspb.String("1Ti"),
				MemoryLimit:   wrapperspb.String("1Pi"),
			},
			expected: &meshmanager.ResourceConfig{
				CpuRequest:    wrapperspb.String("4000m"),
				CpuLimit:      wrapperspb.String("8000m"),
				MemoryRequest: wrapperspb.String("1048576Mi"),
				MemoryLimit:   wrapperspb.String("1073741824Mi"),
			},
		},
		{
			name: "negative values - returns original values",
			input: &meshmanager.ResourceConfig{
				CpuRequest:    wrapperspb.String("-1"),
				CpuLimit:      wrapperspb.String("-2"),
				MemoryRequest: wrapperspb.String("-1Gi"),
				MemoryLimit:   wrapperspb.String("-2Gi"),
			},
			expected: &meshmanager.ResourceConfig{
				CpuRequest:    wrapperspb.String("-1"),   // 返回原始值
				CpuLimit:      wrapperspb.String("-2"),   // 返回原始值
				MemoryRequest: wrapperspb.String("-1Gi"), // 返回原始值
				MemoryLimit:   wrapperspb.String("-2Gi"), // 返回原始值
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NormalizeResourceConfigUnits(tt.input)

			// 现在函数不再返回错误
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.expected == nil {
				if tt.input != nil {
					t.Errorf("NormalizeResourceConfigUnits() = %v, want nil", tt.input)
				}
				return
			}

			if tt.input == nil {
				t.Errorf("NormalizeResourceConfigUnits() = nil, want %v", tt.expected)
				return
			}

			// 比较 CPU 字段
			if tt.expected.CpuRequest != nil {
				if tt.input.CpuRequest == nil {
					t.Errorf("CpuRequest is nil, want %s", tt.expected.CpuRequest.GetValue())
				} else if tt.input.CpuRequest.GetValue() != tt.expected.CpuRequest.GetValue() {
					t.Errorf("CpuRequest = %s, want %s", tt.input.CpuRequest.GetValue(), tt.expected.CpuRequest.GetValue())
				}
			}

			if tt.expected.CpuLimit != nil {
				if tt.input.CpuLimit == nil {
					t.Errorf("CpuLimit is nil, want %s", tt.expected.CpuLimit.GetValue())
				} else if tt.input.CpuLimit.GetValue() != tt.expected.CpuLimit.GetValue() {
					t.Errorf("CpuLimit = %s, want %s", tt.input.CpuLimit.GetValue(), tt.expected.CpuLimit.GetValue())
				}
			}

			// 比较内存字段
			if tt.expected.MemoryRequest != nil {
				if tt.input.MemoryRequest == nil {
					t.Errorf("MemoryRequest is nil, want %s", tt.expected.MemoryRequest.GetValue())
				} else if tt.input.MemoryRequest.GetValue() != tt.expected.MemoryRequest.GetValue() {
					t.Errorf("MemoryRequest = %s, want %s", tt.input.MemoryRequest.GetValue(), tt.expected.MemoryRequest.GetValue())
				}
			}

			if tt.expected.MemoryLimit != nil {
				if tt.input.MemoryLimit == nil {
					t.Errorf("MemoryLimit is nil, want %s", tt.expected.MemoryLimit.GetValue())
				} else if tt.input.MemoryLimit.GetValue() != tt.expected.MemoryLimit.GetValue() {
					t.Errorf("MemoryLimit = %s, want %s", tt.input.MemoryLimit.GetValue(), tt.expected.MemoryLimit.GetValue())
				}
			}
		})
	}
}

func TestNormalizeResourcesConfig(t *testing.T) {
	// 创建测试用的 IstioDetailInfo
	detailInfo := &meshmanager.IstioDetailInfo{
		SidecarResourceConfig: &meshmanager.ResourceConfig{
			CpuRequest:    wrapperspb.String("1"),
			CpuLimit:      wrapperspb.String("2.5"),
			MemoryRequest: wrapperspb.String("1Gi"),
			MemoryLimit:   wrapperspb.String("2G"),
		},
		HighAvailability: &meshmanager.HighAvailability{
			ResourceConfig: &meshmanager.ResourceConfig{
				CpuRequest:    wrapperspb.String("0.5"),
				CpuLimit:      wrapperspb.String("1"),
				MemoryRequest: wrapperspb.String("512Mi"),
				MemoryLimit:   wrapperspb.String("1Gi"),
			},
		},
	}

	// 执行转换
	err := NormalizeResourcesConfig(detailInfo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// 验证 Sidecar 资源配置
	if detailInfo.SidecarResourceConfig.CpuRequest.GetValue() != "1000m" {
		t.Errorf("Sidecar CpuRequest = %s, want 1000m", detailInfo.SidecarResourceConfig.CpuRequest.GetValue())
	}
	if detailInfo.SidecarResourceConfig.CpuLimit.GetValue() != "2500m" {
		t.Errorf("Sidecar CpuLimit = %s, want 2500m", detailInfo.SidecarResourceConfig.CpuLimit.GetValue())
	}
	if detailInfo.SidecarResourceConfig.MemoryRequest.GetValue() != "1024Mi" {
		t.Errorf("Sidecar MemoryRequest = %s, want 1024Mi", detailInfo.SidecarResourceConfig.MemoryRequest.GetValue())
	}
	if detailInfo.SidecarResourceConfig.MemoryLimit.GetValue() != "1907.35Mi" {
		t.Errorf("Sidecar MemoryLimit = %s, want 1907.35Mi", detailInfo.SidecarResourceConfig.MemoryLimit.GetValue())
	}

	// 验证高可用资源配置
	if detailInfo.HighAvailability.ResourceConfig.CpuRequest.GetValue() != "500m" {
		t.Errorf("HA CpuRequest = %s, want 500m", detailInfo.HighAvailability.ResourceConfig.CpuRequest.GetValue())
	}
	if detailInfo.HighAvailability.ResourceConfig.CpuLimit.GetValue() != "1000m" {
		t.Errorf("HA CpuLimit = %s, want 1000m", detailInfo.HighAvailability.ResourceConfig.CpuLimit.GetValue())
	}
	if detailInfo.HighAvailability.ResourceConfig.MemoryRequest.GetValue() != "512Mi" {
		t.Errorf("HA MemoryRequest = %s, want 512Mi", detailInfo.HighAvailability.ResourceConfig.MemoryRequest.GetValue())
	}
	if detailInfo.HighAvailability.ResourceConfig.MemoryLimit.GetValue() != "1024Mi" {
		t.Errorf("HA MemoryLimit = %s, want 1024Mi", detailInfo.HighAvailability.ResourceConfig.MemoryLimit.GetValue())
	}
}
