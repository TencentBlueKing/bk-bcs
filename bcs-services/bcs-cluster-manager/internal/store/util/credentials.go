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

// Package util xxx
package util

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/encryptv2" // nolint

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
)

// EncryptCredentialData encrypt credential for storage
func EncryptCredentialData(en encryptv2.Cryptor, cred *proto.Credential) error {
	if cred == nil {
		return fmt.Errorf("lost credential info")
	}

	// tencentCloud
	if len(cred.Key) > 0 {
		destKey, err := encrypt.Encrypt(en, cred.Key)
		if err != nil {
			return err
		}
		cred.Key = destKey
	}
	if len(cred.Secret) > 0 {
		destSrt, err := encrypt.Encrypt(en, cred.Secret)
		if err != nil {
			return err
		}
		cred.Secret = destSrt
	}

	// gke
	if len(cred.ServiceAccountSecret) != 0 {
		destSas, err := encrypt.Encrypt(en, cred.ServiceAccountSecret)
		if err != nil {
			return err
		}
		cred.ServiceAccountSecret = destSas
	}

	return nil
}

// DecryptCredentialData decrypt credential for storage
func DecryptCredentialData(de encryptv2.Cryptor, cred *proto.Credential) error {
	if cred == nil {
		return fmt.Errorf("lost credential info")
	}

	// tencentCloud
	if len(cred.Key) > 0 {
		destKey, err := encrypt.Decrypt(de, cred.Key)
		if err != nil {
			return err
		}
		cred.Key = destKey
	}
	if len(cred.Secret) > 0 {
		destSrt, err := encrypt.Decrypt(de, cred.Secret)
		if err != nil {
			return err
		}
		cred.Secret = destSrt
	}

	// gke
	if len(cred.ServiceAccountSecret) != 0 {
		destSaKey, err := encrypt.Decrypt(de, cred.ServiceAccountSecret)
		if err != nil {
			return err
		}
		cred.ServiceAccountSecret = destSaKey
	}

	return nil
}

// EncryptCloudAccountData encrypt cloud account for storage
func EncryptCloudAccountData(en encryptv2.Cryptor, account *proto.Account) error {
	if account == nil {
		return fmt.Errorf("lost account info")
	}

	// encrypt cloud account by different cloud
	// tencentCloud
	if len(account.SecretKey) > 0 {
		destKey, err := encrypt.Encrypt(en, account.SecretKey)
		if err != nil {
			return err
		}
		account.SecretKey = destKey
	}

	if len(account.SecretID) > 0 {
		destID, err := encrypt.Encrypt(en, account.SecretID)
		if err != nil {
			return err
		}
		account.SecretID = destID
	}

	// azure
	if len(account.ClientSecret) > 0 {
		destSecret, err := encrypt.Encrypt(en, account.ClientSecret)
		if err != nil {
			return err
		}
		account.ClientSecret = destSecret
	}

	// gke
	if len(account.ServiceAccountSecret) > 0 {
		destSas, err := encrypt.Encrypt(en, account.ServiceAccountSecret)
		if err != nil {
			return err
		}
		account.ServiceAccountSecret = destSas
	}

	return nil
}

// DecryptCloudAccountData decrypt account for storage
func DecryptCloudAccountData(de encryptv2.Cryptor, account *proto.Account) error {
	if account == nil {
		return fmt.Errorf("lost account info")
	}

	// decrypt cloud account by different cloud
	// tencentCloud
	if len(account.SecretKey) > 0 {
		destKey, err := encrypt.Decrypt(de, account.SecretKey)
		if err != nil {
			return err
		}
		account.SecretKey = destKey
	}

	if len(account.SecretID) > 0 {
		destID, err := encrypt.Decrypt(de, account.SecretID)
		if err != nil {
			return err
		}
		account.SecretID = destID
	}

	// azure
	if len(account.ClientSecret) > 0 {
		destSecret, err := encrypt.Decrypt(de, account.ClientSecret)
		if err != nil {
			return err
		}
		account.ClientSecret = destSecret
	}

	// gke
	if len(account.ServiceAccountSecret) > 0 {
		destSas, err := encrypt.Decrypt(de, account.ServiceAccountSecret)
		if err != nil {
			return err
		}
		account.ServiceAccountSecret = destSas
	}

	return nil
}
