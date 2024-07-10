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

// Package pkg xx
package pkg

import (
	"fmt"

	bcsv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/apis/bkbcs/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/plugin/dbprivilege/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/plugin/dbprivilege/pkg/dbm"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/options"
)

// ExternalPrivilege for privilege external system
type ExternalPrivilege interface {
	// DoPri grant privilege to external system
	DoPri(os *options.Env, dbPrivConfig *bcsv1.DbPrivConfigStatus) error
	// CheckFinalStatus check privilege operation status
	CheckFinalStatus() error
}

// InitClient init external system client
func InitClient(os *options.Env) (ExternalPrivilege, error) {

	if os == nil {
		return nil, fmt.Errorf("DoPri api InitClient failed, empty options")
	}

	if len(os.ExternalSysType) == 0 {
		return nil, fmt.Errorf("DoPri api InitClient failed, empty ExternalSysType")
	}

	switch os.ExternalSysType {
	case common.ExternalSysTypeDBM:
		client, err := dbm.NewDBMClient(os)
		if err != nil {
			return nil, err
		}
		return client, nil
	default:
		return nil, fmt.Errorf("DoPri api unknown ExternalSysType %s", os.ExternalSysType)
	}
}
