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

// Package etcd implement machinery v2 backend iface
package etcd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/RichardKnop/machinery/v2/backends/iface"
	"github.com/RichardKnop/machinery/v2/common"
	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/log"
	"github.com/RichardKnop/machinery/v2/tasks"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
)

const (
	groupKey = "/machinery/v2/backend/groups/%s"
	taskKey  = "/machinery/v2/backend/tasks/%s"
)

type etcdBackend struct {
	common.Backend
	ctx    context.Context
	client *clientv3.Client
}

// New ..
func New(ctx context.Context, conf *config.Config) (iface.Backend, error) {
	etcdConf := clientv3.Config{
		Endpoints:   []string{conf.ResultBackend},
		Context:     ctx,
		DialTimeout: time.Second * 5,
		TLS:         conf.TLSConfig,
	}
	client, err := clientv3.New(etcdConf)
	if err != nil {
		return nil, err
	}

	backend := etcdBackend{
		Backend: common.NewBackend(conf),
		ctx:     ctx,
		client:  client,
	}

	return &backend, nil

}

// InitGroup Group related functions
func (b *etcdBackend) InitGroup(groupUUID string, taskUUIDs []string) error {
	lease, err := b.getLease()
	if err != nil {
		return err
	}

	groupMeta := &tasks.GroupMeta{
		GroupUUID: groupUUID,
		TaskUUIDs: taskUUIDs,
		CreatedAt: time.Now().UTC(),
		TTL:       lease.TTL,
	}

	encoded, err := json.Marshal(groupMeta)
	if err != nil {
		return err
	}

	key := fmt.Sprintf(groupKey, groupUUID)
	_, err = b.client.Put(b.ctx, key, string(encoded), clientv3.WithLease(lease.ID))
	return err
}

// GroupCompleted ..
func (b *etcdBackend) GroupCompleted(groupUUID string, groupTaskCount int) (bool, error) {
	groupMeta, err := b.getGroupMeta(groupUUID)
	if err != nil {
		return false, err
	}

	taskStates, err := b.getStates(groupMeta.TaskUUIDs...)
	if err != nil {
		return false, err
	}

	var countSuccessTasks = 0
	for _, taskState := range taskStates {
		if taskState.IsCompleted() {
			countSuccessTasks++
		}
	}

	return countSuccessTasks == groupTaskCount, nil
}

// GroupTaskStates ..
func (b *etcdBackend) GroupTaskStates(groupUUID string, groupTaskCount int) ([]*tasks.TaskState, error) {
	groupMeta, err := b.getGroupMeta(groupUUID)
	if err != nil {
		return nil, err
	}
	if len(groupMeta.TaskUUIDs) != groupTaskCount {
		return nil, fmt.Errorf("group task count not equal, %d != %d", len(groupMeta.TaskUUIDs), groupTaskCount)
	}

	return b.getStates(groupMeta.TaskUUIDs...)
}

// TriggerChord ..
func (b *etcdBackend) TriggerChord(groupUUID string) (bool, error) {
	key := fmt.Sprintf(groupKey, groupUUID)
	resp, err := b.client.Get(b.ctx, key)
	if err != nil {
		return false, err
	}
	if len(resp.Kvs) == 0 {
		return false, fmt.Errorf("task %s not exist", groupUUID)
	}
	kv := resp.Kvs[0]

	meta := new(tasks.GroupMeta)

	decoder := json.NewDecoder(bytes.NewReader(kv.Value))
	decoder.UseNumber()

	if e := decoder.Decode(meta); e != nil {
		return false, e
	}

	if meta.ChordTriggered {
		return false, nil
	}

	lease, err := b.getLease()
	if err != nil {
		return false, err
	}

	// Set flag to true
	meta.ChordTriggered = true
	meta.TTL = lease.TTL

	// Update the group meta
	encoded, err := json.Marshal(&meta)
	if err != nil {
		return false, err
	}

	cmp := clientv3.Compare(clientv3.ModRevision(key), "=", kv.ModRevision)
	update := clientv3.OpPut(key, string(encoded), clientv3.WithLease(lease.ID))

	txnresp, err := b.client.Txn(b.ctx).If(cmp).Then(update).Commit()
	if err != nil {
		return false, err
	}

	// 有写入或者删除竞争
	if !txnresp.Succeeded {
		return false, fmt.Errorf("trigger chord failed, groupId: %s", groupUUID)
	}

	return true, nil

}

func (b *etcdBackend) getGroupMeta(groupUUID string) (*tasks.GroupMeta, error) {
	key := fmt.Sprintf(groupKey, groupUUID)
	resp, err := b.client.Get(b.ctx, key)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("task %s not exist", groupUUID)
	}
	kv := resp.Kvs[0]

	meta := new(tasks.GroupMeta)

	decoder := json.NewDecoder(bytes.NewReader(kv.Value))
	decoder.UseNumber()
	if err := decoder.Decode(meta); err != nil {
		return nil, err
	}
	return meta, nil
}

