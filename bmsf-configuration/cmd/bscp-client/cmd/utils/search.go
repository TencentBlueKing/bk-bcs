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

package utils

import (
	"bk-bscp/cmd/bscp-client/service"
	"bk-bscp/internal/protocol/common"
	"context"
	"fmt"
)

//GetBusinessAndApp fast way to get Business & App by their names
func GetBusinessAndApp(operator *service.AccessOperator, businessName, appName string) (*common.Business, *common.App, error) {
	//check business first
	business, err := operator.GetBusiness(context.TODO(), businessName)
	if err != nil {
		return nil, nil, err
	}
	if business == nil {
		return nil, nil, fmt.Errorf("Business %s Resource Not Found", businessName)
	}
	app, err := operator.GetAppByID(context.TODO(), business.Bid, appName)
	if err != nil {
		return nil, nil, err
	}
	if app == nil {
		return nil, nil, fmt.Errorf("App %s resource Not Found", appName)
	}
	return business, app, nil
}

//GetConfigSet fast way
func GetConfigSet(operator *service.AccessOperator, cfgset *common.ConfigSet) (*common.ConfigSet, error) {
	cfgset, err := operator.QueryConfigSet(context.TODO(), cfgset)
	if err != nil {
		return nil, err
	}
	if cfgset == nil {
		return nil, fmt.Errorf("Cfgset %s resource Not Found", cfgset)
	}
	return cfgset, nil
}
