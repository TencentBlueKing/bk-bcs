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
	"encoding/base64"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"bscp.io/pkg/tools"
)

func (b *backend) pathKvEncrypt() *framework.Path {
	return &framework.Path{
		Pattern: "apps/" + framework.GenericNameRegex("app_id") + "/kvs/" + framework.GenericNameRegex("name") + "/encrypt",
		Fields: map[string]*framework.FieldSchema{
			"app_id": {
				Type: framework.TypeString,
			},
			"name": {
				Type: framework.TypeString,
			},
			"algorithm": {
				Type: framework.TypeString,
			},
			"key_name": {
				Type: framework.TypeString,
			},
		},

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathEncryptWrite,
			},
		},
	}
}

func (b *backend) pathEncryptWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	appID := d.Get("app_id").(string)
	if appID == "" {
		return logical.ErrorResponse("invalid app id"), nil
	}
	name := d.Get("name").(string)
	if name == "" {
		return logical.ErrorResponse("invalid name"), nil
	}

	algorithm := d.Get("algorithm").(string)
	keyName := d.Get("key_name").(string)

	kv, err := b.getKvStorage(ctx, req.Storage, appID, name)
	if err != nil {
		return nil, err
	}

	key, err := b.GetKeyStorage(ctx, req.Storage, appID, keyName, algorithm)
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
		publicKey, e := tools.SM2PublicKeyFromPEM([]byte(kv.Value))
		if e != nil {
			return nil, e
		}
		ciphertextByte, e := tools.SM2EncryptWithPublicKey(publicKey, []byte(key.Key))
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
