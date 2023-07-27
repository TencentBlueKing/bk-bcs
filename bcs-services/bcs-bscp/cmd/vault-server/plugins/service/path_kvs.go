/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func (b *backend) pathKvs() *framework.Path {
	return &framework.Path{
		Pattern: "apps/" + framework.GenericNameRegex("app_id") + "/kvs/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"app_id": {
				Type: framework.TypeString,
			},
			"name": {
				Type: framework.TypeString,
			},
			"value": {
				Type: framework.TypeString,
			},
			"algorithm": {
				Type: framework.TypeString,
			},
			"pki_name": {
				Type: framework.TypeString,
			},
		},

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathKvWrite,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathKvWrite,
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.handleKvRead,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.handleKvDelete,
			},
		},

		ExistenceCheck: b.pathKvExistenceCheck,
	}
}

func (b *backend) pathKvExistenceCheck(ctx context.Context, req *logical.Request, d *framework.FieldData) (bool, error) {

	appID := d.Get("app_id").(string)
	name := d.Get("name").(string)

	path := fmt.Sprintf("apps/%s/kvs/%s", appID, name)

	entry, err := req.Storage.Get(ctx, path)
	if err != nil {
		return false, err
	}
	return entry != nil, nil

}

func (b *backend) pathKvWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	appID := d.Get("app_id").(string)
	name := d.Get("name").(string)
	value := d.Get("value").(string)

	path := fmt.Sprintf("apps/%s/kvs/%s", appID, name)

	kv := &kvStorage{
		AppID: appID,
		Name:  name,
		Value: value,
	}

	kvByte, err := json.Marshal(kv)
	if err != nil {
		return nil, err
	}

	entry := &logical.StorageEntry{
		Key:      path,
		Value:    kvByte,
		SealWrap: false,
	}
	err = req.Storage.Put(ctx, entry)
	if err != nil {
		return nil, err
	}

	resp := &logical.Response{}
	return resp, nil

}

func (b *backend) handleKvRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	appID := d.Get("app_id").(string)
	if appID == "" {
		return logical.ErrorResponse("invalid app id"), nil
	}
	name := d.Get("name").(string)
	if name == "" {
		return logical.ErrorResponse("invalid name"), nil
	}

	kv, err := b.getKvStorage(ctx, req.Storage, appID, name)
	if err != nil {
		return nil, err
	}

	resp := &logical.Response{
		Data: map[string]interface{}{
			"value": kv.Value,
		},
	}

	return resp, nil

}

func (b *backend) getKvStorage(ctx context.Context, s logical.Storage, appID, name string) (*kvStorage, error) {

	path := fmt.Sprintf("apps/%s/kvs/%s", appID, name)

	entry, err := s.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	kv := new(kvStorage)
	err = entry.DecodeJSON(kv)
	if err != nil {
		return nil, err
	}

	return kv, nil

}

func (b *backend) handleKvList(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	appID := d.Get("app_id").(string)
	key := d.Get("key").(string)

	path := fmt.Sprintf("apps/%s/kvs/%s", appID, key)

	_, err := req.Storage.List(ctx, path)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) handleKvDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	appID := d.Get("app_id").(string)
	if appID == "" {
		return logical.ErrorResponse("invalid app id"), nil
	}
	name := d.Get("name").(string)
	if name == "" {
		return logical.ErrorResponse("invalid name"), nil
	}

	path := fmt.Sprintf("apps/%s/kvs/%s", appID, name)

	// Store kv pairs in map at specified path
	err := req.Storage.Delete(ctx, path)
	if err != nil {
		return nil, err
	}

	resp := &logical.Response{}

	return resp, nil
}
