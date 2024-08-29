package etcd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindTaskKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		key   string
	}{
		{
			name:  "running_task1",
			input: "/machinery/v2/broker/running_tasks/machinery_tasks/d30986b4-6634-4013-bf56-88c0463450c2-test-0",
			key:   "machinery_tasks/d30986b4-6634-4013-bf56-88c0463450c2-test-0",
		},
		{
			name:  "pending_task1",
			input: "/machinery/v2/broker/pending_tasks/machinery_tasks/d30986b4-6634-4013-bf56-88c0463450c2-test-0",
			key:   "machinery_tasks/d30986b4-6634-4013-bf56-88c0463450c2-test-0",
		},
		{
			name:  "delayed_task1",
			input: "/machinery/v2/broker/delayed_tasks/eta-0/machinery_tasks/d30986b4-6634-4013-bf56-88c0463450c2-test-0",
			key:   "machinery_tasks/d30986b4-6634-4013-bf56-88c0463450c2-test-0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := findTaskKey(tt.input)
			assert.Equal(t, tt.key, k)
		})
	}
}
