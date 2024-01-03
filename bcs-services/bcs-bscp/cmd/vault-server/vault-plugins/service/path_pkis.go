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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// func (b *backend) pathKvs() *framework.Path {
func (b *backend) pathKeys() *framework.Path {
	return &framework.Path{
		Pattern: "apps/" + framework.GenericNameRegex("app_id") +
			"/pkis/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"app_id": {
				Type:        framework.TypeString,
				Description: "Service ID",
			},
			"name": {
				Type:        framework.TypeString,
				Description: "pki name，the name_algorithm is unique in the same service",
			},
			"algorithm": {
				Type:        framework.TypeString,
				Description: "Encryption algorithm : rsa,sm2",
			},
			"pub_key": {
				Type:        framework.TypeString,
				Description: "The public key, when empty, is created by vault-plugin, and the private key is returned",
			},
			"length": {
				Type:        framework.TypeInt,
				Default:     DefaultKeySize,
				Description: "The value is optional. The default value is 2048",
			},
		},

		ExistenceCheck: b.pathPkiExistenceCheck,

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback:    b.pathPkiWrite,
				Description: "Creating or uploading a key pair",
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback:    b.pathPkiWrite,
				Description: "Update public key",
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback:    b.pathPkiRead,
				Description: "Get a public key",
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback:    b.pathPkiDelete,
				Description: "Delete a public key",
			},
		},
	}
}

func (b *backend) pathPkiExistenceCheck(ctx context.Context, req *logical.Request,
	d *framework.FieldData) (bool, error) {

	appID := d.Get("app_id").(string)
	name := d.Get("name").(string)
	algorithm := d.Get("algorithm").(string)

	path := fmt.Sprintf("apps/%s/pkis/%s/%s", appID, name, algorithm)
	entry, err := req.Storage.Get(ctx, path)
	if err != nil {
		return false, err
	}

	return entry == nil, nil

}

// DefaultKeySize 默认密码长度
const DefaultKeySize = 2048

// PKIEncryptionAlgorithm 表示加密算法的枚举类型
type PKIEncryptionAlgorithm string

const (
	// RSAEncryption 非对称加密算法 rsa
	RSAEncryption PKIEncryptionAlgorithm = "rsa"
	// SM2Encryption 非对称加密算法 sm2
	SM2Encryption PKIEncryptionAlgorithm = "sm2"
)

// verifyPkiPublicKey Check whether the public key is valid
func (b *backend) verifyPkiPublicKey(algorithm PKIEncryptionAlgorithm, pubKey string) error {

	// Check the algorithm used by the public key
	switch algorithm {
	case RSAEncryption:
		if _, err := tools.RSAPublicKeyFromPEM([]byte(pubKey)); err != nil {
			return err
		}
	case SM2Encryption:
		if _, err := tools.SM2PublicKeyFromPEM([]byte(pubKey)); err != nil {
			return err
		}

	}
	return nil
}

// GetPkiStorage Get pki store
func (b *backend) GetPkiStorage(ctx context.Context, s logical.Storage, appID, name,
	algorithm string) (*pkiStorage, error) {

	path := fmt.Sprintf("apps/%s/pkis/%s/%s", appID, name, algorithm)
	entry, err := s.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, errors.New("received an empty entry")
	}

	k := new(pkiStorage)
	err = json.Unmarshal(entry.Value, k)
	if err != nil {
		return nil, err
	}

	return k, nil
}

func (b *backend) SavePkiStorage(ctx context.Context, s logical.Storage, pki *pkiStorage) error {

	pkiJson, err := json.Marshal(pki)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("apps/%s/pkis/%s/%s", pki.AppID, pki.Name, pki.Algorithm)
	entry := &logical.StorageEntry{
		Key:   path,
		Value: pkiJson,
	}
	err = s.Put(ctx, entry)
	if err != nil {
		return err
	}

	return nil

}

// validate Check whether the algorithm field is in the enumeration range
func (e PKIEncryptionAlgorithm) validate() error {
	switch e {
	case RSAEncryption:
	case SM2Encryption:
	default:
		return errors.New("unsupported encryption algorithm")
	}

	return nil
}

