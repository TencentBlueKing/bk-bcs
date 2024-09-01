package etcd

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestHandleDelayedTask(t *testing.T) {
	endpoints := os.Getenv("ETCDCTL_ENDPOINTS")
	if endpoints == "" {
		t.Skip("ETCDCTL_ENDPOINTS is not set")
	}
	t.Parallel()

	ctx := context.Background()
	eta := time.Now().Add(time.Second * 10)
	mytask := &tasks.Signature{
		Name:       "test_delay_task",
		UUID:       "test-delay-0",
		RoutingKey: "test",
		ETA:        &eta,
	}
	broker, err := New(ctx, &config.Config{Broker: endpoints})
	etcdBroker := broker.(*etcdBroker)
	require.NoError(t, err)

	err = broker.Publish(ctx, mytask)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		err := etcdBroker.handleDelayedTask(ctx)
		assert.NoError(t, err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// 排队等待
		time.Sleep(time.Second)

		err := etcdBroker.handleDelayedTask(ctx)
		assert.NoError(t, err)
	}()

	wg.Wait()

	// 完成后，上面的锁需要立即释放
	err = etcdBroker.handleDelayedTask(ctx)
	assert.NoError(t, err)
}

func TestHandleDelayedMultiTask(t *testing.T) {
	endpoints := os.Getenv("ETCDCTL_ENDPOINTS")
	if endpoints == "" {
		t.Skip("ETCDCTL_ENDPOINTS is not set")
	}
	t.Parallel()

	ctx := context.Background()
	eta := time.Now().Add(time.Second * 10)
	mytask := &tasks.Signature{
		Name:       "test_delay_task",
		UUID:       "test-delay-00",
		RoutingKey: "test",
		ETA:        &eta,
	}
	broker, err := New(ctx, &config.Config{Broker: endpoints})
	etcdBroker := broker.(*etcdBroker)
	require.NoError(t, err)

	err = broker.Publish(ctx, mytask)
	require.NoError(t, err)

	eta = time.Now().Add(time.Second * 5)
	err = broker.Publish(ctx, &tasks.Signature{
		Name:       "test_delay_task1",
		UUID:       "test-delay-01",
		RoutingKey: "test",
		ETA:        &eta,
	})
	require.NoError(t, err)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		taskSlice, err := broker.GetDelayedTasks()
		assert.NoError(t, err)
		assert.True(t, len(taskSlice) >= 2)
		// 排队等待
		time.Sleep(time.Second * 10)

		err = etcdBroker.handleDelayedTask(ctx)
		assert.NoError(t, err)
	}()

	wg.Wait()
}

func TestListWatchPendingTask(t *testing.T) {
	endpoints := os.Getenv("ETCDCTL_ENDPOINTS")
	if endpoints == "" {
		t.Skip("ETCDCTL_ENDPOINTS is not set")
	}
	t.Parallel()

	st := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	mytask := &tasks.Signature{
		Name:       "test_task",
		UUID:       "test-0",
		RoutingKey: "test",
	}
	broker, err := New(ctx, &config.Config{Broker: endpoints})
	etcdBroker := broker.(*etcdBroker)
	require.NoError(t, err)

	err = broker.Publish(ctx, mytask)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := etcdBroker.listWatchPendingTask(ctx, "test")
		assert.NoError(t, err)
		_, ok := etcdBroker.pendingTask["test/test-0"]
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(etcdBroker.pendingTask), 1)
		duration := time.Since(st)
		assert.True(t, duration > time.Second*15, "lock duration %s should be greater than 15s", duration)
	}()
	wg.Wait()
}
