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

package worker

import (
	"context"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/mock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

func TestProducer_ClusterProducer(t *testing.T) {
	cmCli := mock.NewMockCmClient()
	queue := mock.NewMockQueue()
	storageCli := mock.NewMockStorage()
	ctx := context.Background()
	pmCli := mock.NewMockPmClient()
	getter := common.NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"}, "stag", pmCli)
	producer := NewProducer(ctx, queue, nil, cmCli, storageCli, storageCli, getter, 100)
	producer.ClusterProducer(types.DimensionMinute)
	assert.NotEqual(t, 0, queue.Length())
	queue.CleanAll()
	producer.ClusterProducer(types.DimensionHour)
	assert.NotEqual(t, 0, queue.Length())
	queue.CleanAll()
	producer.ClusterProducer(types.DimensionDay)
	assert.NotEqual(t, 0, queue.Length())
	queue.CleanAll()
}

func TestProducer_NamespaceProducer(t *testing.T) {
	cmCli := mock.NewMockCmClient()
	queue := mock.NewMockQueue()
	storageCli := mock.NewMockStorage()
	ctx := context.Background()
	pmCli := mock.NewMockPmClient()
	getter := common.NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"}, "stag", pmCli)
	producer := NewProducer(ctx, queue, nil, cmCli, storageCli, storageCli, getter, 100)
	producer.NamespaceProducer(types.DimensionMinute)
	assert.NotEqual(t, 0, queue.Length())
	queue.CleanAll()
	producer.NamespaceProducer(types.DimensionHour)
	assert.NotEqual(t, 0, queue.Length())
	queue.CleanAll()
	producer.NamespaceProducer(types.DimensionDay)
	assert.NotEqual(t, 0, queue.Length())
	queue.CleanAll()
}

func TestProducer_ProjectProducer(t *testing.T) {
	cmCli := mock.NewMockCmClient()
	queue := mock.NewMockQueue()
	storageCli := mock.NewMockStorage()
	ctx := context.Background()
	pmCli := mock.NewMockPmClient()
	getter := common.NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"}, "stag", pmCli)
	producer := NewProducer(ctx, queue, nil, cmCli, storageCli, storageCli, getter, 100)
	producer.ProjectProducer(types.DimensionDay)
	assert.NotEqual(t, 0, queue.Length())
}

func TestProducer_PublicProducer(t *testing.T) {
	cmCli := mock.NewMockCmClient()
	queue := mock.NewMockQueue()
	storageCli := mock.NewMockStorage()
	ctx := context.Background()
	pmCli := mock.NewMockPmClient()
	getter := common.NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"}, "stag", pmCli)
	producer := NewProducer(ctx, queue, nil, cmCli, storageCli, storageCli, getter, 100)
	producer.PublicProducer(types.DimensionDay)
	assert.NotEqual(t, 0, queue.Length())
	fmt.Println(queue.Length())
}

func TestProducer_Run(t *testing.T) {
	ctx := context.Background()
	newcron := cron.New()
	minSpec := "0-59/1 * * * * "
	_, err := newcron.AddFunc(minSpec, func() {
		fmt.Println("cron work")
	})
	assert.Nil(t, err)
	p := &Producer{
		cron: newcron,
		ctx:  ctx,
	}
	p.Run()
	time.Sleep(1 * time.Minute)
	p.Stop()
}

func TestProducer_SendJob(t *testing.T) {
	cmCli := mock.NewMockCmClient()
	queue := mock.NewMockQueue()
	storageCli := mock.NewMockStorage()
	ctx := context.Background()
	pmCli := mock.NewMockPmClient()
	getter := common.NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"}, "stag", pmCli)
	producer := NewProducer(ctx, queue, nil, cmCli, storageCli, storageCli, getter, 100)
	opts := types.JobCommonOpts{
		ObjectType:  types.ClusterType,
		ProjectID:   "testProject",
		ClusterID:   "testCluster",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionMinute,
		CurrentTime: time.Now(),
	}
	err := producer.SendJob(opts)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, queue.Length())
}

func TestProducer_WorkloadProducer(t *testing.T) {
	cmCli := mock.NewMockCmClient()
	queue := mock.NewMockQueue()
	storageCli := mock.NewMockStorage()
	ctx := context.Background()
	pmCli := mock.NewMockPmClient()
	getter := common.NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"}, "stag", pmCli)
	producer := NewProducer(ctx, queue, nil, cmCli, storageCli, storageCli, getter, 100)
	producer.WorkloadProducer(types.DimensionMinute)
	assert.NotEqual(t, 0, queue.Length())
	queue.CleanAll()
	producer.WorkloadProducer(types.DimensionHour)
	assert.NotEqual(t, 0, queue.Length())
	queue.CleanAll()
	producer.WorkloadProducer(types.DimensionDay)
	assert.NotEqual(t, 0, queue.Length())
	queue.CleanAll()
}

func TestProducer_PodAutoscalerProducer(t *testing.T) {
	cmCli := mock.NewMockCmClient()
	queue := mock.NewMockQueue()
	storageCli := mock.NewMockStorage()
	ctx := context.Background()
	pmCli := mock.NewMockPmClient()
	getter := common.NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"}, "stag", pmCli)
	producer := NewProducer(ctx, queue, nil, cmCli, storageCli, storageCli, getter, 100)
	producer.PodAutoscalerProducer(types.DimensionMinute)
	assert.NotEqual(t, 0, queue.Length())
	fmt.Println(queue.Length())
	queue.CleanAll()
	producer.PodAutoscalerProducer(types.DimensionHour)
	assert.NotEqual(t, 0, queue.Length())
	queue.CleanAll()
	producer.PodAutoscalerProducer(types.DimensionDay)
	assert.NotEqual(t, 0, queue.Length())
	queue.CleanAll()
}

func TestProducerInitCronList(t *testing.T) {
	cmCli := mock.NewMockCmClient()
	queue := mock.NewMockQueue()
	storageCli := mock.NewMockStorage()
	ctx := context.Background()
	newcron := cron.New()
	pmCli := mock.NewMockPmClient()
	getter := common.NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"}, "stag", pmCli)
	producer := NewProducer(ctx, queue, newcron, cmCli, storageCli, storageCli, getter, 100)
	err := producer.InitCronList()
	assert.Nil(t, err)
	assert.Equal(t, 11, len(newcron.Entries()))
}
