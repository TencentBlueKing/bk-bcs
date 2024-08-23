package task

import (
	"fmt"
	"testing"
)

func TestRetryIn(t *testing.T) {
	tests := []struct {
		count    int
		expected int
	}{
		{-1, 1},
		{0, 1},
		{1, 2},
		{2, 3},
		{3, 5},
		{4, 8},
		{5, 13},
		{6, 21},
		{7, 34},
		{8, 55},
		{9, 89},
		{10, 144},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("count=%d", tt.count), func(t *testing.T) {
			result := retryNext(tt.count)
			if result != tt.expected {
				t.Errorf("retryNext(%d) = %d; want %d", tt.count, result, tt.expected)
			}
		})
	}
}
