package utils

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestResourceParseQuantity(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedString string
		milliValue     int64 // 用于CPU
		value          int64 // 用于内存
		isCPU          bool
	}{
		// CPU相关
		{"1000m CPU", "1000m", "1", 1000, 1, true},
		{"500m CPU", "500m", "500m", 500, 0, true},
		{"1 CPU", "1", "1", 1000, 1, true},
		{"0.5 CPU", "0.5", "500m", 500, 0, true},
		{"2.5 CPU", "2.5", "2500m", 2500, 0, true},
		{"0 CPU", "0", "0", 0, 0, true},
		{"0m CPU", "0m", "0", 0, 0, true},
		// 内存相关
		{"100Mi Memory", "100Mi", "100Mi", 0, 104857600, false},
		{"1Gi Memory", "1Gi", "1Gi", 0, 1073741824, false},
		{"1G Memory", "1G", "1G", 0, 1000000000, false},
		{"100M Memory", "100M", "100M", 0, 100000000, false},
		{"1Ki Memory", "1Ki", "1Ki", 0, 1024, false},
		{"0 Memory", "0", "0", 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quantity, err := resource.ParseQuantity(tt.input)
			if err != nil {
				t.Fatalf("ParseQuantity(%s) failed: %v", tt.input, err)
			}

			if quantity.String() != tt.expectedString {
				t.Errorf("ParseQuantity(%s).String() = %v, want %v", tt.input, quantity.String(), tt.expectedString)
			}

			if tt.isCPU {
				if quantity.MilliValue() != tt.milliValue {
					t.Errorf("ParseQuantity(%s).MilliValue() = %v, want %v", tt.input, quantity.MilliValue(), tt.milliValue)
				}
			} else {
				if quantity.Value() != tt.value {
					t.Errorf("ParseQuantity(%s).Value() = %v, want %v", tt.input, quantity.Value(), tt.value)
				}
			}

			t.Logf("Input: %s", tt.input)
			t.Logf("  String(): %s", quantity.String())
			t.Logf("  Value(): %d", quantity.Value())
			t.Logf("  MilliValue(): %d", quantity.MilliValue())
			t.Logf("  IsZero(): %v", quantity.IsZero())
		})
	}
}

func TestResourceParseQuantityCPU(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedMilliValue int64
		expectedString     string
	}{
		{"1000m", "1000m", 1000, "1"},
		{"500m", "500m", 500, "500m"},
		{"1", "1", 1000, "1"},
		{"0.5", "0.5", 500, "500m"},
		{"2.5", "2.5", 2500, "2500m"},
		{"0.1", "0.1", 100, "100m"},
		{"0.01", "0.01", 10, "10m"},
		{"0.001", "0.001", 1, "1m"},
		{"0", "0", 0, "0"},
		{"0m", "0m", 0, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quantity, err := resource.ParseQuantity(tt.input)
			if err != nil {
				t.Fatalf("ParseQuantity(%s) failed: %v", tt.input, err)
			}

			// 检查毫核值
			if quantity.MilliValue() != tt.expectedMilliValue {
				t.Errorf("ParseQuantity(%s).MilliValue() = %d, want %d",
					tt.input, quantity.MilliValue(), tt.expectedMilliValue)
			}

			// 检查字符串表示
			if quantity.String() != tt.expectedString {
				t.Errorf("ParseQuantity(%s).String() = %s, want %s",
					tt.input, quantity.String(), tt.expectedString)
			}

			t.Logf("CPU Input: %s -> MilliValue: %d, String: %s",
				tt.input, quantity.MilliValue(), quantity.String())
		})
	}
}

