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

package controller

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
)

const (
	DefaultValidMetricsTimeout = 60 //seconds
)

//update scaler current metrics
func (auto *Autoscaler) updateScalerCurrentMetrics(scaler *commtypes.BcsAutoscaler) error {
	refKind := scaler.Spec.ScaleTargetRef.Kind
	refNs := scaler.Spec.ScaleTargetRef.Namespace
	refName := scaler.Spec.ScaleTargetRef.Name

	for _, current := range scaler.Status.CurrentMetrics {

		switch current.Type {
		case commtypes.ResourceMetricSourceType:
			metrics, err := auto.resourceMetrics.GetResourceMetric(current.Name, scaler.GetUuid())
			if err != nil {
				blog.Errorf("scaler %s ref(%s:%s:%s) get resources %s metrics error %s", scaler.GetUuid(),
					refKind, refNs, refName, current.Name, err.Error())
				break
			}

			var totalValue float32
			var num float32
			for k, metric := range metrics {
				if metric.Value == 0 {
					blog.Warnf("scaler %s taskgroup %s metric value is zero", scaler.GetUuid(), k)
					continue
				}
				num++
				totalValue += metric.Value
			}

			if num > 0 {
				value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", totalValue/num), 32)
				current.Current.AverageUtilization = float32(value)
				current.Timestamp = time.Now()
			}

		case commtypes.ExternalMetricSourceType:
			//todo

		default:
			blog.Errorf("scaler %s metrics %s type %s is invalid", scaler.GetUuid(), current.Name, current.Type)
		}
	}

	return nil
}

