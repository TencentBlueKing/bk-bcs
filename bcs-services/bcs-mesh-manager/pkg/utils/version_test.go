package utils

import "testing"

func TestIsVersionSupported(t *testing.T) {
	tests := []struct {
		name           string
		clusterVersion string
		supportVersion string
		expect         bool
	}{
		{
			name:           "empty support version always true",
			clusterVersion: "1.20.0",
			supportVersion: "",
			expect:         true,
		},
		{
			name:           "exact match",
			clusterVersion: "1.20.0",
			supportVersion: "1.20.0",
			expect:         true,
		},
		{
			name:           "range satisfied",
			clusterVersion: "1.21.0",
			supportVersion: ">=1.20.0,<=1.22.0",
			expect:         true,
		},
		{
			name:           "range satisfied with xxx version",
			clusterVersion: "v1.20.6-xxx.27.1+xxx",
			supportVersion: ">=v1.20,<=v1.22.0",
			expect:         true,
		},
		{
			name:           "range satisfied with version",
			clusterVersion: "1.20.6-xxx.27.1+xxx",
			supportVersion: ">=v1.20,<=v1.22.0",
			expect:         true,
		},
		{
			name:           "range not satisfied",
			clusterVersion: "1.19.0",
			supportVersion: ">=1.20.0,<=1.22.0",
			expect:         false,
		},
		{
			name:           "invalid cluster version",
			clusterVersion: "not-a-version",
			supportVersion: ">=1.20.0",
			expect:         false,
		},
		{
			name:           "invalid support version",
			clusterVersion: "1.20.0",
			supportVersion: "not-a-constraint",
			expect:         false,
		},
		{
			name:           "empty cluster version",
			clusterVersion: "",
			supportVersion: "1.21",
			expect:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsVersionSupported(tt.clusterVersion, tt.supportVersion)
			if got != tt.expect {
				t.Errorf("IsVersionSupported(%q, %q) = %v, want %v", tt.clusterVersion, tt.supportVersion, got, tt.expect)
			}
		})
	}
}
