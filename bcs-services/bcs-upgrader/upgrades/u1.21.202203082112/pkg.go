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

package u1x21x202203082112

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-upgrader/upgrader"
)

func init() {
	upgrader.RegisterUpgrade("u1.21.202203082112", upgrade)
}

func upgrade(ctx context.Context, helper upgrader.UpgradeHelper) error {
	blog.Infof("start execute u1.21.202203082112")

	h, ok := helper.(*upgrader.Helper)
	if !ok {
		return fmt.Errorf("upgrader helper type unknown")
	}

	handle := NewMigrateHandle(h.Config)
	err := handle.run()
	if err != nil {
		blog.Errorf("[upgrade u1.21.202203082112, migrate data failed, err:  %v", err)
		return err
	}

	return nil
}
