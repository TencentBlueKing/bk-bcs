/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func createBackendWithStorage(t testing.TB) (*backend, logical.Storage) {
	config := logical.TestBackendConfig()
	config.StorageView = &logical.InmemStorage{}

	b, _ := Backend(context.Background(), config)
	if b == nil {
		t.Fatalf("failed to create backend")
	}
	err := b.Backend.Setup(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	return b, config.StorageView
}
