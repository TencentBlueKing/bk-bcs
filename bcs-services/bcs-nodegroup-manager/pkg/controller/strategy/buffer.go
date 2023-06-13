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

// BufferStrategyExecutor buffer type strategy executor
type BufferStrategyExecutor struct {
	opt *Options
}

// NewBufferStrategyExecutor return BufferStrategyExecutor
func NewBufferStrategyExecutor(opt *Options) *BufferStrategyExecutor {
	return &BufferStrategyExecutor{
		opt: opt,
	}
}

// IsAbleToScaleDown check if is able to scale down, return the scale down num
func (e *BufferStrategyExecutor) IsAbleToScaleDown(strategy *storage.NodeGroupMgrStrategy) (int,
	bool, error) {
	blog.Infof("controller handle strategy %s for ResourcePool %s", strategy.Name, strategy.ResourcePool)
	consumerID := strategy.ReservedNodeGroup.ConsumerID
	if consumerID == "" {
		return 0, false, fmt.Errorf("strategy %s consumer id is empty", strategy.Name)
	}
	// query relative ResourcePool information from resource-manager
	pool, err := e.opt.ResourceManager.GetResourcePoolByCondition(strategy.ResourcePool, consumerID, "", nil)
	if err != nil {
		blog.Errorf("controller got ResourcePool %s from resource-manager failed, %s",
			strategy.ResourcePool, err.Error())
		return 0, false, fmt.Errorf("get dependent resourcepool %s failed", strategy.ResourcePool)
	}
	total := float64(pool.InitNum + pool.IdleNum + pool.ConsumedNum + pool.ReturnedNum)
	idleNum := pool.IdleNum + pool.InitNum + pool.ReturnedNum
	warnBuffer := float64(strategy.Strategy.Buffer.Low)
	reservedNum := int(math.Ceil(total * warnBuffer / 100))
	blog.Infof("strategy %s, resourcePool:%s: total:%d, idleNum:%d, warnBuffer:%d, reservedNum:%d",
		strategy.Name, pool.ID, int(total), idleNum, int(warnBuffer), reservedNum)
	if idleNum >= reservedNum {
		//resource is enough, do nothing
		blog.Infof("ResourcePool %s resource is idle %d >= reserved %d, elasticNodeGroup don't scaleDown",
			pool.ID, idleNum, reservedNum)
		return 0, false, nil
	}
	//buffer resource is not enough, calculate necessary number for scale down
	scaleDownNum := reservedNum - idleNum
	blog.Infof("strategy:%s, scaleDownNum:%d", strategy.Name, scaleDownNum)
	return scaleDownNum, true, nil

}

// IsAbleToScaleUp check if is able to scale up
func (e *BufferStrategyExecutor) IsAbleToScaleUp(strategy *storage.NodeGroupMgrStrategy) (int,
	bool, error) {
	blog.Infof("controller handle strategy %s for ResourcePool %s", strategy.Name, strategy.ResourcePool)
	// query relative ResourcePool information from resource-manager
	consumerID := strategy.ReservedNodeGroup.ConsumerID
	if consumerID == "" {
		return 0, false, fmt.Errorf("strategy %s consumer id is empty", strategy.Name)
	}
	pool, getErr := e.opt.ResourceManager.GetResourcePoolByCondition(strategy.ResourcePool, consumerID, "", nil)
	if getErr != nil {
		blog.Errorf("controller got ResourcePool %s from resource-manager failed, %s",
			strategy.ResourcePool, getErr.Error())
		return 0, false, fmt.Errorf("get dependent resourcepool %s failed", strategy.ResourcePool)
	}
	total := float64(pool.InitNum + pool.IdleNum + pool.ConsumedNum + pool.ReturnedNum)
	idleNum := pool.IdleNum + pool.InitNum + pool.ReturnedNum
	warnBuffer := float64(strategy.Strategy.Buffer.High)
	reservedNum := int(math.Ceil(total * warnBuffer / 100))
	blog.Infof("strategy %s, resourcePool:%s: total:%d, idleNum:%d, warnBuffer:%d, reservedNum:%d",
		strategy.Name, pool.ID, int(total), idleNum, int(warnBuffer), reservedNum)
	if idleNum <= reservedNum {
		//resource is not idle enough
		blog.Infof("ResourcePool %s idle resource %d <= reserved %d, elasticNodeGroup don't scaleUp",
			pool.ID, idleNum, reservedNum)
		return 0, false, nil
	}
	// check resource pool is idle and stable
	now := time.Now()
	diff := now.Sub(pool.UpdatedTime)
	if diff.Seconds() < float64(strategy.Strategy.MaxIdleDelay*60) {
		blog.Infof("ResourcePool %s is not stable enough for elasticNodeGroup scaleUp, now: %.f, target: %d",
			pool.ID, diff.Seconds(), strategy.Strategy.MaxIdleDelay*60)
		return 0, false, nil
	}
	// resource is more than expected, check if controller can scale up
	scaleUpNum := idleNum - reservedNum
	if scaleUpNum < strategy.Strategy.MinScaleUpSize {
		blog.Infof("ResourcePool %s idle resource %d is less than MinScaleUpSize %d",
			pool.ID, scaleUpNum, strategy.Strategy.MinScaleUpSize)
		return 0, false, nil
	}
	blog.Infof("strategy %s scaleUpNum:%d", strategy.Name, scaleUpNum)
	// feature(DeveloperJim): try to check ScaleUpCoolDown
	// if diff.Seconds() < strategy.Strategy.ScaleUpCoolDown do nothing
	return scaleUpNum, true, nil
}
