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

package lib

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/drivers/mongodb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/drivers/zookeeper"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"
)

func GetZookeeperTank(name string) func() operator.Tank {
	return func() operator.Tank {
		return zookeeper.NewZkTank(apiserver.GetAPIResource().GetZkTankName(name))
	}
}

func GetMongodbTank(name string) func() operator.Tank {
	return func() operator.Tank {
		return mongodb.NewMongodbTank(apiserver.GetAPIResource().GetMongodbTankName(name))
	}
}
