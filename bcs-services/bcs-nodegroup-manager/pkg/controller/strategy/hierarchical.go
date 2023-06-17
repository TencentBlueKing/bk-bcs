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

package strategy

import (
	"fmt"
	"math"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

// HierarchicalStrategyExecutor hierarchical type strategy executor
type HierarchicalStrategyExecutor struct {
	opt *Options
}

// NewHierarchicalStrategyExecutor return HierarchicalStrategyExecutor
func NewHierarchicalStrategyExecutor(opt *Options) *HierarchicalStrategyExecutor {
	return &HierarchicalStrategyExecutor{
		opt: opt,
	}
}

// IsAbleToScaleDown hierarchicalStrategy do not need to scale down by resource situation, just scale down by task
func (e *HierarchicalStrategyExecutor) IsAbleToScaleDown(strategy *storage.NodeGroupMgrStrategy) (int,
	bool, error) {
	blog.Infof("handle strategy %s", strategy.Name)
	blog.Infof("HierarchicalStrategy do not need to do scale down action")
	return 0, false, nil
}

// IsAbleToScaleUp hierarchicalStrategy need to check if there is scale down task executing
func (e *HierarchicalStrategyExecutor) IsAbleToScaleUp(strategy *storage.NodeGroupMgrStrategy) (int, bool, error) {
	blog.Infof("handle strategy %s", strategy.Name)
	isExecutingTask, err := e.checkIfTaskExecuting(strategy.Name)
	if err != nil {
		return 0, false, fmt.Errorf("check if task executing failed: %s", err.Error())
	}
	if isExecutingTask {
		blog.Infof("strategy %s is executing scale down task, skip scale up", strategy.Name)
		return 0, false, nil
	}
	consumerID := strategy.ElasticNodeGroups[0].ConsumerID
	if consumerID == "" {
		return 0, false, fmt.Errorf("strategy %s consumer id is empty", strategy.Name)
		// Note: 补充通过cluster manager 获取consumer id 逻辑
	}
	deviceGroup, err := e.opt.ResourceManager.GetResourcePoolByCondition(strategy.ResourcePool, consumerID, "", nil)
	if err != nil {
		blog.Errorf("controller got device group by consumer id %s from resource-manager failed, %s",
			consumerID, err.Error())
		return 0, false, fmt.Errorf("get dependent resourcepool %s failed", strategy.ResourcePool)
	}
	total := float64(deviceGroup.InitNum + deviceGroup.IdleNum + deviceGroup.ConsumedNum + deviceGroup.ReturnedNum)
	idleNum := deviceGroup.IdleNum + deviceGroup.InitNum + deviceGroup.ReturnedNum
	warnBuffer := float64(strategy.Strategy.Buffer.High)
	reservedNum := int(math.Ceil(total * warnBuffer / 100))
	blog.Infof("strategy %s, consumer id :%s: total:%d, idleNum:%d, warnBuffer:%d, reservedNum:%d",
		strategy.Name, consumerID, int(total), idleNum, int(warnBuffer), reservedNum)
	if idleNum <= reservedNum {
		//resource is not idle enough
		blog.Infof("the device group of consumer id %s idle resource %d <= reserved %d, elasticNodeGroup don't scaleUp",
			consumerID, idleNum, reservedNum)
		return 0, false, nil
	}
	// check resource pool is idle and stable
	now := time.Now()
	diff := now.Sub(deviceGroup.UpdatedTime)
	if diff.Seconds() < float64(strategy.Strategy.MaxIdleDelay*60) {
		blog.Infof("the device group of consumer id %s is not stable enough for elasticNodeGroup scaleUp, "+
			"now: %.f, target: %d", consumerID, diff.Seconds(), strategy.Strategy.MaxIdleDelay*60)
		return 0, false, nil
	}
	// resource is more than expected, check if controller can scale up
	scaleUpNum := idleNum - reservedNum
	if scaleUpNum < strategy.Strategy.MinScaleUpSize {
		blog.Infof("the device group of consumer id %s idle resource %d is less than MinScaleUpSize %d",
			consumerID, scaleUpNum, strategy.Strategy.MinScaleUpSize)
		return 0, false, nil
	}
	blog.Infof("strategy %s scaleUpNum:%d", strategy.Name, scaleUpNum)
	// feature(DeveloperJim): try to check ScaleUpCoolDown
	// if diff.Seconds() < strategy.Strategy.ScaleUpCoolDown do nothing
	return scaleUpNum, true, nil
}

func (e *HierarchicalStrategyExecutor) checkIfTaskExecuting(strategyName string) (bool, error) {
	tasks, err := e.opt.Storage.ListTasksByStrategy(strategyName, &storage.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("list strategy %s tasks err:%s", strategyName, err.Error())
	}
	for _, task := range tasks {
		if time.Now().Add(5 * time.Minute).After(task.BeginExecuteTime) {
			return true, nil
		}
		if task.IsExecuting() {
			return true, nil
		}
	}
	return false, nil
}
