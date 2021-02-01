/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package rssche

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/coreos/etcd/clientv3"
	"github.com/google/uuid"
)

// Reporter reports resource info to etcd.
type Reporter struct {
	// resource name.
	name string

	// resource id.
	id string

	// resource keepalive ttl.
	ttl int64

	// etcd leaseid.
	leaseid clientv3.LeaseID

	// etcd client.
	etcdCli *clientv3.Client

	// mark delete event.
	isDelete bool
}

// NewReporter creates a new resource reporter.
// etcd key: schema/name/id
func NewReporter(name string, ttl int64) *Reporter {
	reporter := &Reporter{
		name: name,
		ttl:  ttl,
	}
	return reporter
}

// Init initialize a new reporter.
func (r *Reporter) Init(cfg clientv3.Config) error {
	cli, err := clientv3.New(cfg)
	if err != nil {
		return err
	}
	r.etcdCli = cli

	id, err := r.genResourceID()
	if err != nil {
		return err
	}
	r.id = id

	return nil
}

func (r *Reporter) genResourceID() (string, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return uuid.String(), nil
}

// AddRes add a new resource, register etcd KV and keepalive it.
func (r *Reporter) AddRes(metadata interface{}) error {
	bytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	// etcd KV.
	key := r.key(DEFAULTSCHEMA, r.name, r.id)
	value := string(bytes)

	resp, err := r.etcdCli.Grant(context.Background(), r.ttl)
	if err != nil {
		return err
	}
	r.leaseid = resp.ID

	if _, err := r.etcdCli.Put(context.Background(), key, value, clientv3.WithLease(r.leaseid)); err != nil {
		return err
	}

	// keepalive.
	if _, err := r.etcdCli.KeepAlive(context.Background(), r.leaseid); err != nil {
		return err
	}

	r.isDelete = false
	return nil
}

func (r *Reporter) key(schema, name, id string) string {
	return fmt.Sprintf("%s/%s/%s", schema, name, id)
}

// DeleteRes delete resource, stop keepalive.
func (r *Reporter) DeleteRes() {
	r.etcdCli.Revoke(context.Background(), r.leaseid)
	r.isDelete = true

	r.etcdCli.Close()
	log.Print("stop resource keepalive now!")
}

// UpdateRes update resource info, such as load info of server instance.
func (r *Reporter) UpdateRes(metadata interface{}) error {
	if r.isDelete {
		return errors.New("reporter stop, resource was deleted")
	}

	bytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	// update KV.
	key := r.key(DEFAULTSCHEMA, r.name, r.id)
	value := string(bytes)

	if _, err := r.etcdCli.Put(context.Background(), key, value, clientv3.WithLease(r.leaseid)); err != nil {
		log.Printf("reporter updates resource, %+v", err)

		resp, err := r.etcdCli.Grant(context.Background(), r.ttl)
		if err != nil {
			return err
		}
		r.leaseid = resp.ID

		if _, err := r.etcdCli.Put(context.Background(), key, value, clientv3.WithLease(r.leaseid)); err != nil {
			return err
		}
		if _, err := r.etcdCli.KeepAlive(context.Background(), r.leaseid); err != nil {
			return err
		}
		log.Print("reporter grant new lease and update success!")
	}
	return nil
}
