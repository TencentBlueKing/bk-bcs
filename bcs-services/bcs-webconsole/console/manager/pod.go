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

package manager

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

// WritePodData 写入用户pod数据
func (m *manager) WritePodData(data *types.UserPodData) {
	m.RWMutex.Lock()
	defer m.RWMutex.Unlock()
	m.PodMap[data.SessionID+"_"+data.ProjectID+"_"+data.ClustersID] = *data
	// TODO 应该用username 代替 sessionID，
}

// ReadPodData 读取用户pod数据
// TODO 应该用username 代替 sessionID
func (m *manager) ReadPodData(sessionID, projectID, clustersID string) (*types.UserPodData, bool) {
	m.RWMutex.RLock()
	defer m.RWMutex.RUnlock()

	data, ok := m.PodMap[sessionID+"_"+projectID+"_"+clustersID]
	return &data, ok

}
