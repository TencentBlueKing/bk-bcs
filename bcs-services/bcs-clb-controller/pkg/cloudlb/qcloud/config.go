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

package qcloud

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/cloudlb/qcloud/qcloudif/sdk"
)

// CLBConfig include all config of qcloud clb
type CLBConfig struct {
	ImplementMode         string   `json:"implementmode"`
	BackendMode           string   `json:"backendmode"`
	Region                string   `json:"region"`
	SecretID              string   `json:"secretid"`
	SecretKey             string   `json:"secretkey"`
	ProjectID             int      `json:"projectid"`
	VpcID                 string   `json:"vpcid"`
	CidrIP                string   `json:"cidrip"`
	ExpireTime            int      `json:"expire"`
	SubnetID              string   `json:"subnet"`
	Security              []string `json:"security"`
	MaxTimeout            int      `json:"maxTimeout"`
	WaitPeriodExceedLimit int      `json:"waitPeriodExceedLimit"`
	WaitPeriodLBDealing   int      `json:"waitPeriodDealing"`
}

// NewCLBCfg return a clb cfg instance
func NewCLBCfg() *CLBConfig {
	return &CLBConfig{}
}

// ToJSONString to json string
func (clbCfg *CLBConfig) ToJSONString() string {
	bytes, _ := json.Marshal(clbCfg)
	return string(bytes)
}