// SetStatePending updates task state to PENDING
func (b *etcdBackend) SetStatePending(signature *tasks.Signature) error {
	taskState := tasks.NewPendingTaskState(signature)
	return b.updateState(taskState)
}

// SetStateReceived updates task state to RECEIVED
func (b *etcdBackend) SetStateReceived(signature *tasks.Signature) error {
	taskState := tasks.NewReceivedTaskState(signature)
	b.mergeNewTaskState(taskState)
	return b.updateState(taskState)
}

// SetStateStarted updates task state to STARTED
func (b *etcdBackend) SetStateStarted(signature *tasks.Signature) error {
	taskState := tasks.NewStartedTaskState(signature)
	b.mergeNewTaskState(taskState)
	return b.updateState(taskState)
}

// SetStateRetry updates task state to RETRY
func (b *etcdBackend) SetStateRetry(signature *tasks.Signature) error {
	taskState := tasks.NewRetryTaskState(signature)
	b.mergeNewTaskState(taskState)
	return b.updateState(taskState)
}

// SetStateSuccess updates task state to SUCCESS
func (b *etcdBackend) SetStateSuccess(signature *tasks.Signature, results []*tasks.TaskResult) error {
	taskState := tasks.NewSuccessTaskState(signature, results)
	b.mergeNewTaskState(taskState)
	return b.updateState(taskState)
}

// SetStateFailure updates task state to FAILURE
func (b *etcdBackend) SetStateFailure(signature *tasks.Signature, err string) error {
	taskState := tasks.NewFailureTaskState(signature, err)
	b.mergeNewTaskState(taskState)
	return b.updateState(taskState)
}

// GetState ..
func (b *etcdBackend) GetState(taskUUID string) (*tasks.TaskState, error) {
	return b.getState(b.ctx, taskUUID)
}

func (b *etcdBackend) getState(ctx context.Context, taskUUID string) (*tasks.TaskState, error) {
	key := fmt.Sprintf(taskKey, taskUUID)
	resp, err := b.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("task %s not exist", taskUUID)
	}
	kv := resp.Kvs[0]

	state := new(tasks.TaskState)

	decoder := json.NewDecoder(bytes.NewReader(kv.Value))
	decoder.UseNumber()
	if err := decoder.Decode(state); err != nil {
		return nil, err
	}
	return state, nil
}

func (b *etcdBackend) mergeNewTaskState(newState *tasks.TaskState) {
	state, err := b.GetState(newState.TaskUUID)
	if err == nil {
		newState.CreatedAt = state.CreatedAt
		newState.TaskName = state.TaskName
	}
}

// PurgeState ..
func (b *etcdBackend) PurgeState(taskUUID string) error {
	key := fmt.Sprintf(taskKey, taskUUID)
	_, err := b.client.Delete(b.ctx, key)
	return err
}

// PurgeGroupMeta ..
func (b *etcdBackend) PurgeGroupMeta(groupUUID string) error {
	key := fmt.Sprintf(groupKey, groupUUID)
	_, err := b.client.Delete(b.ctx, key)
	return err
}

// getStates returns multiple task states
func (b *etcdBackend) getStates(taskUUIDs ...string) ([]*tasks.TaskState, error) {
	eg, ctx := errgroup.WithContext(b.ctx)
	eg.SetLimit(10)
	taskStates := make([]*tasks.TaskState, 0, len(taskUUIDs))
	var mtx sync.Mutex
	for _, taskUUID := range taskUUIDs {
		t := taskUUID
		eg.Go(func() error {
			state, err := b.getState(ctx, t)
			if err != nil {
				return err
			}

			mtx.Lock()
			taskStates = append(taskStates, state)
			mtx.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return taskStates, nil
}

// updateState saves current task state
func (b *etcdBackend) updateState(taskState *tasks.TaskState) error {
	lease, err := b.getLease()
	if err != nil {
		return err
	}
	taskState.TTL = lease.TTL

	encoded, err := json.Marshal(taskState)
	if err != nil {
		return err
	}

	key := fmt.Sprintf(taskKey, taskState.TaskUUID)
	_, err = b.client.Put(b.ctx, key, string(encoded), clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}

	log.DEBUG.Printf("update taskstate %s %s, %s", taskState.TaskName, taskState.TaskUUID, encoded)
	return nil
}

// getLease returns expiration for a stored task state
func (b *etcdBackend) getLease() (*clientv3.LeaseGrantResponse, error) {
	expiresIn := b.GetConfig().ResultsExpireIn
	if expiresIn <= 0 {
		// expire results after 1 hour by default
		expiresIn = config.DefaultResultsExpireIn
	}

	resp, err := b.client.Grant(b.ctx, int64(expiresIn))
	if err != nil {
		return nil, err
	}

	return resp, nil
}
