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
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

func (b *backend) pathKvEncrypt() *framework.Path {
	return &framework.Path{
		Pattern: "apps/" + framework.GenericNameRegex("app_id") + "/kvs/" + // nolint goconst
			framework.GenericNameRegex("name") + "/encrypt",
		Fields: map[string]*framework.FieldSchema{
			"app_id": {
				Type:        framework.TypeString,
				Description: "Service ID",
			},
			"name": {
				Type:        framework.TypeString,
				Description: "kv stores the key name, unique under each service",
			},
			"algorithm": {
				Type:        framework.TypeString,
				Description: "Encryption algorithm : rsa,sm2",
			},
			"key_name": {
				Type:        framework.TypeString,
				Description: "Specifies the name of the encryption public key",
			},
		},

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback:    b.pathEncryptWrite,
				Description: "kv is obtained encrypted",
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback:    b.pathEncryptWrite,
				Description: "kv is obtained encrypted",
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback:    b.pathEncryptWrite,
				Description: "kv is obtained encrypted",
			},
		},

		ExistenceCheck: b.pathEncryptExistenceCheck,
	}
}

func (b *backend) pathEncryptExistenceCheck(ctx context.Context, req *logical.Request,
	d *framework.FieldData) (bool, error) {

	appID := d.Get("app_id").(string)
	name := d.Get("name").(string)
	algorithm := d.Get("algorithm").(string)

	pkiPath := fmt.Sprintf("apps/%s/pkis/%s/%s", appID, name, algorithm)
	entry, err := req.Storage.Get(ctx, pkiPath)
	if err != nil {
		return false, err
	}
	if entry == nil {
		return false, nil
	}

	kvPath := fmt.Sprintf("apps/%s/kvs/%s", appID, name)
	entry, err = req.Storage.Get(ctx, kvPath)
	if err != nil {
		return false, err
	}
	return entry != nil, nil

}

func (b *backend) pathEncryptWrite(ctx context.Context, req *logical.Request,
	d *framework.FieldData) (*logical.Response, error) {

	appID := d.Get("app_id").(string)
	if err := b.ValidateAppID(appID); err != nil {
		return nil, err
	}
	name := d.Get("name").(string)
	if err := b.ValidateAppID(name); err != nil {
		return nil, err
	}
	kv, err := b.getKvStorage(ctx, req.Storage, appID, name)
	if err != nil {
		return nil, err
	}

	algorithm := d.Get("algorithm").(string)
	keyName := d.Get("key_name").(string)
	key, err := b.GetPkiStorage(ctx, req.Storage, appID, keyName, algorithm)
	if err != nil {
		return nil, err
	}

	var ciphertext string
	switch algorithm {
	case string(RSAEncryption):
		publicKey, e := tools.RSAPublicKeyFromPEM([]byte(key.Key))
		if e != nil {
			return nil, e
		}
		ciphertextByte, e := tools.RSAEncryptWithPublicKey(publicKey, []byte(kv.Value))
		if e != nil {
			return nil, e
		}
		ciphertext = string(ciphertextByte)
	case string(SM2Encryption):
		publicKey, e := tools.SM2PublicKeyFromPEM([]byte(key.Key))
		if e != nil {
			return nil, e
		}
		ciphertextByte, e := tools.SM2EncryptWithPublicKey(publicKey, []byte(kv.Value))
		if e != nil {
			return nil, e
		}
		ciphertext = string(ciphertextByte)
	}

	ciphertextBase := base64.StdEncoding.EncodeToString([]byte(ciphertext))

	resp := &logical.Response{
		Data: map[string]interface{}{
			"ciphertext": ciphertextBase,
		},
	}

	return resp, nil

}