func TestResourceParseQuantityMemory(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedBytes  int64
		expectedString string
	}{
		{"100Mi", "100Mi", 104857600, "100Mi"},
		{"1Gi", "1Gi", 1073741824, "1Gi"},
		{"1G", "1G", 1000000000, "1G"},
		{"100M", "100M", 100000000, "100M"},
		{"1Ki", "1Ki", 1024, "1Ki"},
		{"0", "0", 0, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quantity, err := resource.ParseQuantity(tt.input)
			if err != nil {
				t.Fatalf("ParseQuantity(%s) failed: %v", tt.input, err)
			}

			// 检查字节值
			if quantity.Value() != tt.expectedBytes {
				t.Errorf("ParseQuantity(%s).Value() = %d, want %d",
					tt.input, quantity.Value(), tt.expectedBytes)
			}

			// 检查字符串表示
			if quantity.String() != tt.expectedString {
				t.Errorf("ParseQuantity(%s).String() = %s, want %s",
					tt.input, quantity.String(), tt.expectedString)
			}

			t.Logf("Memory Input: %s -> Bytes: %d, String: %s",
				tt.input, quantity.Value(), quantity.String())
		})
	}
}

func TestResourceParseQuantityMemory_UnitConversion(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedBytes  int64
		expectedString string
	}{
		{"1024Mi to Gi", "1024Mi", 1073741824, "1Gi"},
		{"2048Mi to Gi", "2048Mi", 2147483648, "2Gi"},
		{"1Gi to Mi", "1Gi", 1073741824, "1Gi"},
		{"2Gi to Mi", "2Gi", 2147483648, "2Gi"},
		{"1000M to G", "1000M", 1000000000, "1G"},
		{"2000M to G", "2000M", 2000000000, "2G"},
		{"1G to M", "1G", 1000000000, "1G"},
		{"2G to M", "2G", 2000000000, "2G"},
		{"1536Mi to Gi", "1536Mi", 1610612736, "1536Mi"}, // 不足2Gi不会自动进位
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quantity, err := resource.ParseQuantity(tt.input)
			if err != nil {
				t.Fatalf("ParseQuantity(%s) failed: %v", tt.input, err)
			}

			if quantity.Value() != tt.expectedBytes {
				t.Errorf("ParseQuantity(%s).Value() = %d, want %d", tt.input, quantity.Value(), tt.expectedBytes)
			}

			if quantity.String() != tt.expectedString {
				t.Errorf("ParseQuantity(%s).String() = %s, want %s", tt.input, quantity.String(), tt.expectedString)
			}

			t.Logf("Input: %s -> Bytes: %d, String: %s", tt.input, quantity.Value(), quantity.String())
		})
	}
}

func TestResourceParseQuantityEmptyString(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedError  bool
		expectedString string
		expectedValue  int64
		expectedMilli  int64
		expectedIsZero bool
	}{
		{"empty string", "", true, "", 0, 0, true},
		{"whitespace only", "   ", true, "", 0, 0, true},
		{"tab only", "\t", true, "", 0, 0, true},
		{"newline only", "\n", true, "", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quantity, err := resource.ParseQuantity(tt.input)

			if tt.expectedError {
				if err == nil {
					t.Errorf("ParseQuantity(%q) should have failed, but got no error", tt.input)
				}
				t.Logf("ParseQuantity(%q) correctly failed with error: %v", tt.input, err)
				return
			}

			if err != nil {
				t.Fatalf("ParseQuantity(%q) failed unexpectedly: %v", tt.input, err)
			}

			// 检查字符串表示
			if quantity.String() != tt.expectedString {
				t.Errorf("ParseQuantity(%q).String() = %q, want %q",
					tt.input, quantity.String(), tt.expectedString)
			}

			// 检查字节值
			if quantity.Value() != tt.expectedValue {
				t.Errorf("ParseQuantity(%q).Value() = %d, want %d",
					tt.input, quantity.Value(), tt.expectedValue)
			}

			// 检查毫核值
			if quantity.MilliValue() != tt.expectedMilli {
				t.Errorf("ParseQuantity(%q).MilliValue() = %d, want %d",
					tt.input, quantity.MilliValue(), tt.expectedMilli)
			}

			// 检查是否为零
			if quantity.IsZero() != tt.expectedIsZero {
				t.Errorf("ParseQuantity(%q).IsZero() = %v, want %v",
					tt.input, quantity.IsZero(), tt.expectedIsZero)
			}

			t.Logf("Input: %q", tt.input)
			t.Logf("  String(): %q", quantity.String())
			t.Logf("  Value(): %d", quantity.Value())
			t.Logf("  MilliValue(): %d", quantity.MilliValue())
			t.Logf("  IsZero(): %v", quantity.IsZero())
		})
	}
}
