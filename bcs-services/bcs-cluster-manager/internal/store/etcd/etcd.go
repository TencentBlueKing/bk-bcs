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

// Package etcd xxx
package etcd

import (
	"context"
	"encoding/json"
	"path"
	"reflect"

	"github.com/pkg/errors"
	client "go.etcd.io/etcd/client/v3"
)

var (
	// ErrKeyNotFound not found
	ErrKeyNotFound = errors.New("key not found")
	// ErrTypeNotMatch not match
	ErrTypeNotMatch = errors.New("type not match")
)

// Store for etcd client
type Store struct {
	client     *client.Client
	pathPrefix string
}

// NewEtcdStore creates a new etcd store
func NewEtcdStore(prefix string, c *client.Client) *Store {
	return &Store{
		client:     c,
		pathPrefix: prefix,
	}
}

// Create for store data
func (s *Store) Create(ctx context.Context, key string, obj interface{}) error {
	key = path.Join(s.pathPrefix, key)
	objData, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = s.client.Put(ctx, key, string(objData))
	if err != nil {
		return err
	}
	return nil
}

// Delete for delete data
func (s *Store) Delete(ctx context.Context, key string) error {
	key = path.Join(s.pathPrefix, key)
	_, err := s.client.Delete(ctx, key)
	if err != nil {
		return err
	}
	return nil
}

// Get for get data
func (s *Store) Get(ctx context.Context, key string, objPtr interface{}) error {
	key = path.Join(s.pathPrefix, key)
	getResp, err := s.client.KV.Get(ctx, key)
	if err != nil {
		return err
	}
	if len(getResp.Kvs) == 0 {
		return ErrKeyNotFound
	}

	kv := getResp.Kvs[0]
	err = json.Unmarshal(kv.Value, objPtr)
	if err != nil {
		return ErrTypeNotMatch
	}
	return nil
}

// List for list data
func (s *Store) List(ctx context.Context, key string, listPtr interface{}) error {
	key = path.Join(s.pathPrefix, key)
	getResp, err := s.client.KV.Get(ctx, key, client.WithPrefix())
	if err != nil {
		return err
	}

	objType := reflect.TypeOf(listPtr).Elem().Elem()

	list := reflect.ValueOf(listPtr).Elem()
	list.Set(reflect.MakeSlice(reflect.TypeOf(listPtr).Elem(), 0, len(getResp.Kvs)))

	for _, kv := range getResp.Kvs {
		v := reflect.New(objType).Interface()
		errLocal := json.Unmarshal(kv.Value, v)
		if errLocal != nil {
			continue
		}
		list.Set(reflect.Append(list, reflect.ValueOf(v).Elem()))
	}
	return nil
}
