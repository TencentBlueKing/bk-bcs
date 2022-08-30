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
 *
 */

// Package etcd xxx
package etcd

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/storage"
)

// NewStorage create etcd storage
func NewStorage() storage.Storage {
	s := &eStorage{}

	return s
}

// eStorage storage data in etcd
type eStorage struct {
}

// Add xxx
func (e *eStorage) Add(key string, value []byte) error {
	return nil
}

// Delete xxx
func (e *eStorage) Delete(key string) ([]byte, error) {
	return nil, nil
}

// Update xxx
func (e *eStorage) Update(key string, value []byte) error {
	return nil
}

// Get xxx
func (e *eStorage) Get(key string) ([]byte, error) {
	return nil, nil
}

// List xxx
func (e *eStorage) List(key string) ([]string, error) {
	return nil, nil
}

// Register xxx
func (e *eStorage) Register(path string, data []byte) error {
	return nil
}

// RegisterAndWatch xxx
func (e *eStorage) RegisterAndWatch(path string, data []byte) error {
	return nil
}

// Exist xxx
func (e *eStorage) Exist(key string) (bool, error) {
	return true, nil
}

// GetLocker xxx
func (e *eStorage) GetLocker(path string) (storage.Locker, error) {
	return nil, nil
}

// Stop xxx
func (e *eStorage) Stop() {

}