//compute scaler DesiredInstance, and return it
func (auto *Autoscaler) computeScalerDesiredInstance(scaler *commtypes.BcsAutoscaler) (uint, commtypes.AutoscalerOperatorType, error) {
	var (
		describedInstance uint
		scaleUpNumber     int
		scaleDownNumber   int
		scalerOperator    commtypes.AutoscalerOperatorType = commtypes.AutoscalerOperatorNone
	)

	currentInstance := float32(scaler.Status.CurrentInstance)
	//if current instance > max instance
	if uint(currentInstance) > scaler.Spec.MaxInstance {
		return scaler.Spec.MaxInstance, commtypes.AutoscalerOperatorScaleDown, nil
	}

	//if current instance < min instance
	if uint(currentInstance) < scaler.Spec.MinInstance {
		return scaler.Spec.MinInstance, commtypes.AutoscalerOperatorScaleUp, nil
	}

	for _, target := range scaler.Spec.MetricsTarget {
		current, err := scaler.GetSpecifyCurrentMetrics(target.Type, target.Name)
		if err != nil {
			blog.Errorf("scaler %s get current metric(%s:%s) error %s",
				scaler.GetUuid(), target.Type, target.Name, err.Error())
			continue
		}

		var tolerance float32
		valueIsOk := true
		switch target.Target.Type {
		case commtypes.AutoscalerMetricAverageUtilization:
			if current.Current.AverageUtilization == 0 {
				blog.Errorf("scaler %s metrics %s current value is zero", scaler.GetUuid(), current.Name)
				valueIsOk = false
				break
			}
			if time.Now().Unix()-current.Timestamp.Unix() > DefaultValidMetricsTimeout {
				blog.Errorf("scaler %s metrics %s current metric timestamp %s is timeout",
					scaler.GetUuid(), current.Name, current.Timestamp.Format("2006-01-02 15:04:05"))
				valueIsOk = false
				break
			}

			tolerance = float32(current.Current.AverageUtilization) / float32(target.Target.AverageUtilization)
			blog.Infof("scaler %s metric %s current %.2f target %.2f tolerance %.2f", scaler.GetUuid(),
				current.Name, current.Current.AverageUtilization, target.Target.AverageUtilization, tolerance)

		case commtypes.AutoscalerMetricTargetAverageValue:
			if current.Current.AverageValue == 0 {
				blog.Errorf("scaler %s metrics %s current value is zero", scaler.GetUuid(), current.Name)
				valueIsOk = false
				break
			}
			if time.Now().Unix()-current.Timestamp.Unix() > DefaultValidMetricsTimeout {
				blog.Errorf("scaler %s metrics %s current metric timestamp %s is timeout",
					scaler.GetUuid(), current.Name, current.Timestamp.Format("2006-01-02 15:04:05"))
				valueIsOk = false
				break
			}

			tolerance = float32(current.Current.AverageValue) / float32(target.Target.AverageValue)
			blog.Infof("scaler %s metric %s current %.2f target %.2f tolerance %.2f", scaler.GetUuid(),
				current.Name, current.Current.AverageValue, target.Target.AverageValue, tolerance)

		case commtypes.AutoscalerMetricTargetValue:
			if current.Current.Value == 0 {
				blog.Errorf("scaler %s metrics %s current value is zero", scaler.GetUuid(), current.Name)
				valueIsOk = false
				break
			}
			if time.Now().Unix()-current.Timestamp.Unix() > DefaultValidMetricsTimeout {
				blog.Errorf("scaler %s metrics %s current metric timestamp %s is timeout",
					scaler.GetUuid(), current.Name, current.Timestamp.Format("2006-01-02 15:04:05"))
				valueIsOk = false
				break
			}

			tolerance = float32(current.Current.Value) / float32(target.Target.Value)
			blog.Infof("scaler %s metric %s current %.2f target %.2f tolerance %.2f", scaler.GetUuid(),
				current.Name, current.Current.Value, target.Target.Value, tolerance)

		default:
			blog.Errorf("scaler %s metric %s type %s is invalid", scaler.GetUuid(), target.Name, target.Target.Type)
			valueIsOk = false
		}

		// if current metrics value is invalid, then continue
		if !valueIsOk {
			continue
		}

		// The minimum change (from 1.0) in the desired-to-actual metrics ratio
		// for the autoscaler to consider scaling
		// if metric current/metric target > AutoscalerTolerance or < AutoscalerTolerance, then scale
		if tolerance <= (1+auto.config.AutoscalerTolerance) && tolerance >= (1-auto.config.AutoscalerTolerance) {
			blog.V(3).Infof("scaler %s metric(%s:%s) tolerance %.2f, and don't scale it",
				scaler.GetUuid(), current.Type, current.Name, tolerance)
			continue
		}

		//scale up
		if tolerance > (1 + auto.config.AutoscalerTolerance) {
			ceil := uint(math.Ceil(float64(currentInstance * tolerance)))

			//if described instance == current instance, then don't scale it
			if ceil == scaler.Status.CurrentInstance {
				continue
			}
			//if cuttent instance == max instance, then don't scale it
			if scaler.Status.CurrentInstance == scaler.Spec.MaxInstance {
				continue
			}

			scaleUpNumber++
			if ceil > describedInstance {
				describedInstance = ceil
			}
			if describedInstance > scaler.Spec.MaxInstance {
				describedInstance = scaler.Spec.MaxInstance
			}

			blog.Infof("scaler %s metrics %s current instance %f tolerance %.2f described instance %d", scaler.GetUuid(),
				current.Name, currentInstance, tolerance, describedInstance)
		}

		//if scale up number>0, then don't need scale down it
		if scaleUpNumber > 0 {
			continue
		}

		//scale down
		if tolerance < (1 - auto.config.AutoscalerTolerance) {
			ceil := uint(math.Ceil(float64(currentInstance * tolerance)))

			//if described instance == current instance, then don't scale it
			if ceil == scaler.Status.CurrentInstance {
				continue
			}
			//if cuttent instance == min instance, then don't scale it
			if scaler.Status.CurrentInstance == scaler.Spec.MinInstance {
				continue
			}

			scaleDownNumber++
			if describedInstance == 0 {
				describedInstance = ceil
			}
			if describedInstance < ceil {
				describedInstance = ceil
			}
			if describedInstance < scaler.Spec.MinInstance {
				describedInstance = scaler.Spec.MinInstance
			}

			blog.Infof("scaler %s metrics %s tolerance %.2f described instance %d", scaler.GetUuid(),
				current.Name, tolerance, describedInstance)
		}
	}

	//if scale up number>0, then scale up it
	if scaleUpNumber > 0 {
		scalerOperator = commtypes.AutoscalerOperatorScaleUp
		return describedInstance, scalerOperator, nil
	}

	//if scale down number==len(scaler.Spec.MetricsTarget), then scale down it
	if scaleDownNumber == len(scaler.Spec.MetricsTarget) {
		scalerOperator = commtypes.AutoscalerOperatorScaleDown
		return describedInstance, scalerOperator, nil
	}

	//finally don't scale it
	return scaler.Status.CurrentInstance, scalerOperator, nil
}
