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

package config

import "os"

var (
	BK_SYSTEM_ID         = os.Getenv("BK_SYSTEM_ID")
	BK_APP_CODE          = os.Getenv("BK_APP_CODE")
	BK_APP_SECRET        = os.Getenv("BK_APP_SECRET")
	BK_PAAS_HOST         = os.Getenv("BK_PAAS_HOST")
	REDIS_PASSWORD       = os.Getenv("REDIS_PASSWORD")
	BCS_APIGW_TOKEN      = os.Getenv("BCS_APIGW_TOKEN")
	BCS_APIGW_PUBLIC_KEY = os.Getenv("BCS_APIGW_PUBLIC_KEY")
	BCS_ETCD_HOST        = os.Getenv("bcsEtcdHost")
	BKIAM_GATEWAY_SERVER = os.Getenv("BKIAM_GATEWAY_SERVER")
	MONGO_ADDRESS        = os.Getenv("MONGO_ADDRESS")
	MONGO_USERNAME       = os.Getenv("MONGO_USERNAME")
	MONGO_PASSWORD       = os.Getenv("MONGO_PASSWORD")
)
