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

package config

import "os"

var (
	// BK_APP_CODE app_code
	BK_APP_CODE = os.Getenv("BK_APP_CODE")
	// BK_APP_SECRET app_secret
	BK_APP_SECRET = os.Getenv("BK_APP_SECRET")
	// BK_PAAS_HOST xxx
	BK_PAAS_HOST = os.Getenv("BK_PAAS_HOST")
	// BK_IAM_HOST iam host
	BK_IAM_HOST = os.Getenv("BK_IAM_HOST")
	// BK_IAM_GATEWAY_HOST iam gateway host
	BK_IAM_GATEWAY_HOST = os.Getenv("BK_IAM_GATEWAY_HOST")
	// BK_IAM_EXTERNAL set from global.bkIAM.external； 为空以配置文件为准, 如果设置，以环境变量为准, false代表网关模式
	BK_IAM_EXTERNAL = os.Getenv("BK_IAM_EXTERNAL")
	// REDIS_PASSWORD redis密码
	REDIS_PASSWORD = os.Getenv("REDIS_PASSWORD")
	// BCS_APIGW_TOKEN apigw token
	BCS_APIGW_TOKEN = os.Getenv("BCS_APIGW_TOKEN")
	// BCS_APIGW_PUBLIC_KEY gw公钥
	BCS_APIGW_PUBLIC_KEY = os.Getenv("BCS_APIGW_PUBLIC_KEY")
	// BCS_ETCD_HOST etcd host
	BCS_ETCD_HOST = os.Getenv("bcsEtcdHost")
)
