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

package cloudprovider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/notify"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/notify/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/notify/server"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tenant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// SendUserNotifyByTemplates send notify template
func SendUserNotifyByTemplates(clsId, groupId, taskId string, isSuccess bool) error {
	var (
		cls   *proto.Cluster
		group *proto.NodeGroup
		task  *proto.Task
		err   error
	)

	cls, err = GetClusterByID(clsId)
	if err != nil {
		return err
	}
	task, err = GetStorageModel().GetTask(context.Background(), taskId)
	if err != nil {
		return err
	}

	if len(groupId) != 0 {
		group, err = GetNodeGroupByID(groupId)
		if err != nil {
			return err
		}
	}

	templates, err := ListProjectNotifyTemplates(cls.ProjectID)
	if err != nil {
		return err
	}
	blog.Infof("task[%s] ListProjectNotifyTemplates success[%v]", taskId, len(templates))

	if len(templates) == 0 {
		return nil
	}

	ctx, err := tenant.WithTenantIdByResourceForContext(context.Background(), tenant.ResourceMetaData{
		ProjectId: cls.ProjectID,
	})
	if err != nil {
		return err
	}

	for i := range templates {
		err = sendNotifyMessage(ctx, cls, group, task, templates[i], isSuccess)
		if err != nil {
			blog.Errorf(err.Error())
			continue
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func sendNotifyMessage(ctx context.Context, cluster *proto.Cluster, group *proto.NodeGroup, task *proto.Task, // nolint
	nt *proto.NotifyTemplate, isSuccess bool) error {
	if !nt.GetEnable() {
		return fmt.Errorf("task[%s] notifyTemplate[%s] not enable", task.TaskID, nt.NotifyTemplateID)
	}

	config := notify.MessageServer{
		Server:      nt.GetConfig().GetServer(),
		DataId:      int64(nt.GetConfig().GetDataId()),
		AccessToken: nt.GetConfig().GetAccessToken(),
	}

	state := func() string {
		if isSuccess {
			return common.TaskStatusSuccess
		}
		return common.TaskStatusFailure
	}()

	extra := business.ExtraParas{
		Render:       true,
		NodeIPList:   strings.Join(task.NodeIPList, ","),
		OperatorTime: utils.TransTimeFormat(task.GetStart()),
		Operator:     task.Creator,
		Result:       state,
	}

	jobType := task.CommonParams[JobTypeKey.String()]
	if len(jobType) == 0 {
		return fmt.Errorf("task[%s] get jobType failed", task.TaskID)
	}

	var (
		err        error
		title      string
		content    string
		dimensions map[string]string
	)

	switch jobType {
	case CreateNodeGroupJob.String():
		if !nt.CreateNodeGroup.GetEnable() {
			return fmt.Errorf("task[%s] notifyTemplate[%s] not enable", task.TaskID, nt.NotifyTemplateID)
		}
		title = business.CreateNodeGroup.GetTitle()
		content = business.CreateNodeGroup.GetContent()

		if nt.GetCreateNodeGroup().GetTitle() != "" {
			title = nt.GetCreateNodeGroup().GetTitle()
		}
		if nt.GetCreateNodeGroup().GetContent() != "" {
			content = nt.GetCreateNodeGroup().GetContent()
		}
		blog.Infof("task[%s] notify content: %s", task.TaskID, content)
		dimensions = business.BuildNodeGroupDimension(cluster.ClusterID, cluster.BusinessID, group.NodeGroupID, state)

	case UpdateNodeGroupDesiredNodeJob.String():
		desired := task.CommonParams[ScalingNodesNumKey.String()]
		scalingNum, _ := strconv.Atoi(desired)
		extra.NodeNum = scalingNum

		if !nt.GetGroupScaleOutNode().GetEnable() {
			return fmt.Errorf("task[%s] notifyTemplate[%s] not enable", task.TaskID, nt.NotifyTemplateID)
		}
		title = business.NodeGroupScaleOutNodes.GetTitle()
		content = business.NodeGroupScaleOutNodes.GetContent()
		if nt.GetGroupScaleOutNode().GetTitle() != "" {
			title = nt.GetGroupScaleOutNode().GetTitle()
		}
		if nt.GetGroupScaleOutNode().GetContent() != "" {
			content = nt.GetGroupScaleOutNode().GetContent()
		}
		blog.Infof("task[%s] notify content: %s", task.TaskID, content)
		dimensions = business.BuildNodeGroupDimension(cluster.ClusterID, cluster.BusinessID, group.NodeGroupID, state)

	case CleanNodeGroupNodesJob.String():
		extra.NodeNum = len(task.NodeIPList)

		if !nt.GetGroupScaleInNode().GetEnable() {
			return fmt.Errorf("task[%s] notifyTemplate[%s] not enable", task.TaskID, nt.NotifyTemplateID)
		}

		title = business.NodeGroupScaleInNodes.GetTitle()
		content = business.NodeGroupScaleInNodes.GetContent()
		if nt.GetGroupScaleInNode().GetTitle() != "" {
			title = nt.GetGroupScaleInNode().GetTitle()
		}
		if nt.GetGroupScaleInNode().GetContent() != "" {
			content = nt.GetGroupScaleInNode().GetContent()
		}
		blog.Infof("task[%s] notify content: %s", task.TaskID, content)
		dimensions = business.BuildNodeGroupDimension(cluster.ClusterID, cluster.BusinessID, group.NodeGroupID, state)
	default:
		return fmt.Errorf("task[%s] not supported jobType[%s]", task.TaskID, jobType)
	}

	// render template
	content, err = business.GetNotifyTemplateContent(cluster, group, extra, content)
	if err != nil {
		return fmt.Errorf("task[%s] getNotifyTemplateContent failed: %v", task.TaskID, err)
	}
	blog.Infof("task[%s] notify render content: %s", task.TaskID, content)

	err = server.SendMessageToServer(ctx, notify.NotifyType(nt.NotifyType),
		config, notify.MessageBody{
			Users:     nt.Receivers,
			Content:   content,
			EventName: title,
			EventBody: content,
			Dimension: dimensions,
		})
	if err != nil {
		blog.Infof("task[%s] sendMessageToServer failed: %v", task.TaskID, err)
		return err
	}

	return nil
}
