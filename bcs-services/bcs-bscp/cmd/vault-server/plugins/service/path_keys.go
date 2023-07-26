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
	"errors"
	"fmt"

	"bscp.io/pkg/tools"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// func (b *backend) pathKvs() *framework.Path {
func (b *backend) pathKeys() *framework.Path {
	return &framework.Path{
		Pattern: "apps/" + framework.GenericNameRegex("app_id") +
			"/keys/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"app_id": {
				Type:        framework.TypeString,
				Description: "....",
			},
			"name": {
				Type:        framework.TypeString,
				Description: "....",
			},
			"type": {
				Type:        framework.TypeString,
				Description: "....",
			},
			"public_key": {
				Type:        framework.TypeString,
				Description: "....",
			},
			"length": {
				Type:        framework.TypeInt,
				Description: "",
				Default:     DefaultKeySize,
			},
		},

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathKeysCreate,
				Summary:  "创建keys",
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathKeysCreate,
				Summary:  "更新keys",
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathKeysRead,
				Summary:  "读取Kv",
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathKeysDelete,
				Summary:  "",
			},
			logical.ListOperation: &framework.PathOperation{
				Callback: nil,
				Summary:  "",
			},

			//logical.ListOperation: &framework.PathOperation{
			//	Callback: b.pathKvWrite,
			//},
		},
	}
}

const DefaultKeySize = 2048

// EncryptionAlgorithm 表示加密算法的枚举类型
type EncryptionAlgorithm string

const (
	RSAEncryption EncryptionAlgorithm = "RSA"
	AESEncryption EncryptionAlgorithm = "AES"
	SM2Encryption EncryptionAlgorithm = "SM2"
	SM4Encryption EncryptionAlgorithm = "SM4"
	// 添加其他类型的加密算法...
)

// EncryptedKey 保存不同类型的加密密钥及其对应的加密算法
type EncryptedKey struct {
	AppID     string
	Name      string
	Algorithm EncryptionAlgorithm
	Key       string
}

func (e *EncryptedKey) ValidateEncryptionAlgorithm() error {

	switch e.Algorithm {
	case RSAEncryption:
	case AESEncryption:
	case SM2Encryption:
	case SM4Encryption:
	default:
		return errors.New("unsupported encryption algorithm")
	}

	return nil

}

func (e *EncryptedKey) VerifyPublicKey() error {

	// Check the algorithm used by the public key
	switch e.Algorithm {
	case RSAEncryption:
		if err := tools.VerifyRSAPublicKey(e.Key); err != nil {
			return err
		}
	case SM2Encryption:
	}

	return nil

}

func (b *backend) GetKeyStorage(ctx context.Context, s logical.Storage, appID, name, algorithm string) (*keyStorage, error) {

	path := fmt.Sprintf("apps/%s/keys/%s", appID, name)

	entry, err := s.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	k := new(keyStorage)
	err = json.Unmarshal(entry.Value, k)
	if err != nil {
		return nil, err
	}

	return k, nil

}

func (e *EncryptedKey) ValidateCreate() error {

	if e.AppID == "" {
		return errors.New("app_id cannot be empty")
	}

	if e.Name == "" {
		return errors.New("name cannot be empty")
	}

	if err := e.ValidateEncryptionAlgorithm(); err != nil {
		return err
	}

	if e.Key != "" {
		if err := e.VerifyPublicKey(); err != nil {
			return err
		}
	}

	return nil

}

func (b *backend) pathKeysCreate(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	name := d.Get("name").(string)
	appID := d.Get("app_id").(string)
	pubKey := d.Get("public_key").(string)
	algorithm := d.Get("type").(string)
	length := d.Get("length").(int)

	encryptedKey := &EncryptedKey{
		AppID:     appID,
		Name:      name,
		Algorithm: EncryptionAlgorithm(algorithm),
		Key:       pubKey,
	}

	if err := encryptedKey.ValidateCreate(); err != nil {
		return nil, err
	}

	var privKey string
	if encryptedKey.Key == "" {
		switch encryptedKey.Algorithm {
		case RSAEncryption:
			privKeyTmp, pubKeyTmp, err := tools.GenerateRSAKeyPairToString(length)
			if err != nil {
				return nil, err
			}
			encryptedKey.Key = pubKeyTmp
			privKey = privKeyTmp
		case SM2Encryption:
			privKeyTmp, pubKeyTmp, err := tools.GenerateSM2KeyPairToString(b.GetRandomReader())
			if err != nil {
				return nil, err
			}
			encryptedKey.Key = pubKeyTmp
			privKey = privKeyTmp
		}
	}

	encryptedKeyByte, err := json.Marshal(encryptedKey)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("apps/%s/keys/%s", appID, name)
	entry := &logical.StorageEntry{
		Key:   path,
		Value: encryptedKeyByte,
	}
	err = req.Storage.Put(ctx, entry)
	if err != nil {
		return nil, err
	}

	// 如果是服务端创建密钥对，则把私钥返回
	resp := &logical.Response{
		Data: map[string]interface{}{
			"private_key": privKey,
		},
	}

	return resp, err
}

func (b *backend) pathKeysRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	name := d.Get("name").(string)
	appID := d.Get("app_id").(string)

	path := fmt.Sprintf("apps/%s/keys/%s", appID, name)

	entry, err := req.Storage.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	key := new(EncryptedKey)
	err = entry.DecodeJSON(key)
	if err != nil {
		return nil, err
	}

	fetchedData := entry.Value
	if fetchedData == nil {
		resp := logical.ErrorResponse("No value at %v%v", req.MountPoint, path)
		return resp, nil
	}

	// Generate the response
	resp := &logical.Response{
		Data: map[string]interface{}{
			"data": key,
		},
	}

	return resp, nil

}

func (b *backend) pathKeysDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	name := d.Get("name").(string)
	appID := d.Get("app_id").(string)

	path := fmt.Sprintf("apps/%s/keys/%s", appID, name)

	if err := req.Storage.Delete(ctx, path); err != nil {
		return nil, err
	}

	return nil, nil

}
