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

// Package etcd is broker use etcd
package etcd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/RichardKnop/machinery/v2/brokers/errs"
	"github.com/RichardKnop/machinery/v2/brokers/iface"
	"github.com/RichardKnop/machinery/v2/common"
	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/log"
	"github.com/RichardKnop/machinery/v2/tasks"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type etcdBroker struct {
	common.Broker
	ctx       context.Context
	client    *clientv3.Client
	wg        sync.WaitGroup
	assignMap map[string]bool
	mtx       sync.Mutex
}

// New ..
func New(ctx context.Context, conf *config.Config) (iface.Broker, error) {
	etcdConf := clientv3.Config{
		Endpoints:   []string{conf.Broker},
		Context:     ctx,
		DialTimeout: time.Second * 5,
		TLS:         conf.TLSConfig,
	}

	client, err := clientv3.New(etcdConf)
	if err != nil {
		return nil, err
	}

	broker := etcdBroker{
		Broker:    common.NewBroker(conf),
		ctx:       ctx,
		client:    client,
		assignMap: map[string]bool{},
	}

	return &broker, nil

}

// StartConsuming ..
func (b *etcdBroker) StartConsuming(consumerTag string, concurrency int, taskProcessor iface.TaskProcessor) (bool, error) {
	if concurrency < 1 {
		concurrency = runtime.NumCPU()
	}
	b.Broker.StartConsuming(consumerTag, concurrency, taskProcessor)

	log.INFO.Printf("[*] Waiting for messages, concurrency=%d. To exit press CTRL+C", concurrency)

	// Channel to which we will push tasks ready for processing by worker
	deliveries := make(chan Delivery)

	ctx, cancel := context.WithCancel(b.ctx)
	defer cancel()

	// list watch task
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()

		for {
			select {
			// A way to stop this goroutine from b.StopConsuming
			case <-b.GetStopChan():
				return
			default:
				err := b.listWatchTasks(ctx, getQueue(b.GetConfig(), taskProcessor))
				if err != nil {
					log.ERROR.Printf("handle list watch task err: %s", err)
				}
			}
		}

	}()

	// A receiving goroutine keeps popping messages from the queue by BLPOP
	// If the message is valid and can be unmarshaled into a proper structure
	// we send it to the deliveries channel
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		defer cancel()

		for {
			select {
			// A way to stop this goroutine from b.StopConsuming
			case <-b.GetStopChan():
				close(deliveries)
				return

			default:
				if !taskProcessor.PreConsumeHandler() {
					continue
				}

				task := b.nextTask(getQueue(b.GetConfig(), taskProcessor), consumerTag)
				if task == nil {
					time.Sleep(time.Second)
					continue
				}

				deliveries <- task
			}
		}
	}()

	// A goroutine to watch for delayed tasks and push them to deliveries
	// channel for consumption by the worker
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		defer cancel()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			// A way to stop this goroutine from b.StopConsuming
			case <-b.GetStopChan():
				return

			case <-ticker.C:
				err := b.handleDelayedTask(ctx)
				if err != nil {
					log.ERROR.Printf("handleDelayedTask err: %s", err)
				}
			}
		}
	}()

	if err := b.consume(deliveries, concurrency, taskProcessor); err != nil {
		return b.GetRetry(), err
	}

	b.wg.Wait()

	return b.GetRetry(), nil
}

// consume takes delivered messages from the channel and manages a worker pool
// to process tasks concurrently
func (b *etcdBroker) consume(deliveries <-chan Delivery, concurrency int, taskProcessor iface.TaskProcessor) error {
	eg, ctx := errgroup.WithContext(b.ctx)

	for i := 0; i < concurrency; i++ {
		eg.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()

				case t, ok := <-deliveries:
					if !ok {
						return nil
					}

					if err := b.consumeOne(t, taskProcessor); err != nil {
						return err
					}
				}
			}
		})
	}

	return eg.Wait()

}

// consumeOne processes a single message using TaskProcessor
func (b *etcdBroker) consumeOne(delivery Delivery, taskProcessor iface.TaskProcessor) error {
	// If the task is not registered, we requeue it,
	// there might be different workers for processing specific tasks
	if !b.IsTaskRegistered(delivery.Signature().Name) {
		log.INFO.Printf("Task not registered with this worker. Requeuing message: %s", delivery.Body())

		if !delivery.Signature().IgnoreWhenTaskNotRegistered {
			delivery.Nack()
		}
		return nil
	}

	log.DEBUG.Printf("Received new message: %s", delivery.Body())
	defer delivery.Ack()

	return taskProcessor.Process(delivery.Signature())
}

// StopConsuming 停止
func (b *etcdBroker) StopConsuming() {
	b.Broker.StopConsuming()

	b.wg.Wait()
}

// Publish put kvs to etcd stor
func (b *etcdBroker) Publish(ctx context.Context, signature *tasks.Signature) error {
	// Adjust routing key (this decides which queue the message will be published to)
	b.Broker.AdjustRoutingKey(signature)

	now := time.Now()
	msg, err := json.Marshal(signature)
	if err != nil {
		return fmt.Errorf("JSON marshal error: %s", err)
	}

	key := fmt.Sprintf("/machinery/v2/broker/pending_tasks/%s/%s", signature.RoutingKey, signature.UUID)

	// Check the ETA signature field, if it is set and it is in the future,
	// delay the task
	if signature.ETA != nil && signature.ETA.After(now) {
		key = fmt.Sprintf("/machinery/v2/broker/delayed_tasks/eta-%d/%s/%s",
			signature.ETA.UnixMilli(), signature.RoutingKey, signature.UUID)
		_, err = b.client.Put(ctx, key, string(msg))
		return err
	}

	_, err = b.client.Put(ctx, key, string(msg))
	return err
}

