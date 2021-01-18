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

package backend

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

func (b *backend) CommitImage(taskgroup, image, url string) (*types.BcsMessage, error) {

	taskGroup, err := b.store.FetchTaskGroup(taskgroup)
	if err != nil {
		blog.Errorf("CommitImage: fetch taskgroup(%s) error %s", taskgroup, err.Error())
		return nil, err
	}

	if taskGroup.Status != types.TASKGROUP_STATUS_RUNNING {
		blog.Warnf("taskgroup(%s) cannot commit image under status(%s)", taskgroup, taskGroup.Status)
		return nil, fmt.Errorf("taskgroup(%s) cannot commit image under status(%s)", taskgroup, taskGroup.Status)
	}

	msg := &types.Msg_CommitTask{}

	exist := false
	for _, task := range taskGroup.Taskgroup {

		if task.Image == image {
			exist = true
			msg.Tasks = append(msg.Tasks, &types.CommitTask{
				TaskId: &task.ID,
				Image:  &url,
			})
		}
	}

	if exist == false {
		blog.Errorf("image(%s) not found in taskgroup(%s)", image, taskgroup)
		return nil, fmt.Errorf("image(%s) not found in taskgroup(%s)", image, taskgroup)
	}

	bcsMsg := &types.BcsMessage{
		Type:       types.Msg_COMMIT_TASK.Enum(),
		CommitTask: msg,
	}

	return b.sched.SendBcsMessage(taskGroup, bcsMsg)
}
