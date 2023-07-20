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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
	"strings"
)

// GetCloudZones get cloud region zones
func GetCloudZones(cls *proto.Cluster, cloud *proto.Cloud) ([]*proto.ZoneInfo, error) {
	nodeMgr, err := cloudprovider.GetNodeMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s NodeManager getCloudZones failed, %s", cloud.CloudProvider, err.Error())
		return nil, err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cls.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s list SecurityGroups failed, %s",
			cloud.CloudID, cloud.CloudProvider, err.Error())
		return nil, err
	}
	cmOption.Region = cls.Region

	return nodeMgr.GetZoneList(cmOption)
}

// FormatTaskTime format task time
func FormatTaskTime(t *proto.Task) {
	if t.Start != "" {
		t.Start = utils.TransTimeFormat(t.Start)
	}
	if t.End != "" {
		t.End = utils.TransTimeFormat(t.End)
	}
	for i := range t.Steps {
		if t.Steps[i].Start != "" {
			t.Steps[i].Start = utils.TransTimeFormat(t.Steps[i].Start)
		}
		if t.Steps[i].End != "" {
			t.Steps[i].End = utils.TransTimeFormat(t.Steps[i].End)
		}
	}
}

// Passwd flag
var Passwd = []string{"password", "passwd"}

// HiddenTaskPassword hidden passwd
func HiddenTaskPassword(task *proto.Task) {
	if task != nil && len(task.Steps) > 0 {
		for i := range task.Steps {
			for k := range task.Steps[i].Params {
				if utils.StringInSlice(k,
					[]string{cloudprovider.BkSopsTaskUrlKey.String(), cloudprovider.ShowSopsUrlKey.String()}) {
					continue
				}
				delete(task.Steps[i].Params, k)
			}
		}
	}

	if task != nil && len(task.CommonParams) > 0 {
		for k, v := range task.CommonParams {
			if utils.StringInSlice(strings.ToLower(k), Passwd) || utils.StringContainInSlice(v, Passwd) ||
				utils.StringInSlice(k, []string{cloudprovider.DynamicClusterKubeConfigKey.String()}) {
				delete(task.CommonParams, k)
			}
		}
	}
}