func (b *etcdBroker) getTasks(ctx context.Context, key string) ([]*tasks.Signature, error) {
	resp, err := b.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	result := make([]*tasks.Signature, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		if strings.Contains(string(kv.Key), "/assign") {
			continue
		}

		signature := new(tasks.Signature)
		decoder := json.NewDecoder(bytes.NewReader(kv.Value))
		decoder.UseNumber()
		if err := decoder.Decode(signature); err != nil {
			return nil, errs.NewErrCouldNotUnmarshalTaskSignature(kv.Value, err)
		}

		result = append(result, signature)
	}

	return result, nil
}

// GetPendingTasks 获取执行队列, 任务统计可使用
func (b *etcdBroker) GetPendingTasks(queue string) ([]*tasks.Signature, error) {
	if queue == "" {
		queue = b.GetConfig().DefaultQueue
	}

	key := fmt.Sprintf("/machinery/v2/broker/pending_tasks/%s", queue)
	items, err := b.getTasks(b.ctx, key)
	if err != nil {
		return nil, err
	}

	return items, nil
}

// GetDelayedTasks 任务统计可使用
func (b *etcdBroker) GetDelayedTasks() ([]*tasks.Signature, error) {
	key := "/machinery/v2/broker/delayed_tasks"

	items, err := b.getTasks(b.ctx, key)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (b *etcdBroker) nextTask(queue string, consumerTag string) Delivery {
	b.mtx.Lock()
	assignMap := make(map[string]bool, len(b.assignMap))
	for k, v := range b.assignMap {
		assignMap[k] = v
	}
	b.mtx.Unlock()

	for k, assigned := range assignMap {
		if assigned {
			continue
		}
		if !strings.Contains(k, queue) {
			continue
		}

		d, err := NewDelivery(b.ctx, b.client, k, consumerTag)
		if err != nil {
			continue
		}

		return d
	}

	return nil
}

func (b *etcdBroker) setAssign(key string, assign bool) bool {
	if !strings.Contains(key, "/assign") {
		return false
	}
	k := strings.TrimSuffix(key, "/assign")

	if _, ok := b.assignMap[k]; ok {
		b.assignMap[k] = assign
	}

	return true
}

func (b *etcdBroker) listWatchTasks(ctx context.Context, queue string) error {
	keyPrefix := fmt.Sprintf("/machinery/v2/broker/pending_tasks/%s", queue)

	// List
	listCtx, listCancel := context.WithTimeout(ctx, time.Second*10)
	defer listCancel()
	resp, err := b.client.Get(listCtx, keyPrefix, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		return err
	}

	b.mtx.Lock()
	b.assignMap = map[string]bool{}
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		if b.setAssign(key, true) {
			continue
		}
		b.assignMap[key] = false
	}
	b.mtx.Unlock()

	// Watch
	watchCtx, watchCancel := context.WithTimeout(ctx, time.Minute*10)
	defer watchCancel()
	wc := b.client.Watch(watchCtx, keyPrefix, clientv3.WithPrefix(), clientv3.WithKeysOnly(), clientv3.WithRev(resp.Header.Revision))
	for wresp := range wc {
		if wresp.Err() != nil {
			return watchCtx.Err()
		}

		b.mtx.Lock()
		for _, ev := range wresp.Events {
			key := string(ev.Kv.Key)
			if ev.Type == clientv3.EventTypeDelete {
				if b.setAssign(key, false) {
					continue
				}

				delete(b.assignMap, key)
			}

			if ev.Type == clientv3.EventTypePut {
				if b.setAssign(key, true) {
					continue
				}

				b.assignMap[key] = false
			}
		}

		b.mtx.Unlock()
	}

	return nil
}

func (b *etcdBroker) handleDelayedTask(ctx context.Context) error {
	ttl := time.Second * 10
	ctx, cancel := context.WithTimeout(ctx, ttl)
	defer cancel()

	// 创建一个新的session
	s, err := concurrency.NewSession(b.client, concurrency.WithTTL(int(ttl.Seconds())))
	if err != nil {
		return err
	}
	defer s.Orphan()

	lockKey := "/machinery/v2/lock/delayed_tasks"
	m := concurrency.NewMutex(s, lockKey)

	if err = m.Lock(ctx); err != nil {
		return err
	}
	defer m.Unlock(ctx) // nolint

	keyPrefix := "/machinery/v2/broker/delayed_tasks/eta-"
	end := strconv.FormatInt(time.Now().UnixMilli(), 10)
	resp, err := b.client.Get(b.ctx, keyPrefix+"0", clientv3.WithRange(keyPrefix+end))
	if err != nil {
		return err
	}
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		parts := strings.Split(key, "/")
		if len(parts) != 8 {
			log.WARNING.Printf("invalid delay task %s, continue", key)
			continue
		}
		cmp := clientv3.Compare(clientv3.ModRevision(key), "=", kv.ModRevision)
		deleteReq := clientv3.OpDelete(key)
		pendingKey := fmt.Sprintf("/machinery/v2/broker/pending_tasks/%s/%s", parts[6], parts[7])
		putReq := clientv3.OpPut(pendingKey, string(kv.Value))
		c, err := b.client.Txn(b.ctx).If(cmp).Then(deleteReq, putReq).Commit()
		if err != nil {
			return fmt.Errorf("handle delay task %s: %w", key, err)
		}
		if !c.Succeeded {
			log.WARNING.Printf("handle delay task %s not success", key)
			continue
		}
		log.DEBUG.Printf("send delay task %s to pending queue done", key)
	}

	return nil
}

func getQueue(config *config.Config, taskProcessor iface.TaskProcessor) string {
	customQueue := taskProcessor.CustomQueue()
	if customQueue == "" {
		return config.DefaultQueue
	}
	return customQueue
}
