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

package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
)

const (
	// ExternalSysTypeDBM external system DBM
	ExternalSysTypeDBM = "DBM"
)

// Option xxx
type Option struct {
	RequestESB        *RequestEsb
	PrivilegeIP       string
	ExternalSysType   string
	ExternalSysConfig string
	DBPrivEnvList     []DBPrivEnv
}

func LoadOption() *Option {
	var ret = &Option{RequestESB: &RequestEsb{}}
	ret.RequestESB.AppCode = os.Getenv("io_tencent_bcs_app_code")
	ret.RequestESB.AppSecret = os.Getenv("io_tencent_bcs_app_secret")
	ret.RequestESB.Operator = os.Getenv("io_tencent_bcs_app_operator")

	ret.PrivilegeIP = os.Getenv("io_tencent_bcs_privilege_ip")
	podIP := os.Getenv("io_tencent_bcs_pod_ip")
	if podIP != "" && podIP != ret.PrivilegeIP {
		ret.PrivilegeIP = fmt.Sprintf("%s,%s", ret.PrivilegeIP, podIP)
	}

	ret.ExternalSysType = os.Getenv("external_sys_type")
	ret.ExternalSysConfig = os.Getenv("external_sys_config")

	envstr := []byte(os.Getenv("io_tencent_bcs_db_privilege_env"))
	err := json.Unmarshal(envstr, &(ret.DBPrivEnvList))
	if err != nil {
		blog.Errorf("Unmarshall json str(%s) to []DBPrivEnv failed: %s\n", string(envstr), err.Error())
		os.Exit(1)
	}

	if ret.RequestESB.AppCode == "" || ret.RequestESB.AppSecret == "" || ret.RequestESB.Operator == "" ||
		len(ret.DBPrivEnvList) == 0 {
		blog.Error("dbPrivEnvList is empty")
		os.Exit(1)
	}

	decryptedAppCode, err := encrypt.DesDecryptFromBase([]byte(ret.RequestESB.AppCode))
	if err != nil {
		blog.Error("unable to decrypt appCode: %s", err.Error())
		os.Exit(1)
	}
	decryptedAppSecret, err := encrypt.DesDecryptFromBase([]byte(ret.RequestESB.AppSecret))
	if err != nil {
		blog.Error("unable to decrypt appSecret: %s", err.Error())
		os.Exit(1)
	}
	decryptedAppOperator, err := encrypt.DesDecryptFromBase([]byte(ret.RequestESB.Operator))
	if err != nil {
		blog.Error("unable to decrypt appOperator: %s", err.Error())
		os.Exit(1)
	}

	ret.RequestESB.AppCode = string(decryptedAppCode)
	ret.RequestESB.AppSecret = string(decryptedAppSecret)
	ret.RequestESB.Operator = string(decryptedAppOperator)

	return ret
}
