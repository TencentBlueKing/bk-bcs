/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"time"

	"github.com/robfig/cron"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs"
	contextinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/context"
)

const TIME_LAYOUT = "2006-01-02 15:04:05"

// TimeRange allow user to define a crontab regular
type TimeRange struct {
	// 存放 crontab 语句，如 "* 1-3 * * *"
	Schedule string
	// 指定时区，如 "Asia/Shanghai"
	Zone string
	// 期望节点数
	DesiredNum int
}

// doCron set the minSize of nodegroups according to the rules
func (b *BufferedAutoscaler) doCron(context *contextinternal.Context,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	currentTime time.Time) errors.AutoscalerError {
	bcsProvider, ok := context.CloudProvider.(*bcs.Provider)
	if !ok {
		return errors.NewAutoscalerError(errors.InternalError,
			"Cannot transform cloudprovider to BCSProvider, should run in BCS environment")
	}

	nodegroups := context.CloudProvider.NodeGroups()
	for _, group := range nodegroups {
		ng, ok := group.(*bcs.NodeGroup)
		if !ok {
			return errors.NewAutoscalerError(errors.InternalError,
				"Cannot transform cloudprovider to BCSProvider, should run in BCS environment")
		}
		minSize := ng.MinSize()
		maxSize := ng.MaxSize()
		targetSize, err := ng.TargetSize()
		if err != nil {
			return errors.NewAutoscalerError(errors.ApiCallError,
				"failed to get target size of nodegroup %v: %v", ng.Id(), err)
		}
		timeRanges, err := ng.TimeRanges()
		if err != nil {
			return errors.NewAutoscalerError(errors.ApiCallError,
				"failed to get time ranges of nodegroup %v: %v", ng.Id(), err)
		}

		// get desired num
		desired, err := getDesiredNumForNodeGroupWithTime(ng, currentTime, timeRanges)
		if err != nil {
			return errors.NewAutoscalerError(errors.InternalError,
				"failed to get desiredNum for node group %s in cron mode: %v", ng.Id(), err)
		}
		switch {
		case desired < 0:
			klog.V(4).Infof("CronMode: for nodegroup %v, now is not in the time ranges", ng.Id())
			continue
		case desired > maxSize:
			klog.V(4).Infof("CronMode: for nodegroup %v, desiredNum %d is larger than MaxSize %d",
				ng.Id(), desired, maxSize)
			continue
		case desired == minSize && desired <= targetSize:
			klog.V(4).Infof("CronMode: for nodegroup %v, DesiredNum is %v, MinSize is %v, TargetSize is %v, already satistified",
				ng.Id(), desired, minSize, targetSize)
			continue
		}

		// set minsize
		if minSize != desired {
			err = bcsProvider.NodeGroupCache.SetNodeGroupMinSize(ng.Id(), desired)
			if err != nil {
				return errors.NewAutoscalerError(errors.InternalError,
					"failed to set minSize %d for nodegroup %v: %v", desired, ng.Id(), err)
			}
		}

		// change targetsize
		if targetSize >= desired {
			continue
		}
		info := nodegroupset.ScaleUpInfo{
			Group:       ng,
			CurrentSize: targetSize,
			NewSize:     desired,
			MaxSize:     maxSize,
		}
		err = executeScaleUp(context.AutoscalingContext, clusterStateRegistry, info, "", time.Now())
		if err != nil {
			return errors.NewAutoscalerError(errors.ApiCallError,
				"failed to scale up nodegroup %v to %v: %v", ng.Id(), desired, err)
		}
		klog.V(4).Infof("CronMode: set minsize of %v to %v Successfully", ng.Id(), desired)
	}
	return nil
}

func getDesiredNumForNodeGroupWithTime(ng cloudprovider.NodeGroup,
	currentTime time.Time, timeRanges []*bcs.TimeRange) (int, error) {
	max := -1
	for _, t := range timeRanges {
		_, finalMatch, err := getFinalMatchAndMisMatch(t.Schedule, currentTime, t.Zone)
		if err != nil {
			klog.Errorf("CronMode: failed to get match for timerange \"%v\": %v", t, err)
			return max, err
		}
		if finalMatch == nil {
			continue
		}
		if max < t.DesiredNum {
			max = t.DesiredNum
		}
		klog.V(4).Infof("CronMode: Nodegroup %v, Schedule \"%v\", DesiredNum %v, Max %v",
			ng.Id(), t.Schedule, t.DesiredNum, max)
	}

	return max, nil
}

func getFinalMatchAndMisMatch(schedule string, currentTime time.Time, zone string) (*time.Time, *time.Time, error) {
	currentTime, err := parseTimeWithZone(currentTime, zone)
	if err != nil {
		return nil, nil, err
	}
	sched, err := cron.ParseStandard(schedule)
	if err != nil {
		return nil, nil, err
	}
	lastTime := currentTime.Add(-2 * time.Minute)
	match := lastTime
	misMatch := lastTime
	t := lastTime
	for {
		if !t.After(currentTime) {
			misMatch = t
			t = sched.Next(t)
			continue
		}
		match = t
		break
	}
	if currentTime.Sub(misMatch).Minutes() <= 1 && match.Sub(currentTime).Minutes() <= 1 {
		return &misMatch, &match, nil
	}

	return nil, nil, nil
}

func parseTimeWithZone(currentTime time.Time, zone string) (time.Time, error) {
	local, err := time.LoadLocation(zone)
	if err != nil {
		return time.Time{}, err
	}
	localTime := currentTime.In(local)
	return localTime, nil
}
