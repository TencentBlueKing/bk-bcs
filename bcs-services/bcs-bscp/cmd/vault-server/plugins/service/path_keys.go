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
				Type: framework.TypeString,
			},
			"name": {
				Type: framework.TypeString,
			},
			"algorithm": {
				Type: framework.TypeString,
			},
			"public_key": {
				Type: framework.TypeString,
			},
			"length": {
				Type:    framework.TypeInt,
				Default: DefaultKeySize,
			},
		},

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathKeysCreate,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathKeysCreate,
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathKeysRead,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathKeysDelete,
			},
		},
	}
}

// DefaultKeySize 默认密码长度
const DefaultKeySize = 2048

// EncryptionAlgorithm 表示加密算法的枚举类型
type EncryptionAlgorithm string

const (
	RSAEncryption EncryptionAlgorithm = "rsa"
	AESEncryption EncryptionAlgorithm = "aes"
	SM2Encryption EncryptionAlgorithm = "sm2"
	SM4Encryption EncryptionAlgorithm = "sm4"
)

// verifyPublicKey Check whether the public key is valid
func (s *keyStorage) verifyPublicKey() error {

	// Check the algorithm used by the public key
	switch s.Algorithm {
	case RSAEncryption:
		if _, err := tools.RSAPublicKeyFromPEM([]byte(s.Key)); err != nil {
			return err
		}
	case SM2Encryption:
		if _, err := tools.SM2PublicKeyFromPEM([]byte(s.Key)); err != nil {
			return err
		}

	}

	return nil
}

// GetKeyStorage Get key store
func (b *backend) GetKeyStorage(ctx context.Context, s logical.Storage, appID, name, algorithm string) (*keyStorage, error) {

	path := fmt.Sprintf("apps/%s/keys/%s/%s", appID, name, algorithm)

	entry, err := s.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, errors.New("")
	}

	k := new(keyStorage)
	err = json.Unmarshal(entry.Value, k)
	if err != nil {
		return nil, err
	}

	return k, nil

}

// validate Check whether the algorithm field is in the enumeration range
func (e EncryptionAlgorithm) validate() error {
	switch e {
	case RSAEncryption:
	case AESEncryption:
	case SM2Encryption:
	case SM4Encryption:
	default:
		return errors.New("unsupported encryption algorithm")
	}

	return nil
}

// validateCreate It is used to verify the key saving
func (s *keyStorage) validateCreate() error {
	if s.AppID == "" {
		return errors.New("app_id cannot be empty")
	}

	if s.Name == "" {
		return errors.New("name cannot be empty")
	}

	if err := s.Algorithm.validate(); err != nil {
		return err
	}

	if s.Key != "" {
		if err := s.verifyPublicKey(); err != nil {
			return err
		}
	}

	return nil
}

// pathKeysCreate
func (b *backend) pathKeysCreate(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	name := d.Get("name").(string)
	appID := d.Get("app_id").(string)
	pubKey := d.Get("public_key").(string)
	algorithm := d.Get("algorithm").(string)
	length := d.Get("length").(int)

	s := &keyStorage{
		AppID:     appID,
		Name:      name,
		Algorithm: EncryptionAlgorithm(algorithm),
		Key:       pubKey,
	}

	if err := s.validateCreate(); err != nil {
		return nil, err
	}

	var privKey string
	if s.Key == "" {
		switch s.Algorithm {
		case RSAEncryption:
			privateKey, publicKey, err := tools.GenerateRSAKeyPair(length)
			if err != nil {
				return nil, err
			}
			// 转换为PEM
			s.Key = string(tools.RSAPublicKeyToPEM(publicKey))
			privKey = string(tools.RSAPrivateKeyToPEM(privateKey))
		case SM2Encryption:
			privateKey, err := tools.GenerateSM2KeyPair(b.GetRandomReader())
			if err != nil {
				return nil, err
			}
			privateKeyPem, err := tools.SM2PrivateKeyToPEM(privateKey)
			if err != nil {
				return nil, err
			}
			publicKeyPem, err := tools.SM2PublicKeyToPEM(&privateKey.PublicKey)
			if err != nil {
				return nil, err
			}

			s.Key = string(publicKeyPem)
			privKey = string(privateKeyPem)
		}
	}

	encryptedKeyByte, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("apps/%s/keys/%s/%s", appID, name, algorithm)
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
	if name == "" {
		return logical.ErrorResponse("missing name parameter"), nil
	}
	appID := d.Get("app_id").(string)
	if appID == "" {
		return logical.ErrorResponse("missing app_id parameter"), nil
	}
	algorithm := d.Get("algorithm").(string)
	if err := EncryptionAlgorithm(algorithm).validate(); err != nil {
		return nil, err
	}

	s, err := b.GetKeyStorage(ctx, req.Storage, appID, name, algorithm)
	if err != nil {
		return nil, err
	}

	if s == nil {
		resp := logical.ErrorResponse("No value at %v", req.MountPoint)
		return resp, nil
	}

	// Generate the response
	resp := &logical.Response{
		Data: map[string]interface{}{
			"pub_key": s.Key,
		},
	}

	return resp, nil

}

func (b *backend) pathKeysDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	name := d.Get("name").(string)
	if name == "" {
		return logical.ErrorResponse("missing name parameter"), nil
	}
	appID := d.Get("app_id").(string)
	if appID == "" {
		return logical.ErrorResponse("missing app_id parameter"), nil
	}
	algorithm := d.Get("algorithm").(string)
	if err := EncryptionAlgorithm(algorithm).validate(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("apps/%s/keys/%s/%s", appID, name, algorithm)

	if err := req.Storage.Delete(ctx, path); err != nil {
		return nil, err
	}

	return nil, nil

}
