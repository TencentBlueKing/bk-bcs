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

package util

import (
	"fmt"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
)

// EncryptProjectCred encrypt credential for storage
func EncryptProjectCred(pro *proto.Project) error {
	if pro.Credentials != nil {
		for cloudID, cred := range pro.Credentials {
			if err := EncryptCredential(cred); err != nil {
				return fmt.Errorf("cloud %s credential encrypt failed %s", cloudID, err.Error())
			}
		}
	}
	return nil
}

// DecryptProjectCred encrypt credential for storage
func DecryptProjectCred(pro *proto.Project) error {
	if pro.Credentials != nil {
		for cloudID, cred := range pro.Credentials {
			if err := DecryptCredential(cred); err != nil {
				blog.Errorf("cloud %s credential retrieve failed, %s",
					cloudID, err.Error(),
				)
				return fmt.Errorf("cloud %s credential dencrypt failed %s", cloudID, err.Error())
			}
		}
	}
	return nil
}

func isCredentialValid(cred *proto.Credential) bool {
	if len(cred.ServiceAccountSecret) == 0 && (len(cred.Key) == 0 || len(cred.Secret) == 0) {
		return false
	}
	return true
}

func isTKECredentialValid(cred *proto.Credential) bool {
	if len(cred.Key) == 0 || len(cred.Secret) == 0 {
		return false
	}
	return true
}

// EncryptCredential encrypt credential for storage
func EncryptCredential(cred *proto.Credential) error {
	if cred == nil {
		return fmt.Errorf("lost credential info")
	}
	if !isCredentialValid(cred) {
		return fmt.Errorf("lost serviceAccountSecret, key or secret information")
	}
	if isTKECredentialValid(cred) {
		destKey, err := encrypt.DesEncryptToBase([]byte(cred.Key))
		if err != nil {
			return err
		}
		keyStr := string(destKey)
		destSrt, err := encrypt.DesEncryptToBase([]byte(cred.Secret))
		if err != nil {
			return err
		}
		cred.Secret = string(destSrt)
		cred.Key = keyStr
		return nil
	}
	if len(cred.ServiceAccountSecret) != 0 {
		destSas, err := encrypt.DesEncryptToBase([]byte(cred.ServiceAccountSecret))
		if err != nil {
			return err
		}
		cred.ServiceAccountSecret = string(destSas)
		return nil
	}

	return nil
}

// DecryptCredential encrypt credential for storage
func DecryptCredential(cred *proto.Credential) error {
	if cred == nil {
		return fmt.Errorf("lost credential info")
	}
	if !isCredentialValid(cred) {
		return fmt.Errorf("lost serviceAccountSecret, key or secret information")
	}
	if isTKECredentialValid(cred) {
		destKey, err := encrypt.DesDecryptFromBase([]byte(cred.Key))
		if err != nil {
			return err
		}
		keyStr := string(destKey)
		destSrt, err := encrypt.DesDecryptFromBase([]byte(cred.Secret))
		if err != nil {
			return err
		}
		cred.Secret = string(destSrt)
		cred.Key = keyStr
		return nil
	}
	if len(cred.ServiceAccountSecret) != 0 {
		destSaKey, err := encrypt.DesDecryptFromBase([]byte(cred.ServiceAccountSecret))
		if err != nil {
			return err
		}
		SaKeyStr := string(destSaKey)
		cred.ServiceAccountSecret = SaKeyStr
		return nil
	}
	return nil
}