// LoadFromEnv load clb config from env
func (clbCfg *CLBConfig) LoadFromEnv() error {
	var err error
	clbCfg.ImplementMode = os.Getenv(ConfigBcsClbImplement)
	if clbCfg.ImplementMode != ConfigBcsClbImplementAPI &&
		clbCfg.ImplementMode != ConfigBcsClbImplementSDK {
		blog.Errorf("implement type [%s] from env %s is invalid", clbCfg.BackendMode, ConfigBcsClbImplement)
		return fmt.Errorf("backend type [%s] from env %s is invalid", clbCfg.BackendMode, ConfigBcsClbImplement)
	}
	clbCfg.BackendMode = os.Getenv(ConfigBcsClbBackendMode)
	if clbCfg.BackendMode != ConfigBcsClbBackendModeENI &&
		clbCfg.BackendMode != ConfigBcsClbBackendModeCVM {
		blog.Errorf("backend type [%s] from env %s is invalid", clbCfg.BackendMode, ConfigBcsClbBackendMode)
		return fmt.Errorf("backend type [%s] from env %s is invalid", clbCfg.BackendMode, ConfigBcsClbBackendMode)
	}
	clbCfg.Region = os.Getenv(ConfigBcsClbRegion)
	if !CheckRegion(clbCfg.Region) {
		blog.Errorf("region [%s] is invalid", clbCfg.Region)
		return fmt.Errorf("region [%s] is invalid", clbCfg.Region)
	}
	clbCfg.SecretID = os.Getenv(ConfigBcsClbSecretID)
	if len(clbCfg.SecretID) == 0 {
		blog.Errorf("secret id cannot be empty")
		return fmt.Errorf("secret id cannot be empty")
	}
	clbCfg.SecretKey = os.Getenv(ConfigBcsClbSecretKey)
	if len(clbCfg.SecretKey) == 0 {
		blog.Errorf("secret key cannot be empty")
		return fmt.Errorf("secret key cannot be empty")
	}

	projectID := os.Getenv(ConfigBcsClbProjectID)
	if len(projectID) == 0 {
		blog.Errorf("project id cannot be empty")
		return fmt.Errorf("project id cannot be empty")
	}
	clbCfg.ProjectID, err = strconv.Atoi(projectID)
	if err != nil {
		blog.Errorf("convert project id %s to int failed, err %s", projectID, err.Error())
		return fmt.Errorf("convert project id %s to int failed, err %s", projectID, err.Error())
	}
	clbCfg.VpcID = os.Getenv(ConfigBcsClbVpcID)
	if len(clbCfg.VpcID) == 0 {
		blog.Errorf("vpc id cannot be empty")
		return fmt.Errorf("vpc id cannot be empty")
	}

	//load expire time
	expireTime := os.Getenv(ConfigBcsClbExpireTime)
	if len(expireTime) != 0 {
		eTime, err := strconv.Atoi(expireTime)
		if err != nil {
			blog.Errorf("expire time %s invalid, set default value 0", expireTime)
			clbCfg.ExpireTime = 0
		} else {
			//expire time: range 30~3600
			if eTime < 30 {
				clbCfg.ExpireTime = 30
			} else if eTime > 3600 {
				clbCfg.ExpireTime = 3600
			} else {
				clbCfg.ExpireTime = eTime
			}
		}
	} else {
		//default 0: means do not set expire time
		clbCfg.ExpireTime = 0
	}

	clbCfg.SubnetID = os.Getenv(ConfigBcsClbSubnet)
	maxTimeout := os.Getenv(ConfigBcsClbMaxTimeout)
	if len(maxTimeout) != 0 {
		timeout, err := strconv.Atoi(maxTimeout)
		if err != nil {
			blog.Errorf("convert max timeout %s to int error, err %s, set default value 180", maxTimeout, err.Error())
			clbCfg.MaxTimeout = DefaultClbMaxTimeout
		} else {
			clbCfg.MaxTimeout = timeout
		}
	} else {
		clbCfg.MaxTimeout = DefaultClbMaxTimeout
	}
	waitPeriodExceedLimit := os.Getenv(ConfigBcsClbWaitPeriodExceedLimit)
	if len(waitPeriodExceedLimit) != 0 {
		period, err := strconv.Atoi(waitPeriodExceedLimit)
		if err != nil {
			blog.Errorf("convert wait period exceed limit %s to int error, err %s, set default value 10",
				waitPeriodExceedLimit, err.Error())
			clbCfg.WaitPeriodExceedLimit = DefaultClbWaitPeriodExceedLimit
		} else {
			clbCfg.WaitPeriodExceedLimit = period
		}
	} else {
		clbCfg.WaitPeriodExceedLimit = DefaultClbWaitPeriodExceedLimit
	}
	waitPeriodLBDealing := os.Getenv(ConfigBcsClbWaitPeriodDealing)
	if len(waitPeriodLBDealing) != 0 {
		period, err := strconv.Atoi(waitPeriodLBDealing)
		if err != nil {
			blog.Errorf("convert wait period lb dealing limit %s to int error, err %s, set default value 3",
				waitPeriodLBDealing, err.Error())
			clbCfg.WaitPeriodLBDealing = DefaultClbWaitPeriodDealing
		} else {
			clbCfg.WaitPeriodLBDealing = period
		}
	} else {
		clbCfg.WaitPeriodLBDealing = DefaultClbWaitPeriodDealing
	}

	blog.Infof("load clb config successfully\n")
	return nil
}

// GenerateSdkConfig generate sdk config
func (clbCfg *CLBConfig) GenerateSdkConfig() *sdk.Config {
	backendType := sdk.ClbBackendTargetTypeCVM
	if clbCfg.BackendMode == ConfigBcsClbBackendModeENI {
		backendType = sdk.ClbBackendTargetTypeENI
	}
	return &sdk.Config{
		BackendType:           backendType,
		Region:                clbCfg.Region,
		ProjectID:             clbCfg.ProjectID,
		SubnetID:              clbCfg.SubnetID,
		VpcID:                 clbCfg.VpcID,
		SecretID:              clbCfg.SecretID,
		SecretKey:             clbCfg.SecretKey,
		MaxTimeout:            clbCfg.MaxTimeout,
		WaitPeriodExceedLimit: clbCfg.WaitPeriodExceedLimit,
		WaitPeriodLBDealing:   clbCfg.WaitPeriodLBDealing,
	}
}
