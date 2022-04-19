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

package sqlstore

import (
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/dchest/uniuri"
)

// RegisterTokenLen token length
const RegisterTokenLen = 128

// GetRegisterToken return the registerToken by clusterId
func GetRegisterToken(clusterID string) *models.BcsRegisterToken {
	token := models.BcsRegisterToken{}
	GCoreDB.Where(&models.BcsRegisterToken{ClusterId: clusterID}).First(&token)
	if token.ID != 0 {
		return &token
	}
	return nil
}

// CreateRegisterToken creates a new registerToken for given clusterId
func CreateRegisterToken(clusterID string) error {
	token := models.BcsRegisterToken{
		ClusterId: clusterID,
		Token:     uniuri.NewLen(RegisterTokenLen),
	}
	err := GCoreDB.Create(&token).Error
	if err == nil {
		return err
	}

	// Transform raw db error messsage
	if strings.HasPrefix(err.Error(), "Error 1062: Duplicate entry") {
		return fmt.Errorf("token already exists")
	}
	return err
}
