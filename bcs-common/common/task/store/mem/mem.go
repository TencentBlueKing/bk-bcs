package mem

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/store/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

type memStore struct {
	mtx   sync.Mutex
	tasks map[string]*types.Task
}

func New() iface.Store {
	s := &memStore{
		tasks: make(map[string]*types.Task),
	}
	return s
}

// EnsureTable 创建db表
func (s *memStore) EnsureTable(ctx context.Context, dst ...any) error {
	return nil
}

func (s *memStore) CreateTask(ctx context.Context, task *types.Task) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.tasks[task.GetTaskID()] = task
	return nil
}

func (s *memStore) ListTask(ctx context.Context, opt *iface.ListOption) ([]types.Task, error) {
	return nil, nil
}

func (s *memStore) UpdateTask(ctx context.Context, task *types.Task) error {
	return nil
}

func (s *memStore) DeleteTask(ctx context.Context, taskID string) error {
	return nil
}

func (s *memStore) GetTask(ctx context.Context, taskID string) (*types.Task, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	t, ok := s.tasks[taskID]
	if ok {
		return t, nil
	}
	return nil, fmt.Errorf("not found")
}

func (s *memStore) PatchTask(ctx context.Context, taskID string, patchs map[string]interface{}) error {
	return nil
}