// generatePkiKeyByAlgorithm 根据加密算法生成密钥对
func (b *backend) generatePkiKeyByAlgorithm(algorithm PKIEncryptionAlgorithm, length int) (string, string, error) {
	var privateKeyPEM, publicKeyPEM string
	switch algorithm {
	case RSAEncryption:
		privateKey, publicKey, err := tools.GenerateRSAKeyPair(length)
		if err != nil {
			return "", "", err
		}
		// 转换为PEM
		publicKeyPEM = string(tools.RSAPublicKeyToPEM(publicKey))
		privateKeyPEM = string(tools.RSAPrivateKeyToPEM(privateKey))
	case SM2Encryption:
		privateKey, err := tools.GenerateSM2KeyPair(b.GetRandomReader())
		if err != nil {
			return "", "", err
		}
		privateKeyPem, err := tools.SM2PrivateKeyToPEM(privateKey)
		if err != nil {
			return "", "", err
		}
		publicKeyPem, err := tools.SM2PublicKeyToPEM(&privateKey.PublicKey)
		if err != nil {
			return "", "", err
		}
		publicKeyPEM = string(publicKeyPem)
		privateKeyPEM = string(privateKeyPem)
	}

	return privateKeyPEM, publicKeyPEM, nil
}

func (b *backend) pathPkiWrite(ctx context.Context, req *logical.Request,
	d *framework.FieldData) (*logical.Response, error) {

	appID := d.Get("app_id").(string)
	if err := b.ValidateAppID(appID); err != nil {
		return nil, err
	}
	name := d.Get("name").(string)
	if err := b.ValidateName(name); err != nil {
		return nil, err
	}
	algorithm := d.Get("algorithm").(string)
	if err := PKIEncryptionAlgorithm(algorithm).validate(); err != nil {
		return nil, err
	}

	length := d.Get("length").(int)
	pubKey := d.Get("pub_key").(string)
	var privKey string
	if pubKey != "" {
		if err := b.verifyPkiPublicKey(PKIEncryptionAlgorithm(algorithm), pubKey); err != nil {
			return nil, err
		}
	} else {
		privKeyPEM, pubKeyPEM, err := b.generatePkiKeyByAlgorithm(PKIEncryptionAlgorithm(algorithm), length)
		if err != nil {
			return nil, err
		}
		privKey = privKeyPEM
		pubKey = pubKeyPEM
	}

	pki := &pkiStorage{
		AppID:     appID,
		Name:      name,
		Algorithm: PKIEncryptionAlgorithm(algorithm),
		Key:       pubKey,
	}
	if err := b.SavePkiStorage(ctx, req.Storage, pki); err != nil {
		return nil, err
	}

	// 如果是服务端创建密钥对，则把私钥返回
	resp := &logical.Response{
		Data: map[string]interface{}{
			"private_key": privKey,
		},
	}

	return resp, nil
}

func (b *backend) pathPkiRead(ctx context.Context, req *logical.Request,
	d *framework.FieldData) (*logical.Response, error) {

	appID := d.Get("app_id").(string)
	if err := b.ValidateAppID(appID); err != nil {
		return nil, err
	}
	name := d.Get("name").(string)
	if err := b.ValidateName(name); err != nil {
		return nil, err
	}
	algorithm := d.Get("algorithm").(string)
	if err := PKIEncryptionAlgorithm(algorithm).validate(); err != nil {
		return nil, err
	}

	s, err := b.GetPkiStorage(ctx, req.Storage, appID, name, algorithm)
	if err != nil {
		return nil, err
	}

	// Generate the response
	resp := &logical.Response{
		Data: map[string]interface{}{
			"pub_key": s.Key,
		},
	}

	return resp, nil

}

func (b *backend) pathPkiDelete(ctx context.Context, req *logical.Request,
	d *framework.FieldData) (*logical.Response, error) {

	appID := d.Get("app_id").(string)
	if err := b.ValidateAppID(appID); err != nil {
		return nil, err
	}
	name := d.Get("name").(string)
	if err := b.ValidateName(name); err != nil {
		return nil, err
	}
	algorithm := d.Get("algorithm").(string)
	if err := PKIEncryptionAlgorithm(algorithm).validate(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("apps/%s/pkis/%s/%s", appID, name, algorithm)

	if err := req.Storage.Delete(ctx, path); err != nil {
		return nil, err
	}

	return nil, nil

}
