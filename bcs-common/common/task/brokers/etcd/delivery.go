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

package etcd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RichardKnop/machinery/v2/brokers/errs"
	"github.com/RichardKnop/machinery/v2/log"
	"github.com/RichardKnop/machinery/v2/tasks"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Delivery task delivery with ack and nack
type Delivery interface {
	Ack()
	Nack()
	Body() []byte
	Signature() *tasks.Signature
}

type deliver struct {
	ctx         context.Context
	client      *clientv3.Client
	signature   *tasks.Signature
	value       []byte
	key         string
	node        string
	aliveCancel func()
}

// NewDelivery create the task delivery
func NewDelivery(ctx context.Context, client *clientv3.Client, key string, node string) (Delivery, error) {
	d := &deliver{
		ctx:    ctx,
		client: client,
		key:    key,
		node:   node,
	}

	if err := d.assign(key, node); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *deliver) assign(key string, node string) error {
	assignKey := fmt.Sprintf("%s/assign", key)
	ctx, cancel := context.WithTimeout(d.ctx, time.Second*5)
	defer cancel()

	grantResp, err := d.client.Grant(ctx, 30)
	if err != nil {
		return err
	}

	keyExist := clientv3.Compare(clientv3.CreateRevision(key), ">", 0)
	assignNotExist := clientv3.Compare(clientv3.CreateRevision(assignKey), "=", 0)
	putReq := clientv3.OpPut(assignKey, node, clientv3.WithLease(grantResp.ID))
	getReq := clientv3.OpGet(key)
	resp, err := d.client.Txn(ctx).If(keyExist, assignNotExist).Then(putReq, getReq).Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		return fmt.Errorf("task %s not exist or already assign", key)
	}

	if len(resp.Responses) < 2 {
		return fmt.Errorf("task %s tnx resp invalid, count=%d", key, len(resp.Responses))
	}
	getResp := resp.Responses[1].GetResponseRange()
	if len(getResp.Kvs) == 0 || len(getResp.Kvs[0].Value) == 0 {
		return fmt.Errorf("task %s have no body", key)
	}
	kv := getResp.Kvs[0]

	signature := new(tasks.Signature)
	decoder := json.NewDecoder(bytes.NewReader(kv.Value))
	decoder.UseNumber()
	if err = decoder.Decode(signature); err != nil {
		return errs.NewErrCouldNotUnmarshalTaskSignature(kv.Value, err)
	}

	aliveCtx, aliveCancel := context.WithCancel(d.ctx)
	keepRespCh, err := d.client.KeepAlive(aliveCtx, grantResp.ID)
	if err != nil {
		aliveCancel()
		return err
	}

	go func() {
		defer aliveCancel()

		for {
			select {
			case <-d.ctx.Done():
				return
			case <-aliveCtx.Done():
				return
			case _, ok := <-keepRespCh:
				if !ok {
					return
				}
			}
		}
	}()

	d.aliveCancel = aliveCancel
	d.signature = signature
	d.value = kv.Value
	return nil
}

// Ack acknowledged the task is done
func (d *deliver) Ack() {
	defer d.aliveCancel()

	ctx, cancel := context.WithTimeout(d.ctx, time.Second*2)
	defer cancel()

	_, err := d.client.Delete(ctx, d.key, clientv3.WithPrefix())
	if err != nil {
		log.ERROR.Printf("ack task %s err: %s", d.key, err)
	}
}

// Nack negatively acknowledge the delivery of task should handle again
func (d *deliver) Nack() {
	defer d.aliveCancel()

	assignKey := fmt.Sprintf("%s/assign", d.key)
	ctx, cancel := context.WithTimeout(d.ctx, time.Second*2)
	defer cancel()

	_, err := d.client.Delete(ctx, assignKey)
	if err != nil {
		log.ERROR.Printf("nack task %s err: %s", d.key, err)
	}
}

// Signature return the task Signature
func (d *deliver) Signature() *tasks.Signature {
	return d.signature
}

// Body return the task body
func (d *deliver) Body() []byte {
	return d.value
}
