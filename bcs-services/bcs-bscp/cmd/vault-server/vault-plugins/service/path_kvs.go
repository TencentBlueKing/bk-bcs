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

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func (b *backend) pathKvs() *framework.Path {
	return &framework.Path{
		Pattern: "apps/" + framework.GenericNameRegex("app_id") + "/kvs/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"app_id": {
				Type:        framework.TypeString,
				Description: "Service ID",
			},
			"name": {
				Type:        framework.TypeString,
				Description: "kv stores the key name, unique under each service",
			},
			"value": {
				Type:        framework.TypeString,
				Description: "kv stored key values, it is recommended to upload the format through base64 encoding",
			},
		},

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback:    b.pathKvWrite,
				Description: "Create kv",
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback:    b.pathKvWrite,
				Description: "Updated kv",
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback:    b.pathKvRead,
				Description: "Read the key value from the key name",
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback:    b.pathKvDelete,
				Description: "Please exercise caution when deleting kv. Data deletion cannot be restored",
			},
			logical.ListOperation: &framework.PathOperation{
				Callback:    b.pathKvList,
				Description: "Get a list of kv under the service",
			},
		},

		ExistenceCheck: b.pathKvExistenceCheck,
	}
}

func (b *backend) ValidateAppID(appID string) error {
	if appID == "" {
		return errors.New("invalid app_id")
	}
	return nil
}

func (b *backend) ValidateName(name string) error {
	if name == "" {
		return errors.New("invalid name")
	}
	return nil
}

func (b *backend) pathKvExistenceCheck(ctx context.Context, req *logical.Request,
	d *framework.FieldData) (bool, error) {

	appID := d.Get("app_id").(string)
	name := d.Get("name").(string)

	path := fmt.Sprintf("apps/%s/kvs/%s", appID, name)
	entry, err := req.Storage.Get(ctx, path)
	if err != nil {
		return false, err
	}
	return entry != nil, nil
}

func (b *backend) pathKvList(ctx context.Context, req *logical.Request,
	d *framework.FieldData) (*logical.Response, error) {

	appID := d.Get("app_id").(string)
	if err := b.ValidateAppID(appID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("apps/%s/kvs/", appID)
	entrys, err := req.Storage.List(ctx, path)
	if err != nil {
		return nil, err
	}

	kvs := make(map[string]interface{})
	for _, entry := range entrys {
		kv, e := b.getKvStorage(ctx, req.Storage, appID, entry)
		if e != nil {
			return nil, e
		}
		kvs[kv.Name] = kv.Value
	}

	return &logical.Response{
		Data: kvs,
	}, nil
}

func (b *backend) pathKvWrite(ctx context.Context, req *logical.Request,
	d *framework.FieldData) (*logical.Response, error) {

	appID := d.Get("app_id").(string)
	if err := b.ValidateAppID(appID); err != nil {
		return nil, err
	}
	name := d.Get("name").(string)
	if err := b.ValidateName(name); err != nil {
		return nil, err
	}
	value := d.Get("value").(string)

	kv := &kvStorage{
		AppID: appID,
		Name:  name,
		Value: value,
	}
	err := b.SaveKvStorage(ctx, req.Storage, kv)
	if err != nil {
		return nil, err
	}

	return nil, nil

}

func (b *backend) pathKvRead(ctx context.Context, req *logical.Request,
	d *framework.FieldData) (*logical.Response, error) {

	appID := d.Get("app_id").(string)
	if err := b.ValidateAppID(appID); err != nil {
		return nil, err
	}
	name := d.Get("name").(string)
	if err := b.ValidateName(name); err != nil {
		return nil, err
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

// getKvStorage get Kv Storage
func (b *backend) getKvStorage(ctx context.Context, s logical.Storage, appID, name string) (*kvStorage, error) {

	path := fmt.Sprintf("apps/%s/kvs/%s", appID, name)

	entry, err := s.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, errors.New("received an empty entry")
	}
	kv := new(kvStorage)
	err = entry.DecodeJSON(kv)
	if err != nil {
		return nil, err
	}

	return kv, nil

}

// SaveKvStorage save Kv Storage
func (b *backend) SaveKvStorage(ctx context.Context, s logical.Storage, kv *kvStorage) error {

	path := fmt.Sprintf("apps/%s/kvs/%s", kv.AppID, kv.Name)

	kvJson, err := json.Marshal(kv)
	if err != nil {
		return err
	}

	entry := &logical.StorageEntry{
		Key:   path,
		Value: kvJson,
	}

	err = s.Put(ctx, entry)
	if err != nil {
		return err
	}

	return nil

}

func (b *backend) pathKvDelete(ctx context.Context, req *logical.Request,
	d *framework.FieldData) (*logical.Response, error) {

	appID := d.Get("app_id").(string)
	if err := b.ValidateAppID(appID); err != nil {
		return nil, err
	}
	name := d.Get("name").(string)
	if err := b.ValidateName(name); err != nil {
		return nil, err
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
