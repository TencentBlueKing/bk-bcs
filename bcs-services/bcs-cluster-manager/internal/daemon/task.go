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

package daemon

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/types"
)

func (d *Daemon) reportMachineryTaskNum(error chan<- error) {
	taskNames, err := d.model.GetTasksFieldDistinct(d.ctx, types.FieldTaskName, bson.D{})
	if err != nil {
		blog.Errorf("reportMachineryTaskNum GetTasksFieldDistinct failed: %v", err)
		error <- err
		return
	}

	for _, taskName := range taskNames {
		if taskName == "" {
			continue
		}

		for _, state := range types.TaskState {
			cond := operator.NewLeafCondition(operator.Eq, operator.M{
				"task_name": taskName,
				"state":     state,
			})

			tasks, errLocal := d.model.ListMachineryTasks(d.ctx, cond, &storeopt.ListOption{All: true})
			if errLocal != nil {
				blog.Errorf("reportMachineryTaskNum ListMachineryTasks failed: %v", errLocal)
				error <- errLocal
				continue
			}

			metrics.ReportMachineryTaskNum(taskName, state, float64(len(tasks)))
		}
	}
}
