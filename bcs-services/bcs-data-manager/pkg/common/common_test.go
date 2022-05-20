/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"context"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestGetClusterIDList(t *testing.T) {
	ctx := context.Background()
	cmCli := mock.NewMockCm()
	getter := NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"})
	clusterList, err := getter.GetClusterIDList(ctx, cmCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(clusterList))
}

func TestGetNamespaceList(t *testing.T) {
	ctx := context.Background()
	cmCli := mock.NewMockCm()
	storageCli := mock.NewMockStorage()
	getter := NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"})
	namespaceList, err := getter.GetNamespaceList(ctx, cmCli, storageCli, storageCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 4, len(namespaceList))
}

func TestGetProjectIDList(t *testing.T) {
	ctx := context.Background()
	cmCli := mock.NewMockCm()
	getter := NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"})
	projectList, err := getter.GetProjectIDList(ctx, cmCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(projectList))
}

func TestGetWorkloadList(t *testing.T) {
	ctx := context.Background()
	storageCli := &mock.MockStorage{}
	cmCli := mock.NewMockCm()
	getter := NewGetter(true, []string{"BCS-K8S-15091"})
	namespaceList, err := getter.GetNamespaceList(ctx, cmCli, storageCli, storageCli)
	assert.Nil(t, err)
	workloadList, err := getter.GetK8sWorkloadList(namespaceList, storageCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 6, len(workloadList))
}

func TestFormatTimeIgnoreMin(t *testing.T) {
	original := "2022-03-08 22:00:00"
	local := time.Local
	originalTime, _ := time.ParseInLocation(HourTimeFormat, original, local)
	result := formatTimeIgnoreMin(originalTime)
	assert.Equal(t, "2022-03-08 22:00:00 +0800 CST", result.String())
}

func TestFormatTimeIgnoreSec(t *testing.T) {
	original := "2022-03-08 22:11:00"
	local := time.Local
	originalTime, _ := time.ParseInLocation(MinuteTimeFormat, original, local)
	result := formatTimeIgnoreSec(originalTime)
	assert.Equal(t, "2022-03-08 22:11:00 +0800 CST", result.String())
}

func TestFormatTimeIgnoreHour(t *testing.T) {
	original := "2022-03-08"
	local := time.Local
	originalTime, _ := time.ParseInLocation(DayTimeFormat, original, local)
	result := formatTimeIgnoreHour(originalTime)
	assert.Equal(t, "2022-03-08 00:00:00 +0800 CST", result.String())
}

func TestGetIndex(t *testing.T) {
	origin := "2022-03-08 22:11:00"
	local := time.Local
	originalTime, _ := time.ParseInLocation(MinuteTimeFormat, origin, local)
	day := GetIndex(originalTime, "day")
	hour := GetIndex(originalTime, "hour")
	minute := GetIndex(originalTime, "minute")
	assert.Equal(t, 8, day)
	assert.Equal(t, 22, hour)
	assert.Equal(t, 11, minute)
}
