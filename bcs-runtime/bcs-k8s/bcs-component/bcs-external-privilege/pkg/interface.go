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

package pkg

import (
	"fmt"

	"github.com/TencentBlueKing/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-external-privilege/common"
	"github.com/TencentBlueKing/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-external-privilege/pkg/dbm"
)

// ExternalPrivilege for privilege external system
type ExternalPrivilege interface {
	// DoPri grant privilege to external system
	DoPri(op *common.Option, env *common.DBPrivEnv) error
	// CheckFinalStatus check privilege operation status
	CheckFinalStatus() error
}

// InitClient init external system client
func InitClient(op *common.Option, env *common.DBPrivEnv) (ExternalPrivilege, error) {
	if op == nil {
		return nil, fmt.Errorf("InitClient failed, empty options")
	}
	if len(op.ExternalSysType) == 0 {
		return nil, fmt.Errorf("InitClient failed, empty ExternalSysType")
	}

	switch op.ExternalSysType {
	case common.ExternalSysTypeDBM:
		client, err := dbm.NewDBMClient(op)
		if err != nil {
			return nil, err
		}
		return client, nil
	default:
		return nil, fmt.Errorf("unknown ExternalSysType %s", op.ExternalSysType)
	}
}
