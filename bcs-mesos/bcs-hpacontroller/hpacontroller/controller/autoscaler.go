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
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	schedtypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-hpacontroller/hpacontroller/config"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-hpacontroller/hpacontroller/metrics"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-hpacontroller/hpacontroller/reflector"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-hpacontroller/hpacontroller/scaler"
)

const (
	DefaultMinContainerInstance = 1
	DefaultMaxContainerInstance = 2
)

type Autoscaler struct {
	sync.RWMutex
	//hpa controller config
	config *config.Config

	// Reflector watches a specified resource and causes all changes to be reflected in the given store
	store reflector.Reflector

	// MetricsController collect external metrics or taskgroup resource metrics
	resourceMetrics metrics.MetricsController

	// MetricsController collect external metrics or taskgroup resource metrics
	externalMetrics metrics.MetricsController

	//ScalerProcess can scale up/down target ref deployment/application instance
	scalerController scaler.ScalerProcess

	//hpa autoscaler work queue, key = BcsAutoscaler.GetUuid()
	workQueue map[string]*commtypes.BcsAutoscaler
}

func NewAutoscaler(conf *config.Config, store reflector.Reflector, resourcesMetrics metrics.MetricsController,
	externalMetrics metrics.MetricsController, scalerController scaler.ScalerProcess) *Autoscaler {

	auto := &Autoscaler{
		config:           conf,
		store:            store,
		resourceMetrics:  resourcesMetrics,
		externalMetrics:  externalMetrics,
		scalerController: scalerController,
		workQueue:        make(map[string]*commtypes.BcsAutoscaler),
	}

	return auto
}

//start autoscaler controller asynchronous worker
func (auto *Autoscaler) Start() error {

	//ticker list zk autoscalers and sync these autoscalers to workqueue
	go auto.tickerSyncAutoscalerQueue()

	//ticker handler autoscaler
	go auto.tickerHandlerAutoscaler()
	return nil
}

//ticker list zk autoscalers and sync these autoscalers to workqueue
func (auto *Autoscaler) tickerSyncAutoscalerQueue() {
	ticker := time.NewTicker(time.Second * time.Duration(auto.config.MetricsSyncPeriod))
	defer ticker.Stop()

	for {

		select {
		case <-ticker.C:
			blog.V(3).Infof("ticker sync autoscaler queue start...")
		}

		autoscalers, err := auto.store.ListAutoscalers()
		if err != nil {
			blog.Errorf("auto list Autoscalers error %s", err.Error())
			continue
		}

		auto.Lock()
		currentQueue := make(map[string]*commtypes.BcsAutoscaler, len(auto.workQueue))
		for k, scaler := range auto.workQueue {
			currentQueue[k] = scaler
		}

		for _, scaler := range autoscalers {
			blog.V(3).Infof("ticker sync autoscaler %s start...", scaler.GetUuid())

			//check scaler is invalid
			err := auto.checkScalerIsValid(scaler)
			if err != nil {
				blog.Errorf("check scaler failed, error %s", err.Error())
			}

			// if zk scaler exist, then delete currentQueue
			delete(currentQueue, scaler.GetUuid())

			//if scaler is already in the workQueue, then continue
			_, ok := auto.workQueue[scaler.GetUuid()]
			if ok {
				//blog.V(3).Infof("ticker sync scaler %s already exists", scaler.GetUuid())
				continue
			}

			// init scaler status
			if scaler.Status == nil || len(scaler.Status.CurrentMetrics) == 0 {
				blog.Infof("init autoscaler(%s:%s)", scaler.NameSpace, scaler.Name)
				scaler.InitAutoscalerStatus()
			}

			by, _ := json.Marshal(scaler)
			blog.Infof("store scaler %s", string(by))

			//store scaler in zk
			err = auto.store.StoreAutoscaler(scaler)
			if err != nil {
				blog.Errorf("store autoscaler %s error %s", scaler.GetUuid(), err.Error())
				continue
			}

			//start collect autoscaler ref metrics
			blog.Infof("start collect scaler %s metrics", scaler.GetUuid())
			auto.resourceMetrics.StartScalerMetrics(scaler)

			//add scaler into workqueue
			blog.Infof("add scaler %s into workqueue", scaler.GetUuid())
			auto.workQueue[scaler.GetUuid()] = scaler
		}

		//delete invalid scaler in workqueue
		for k, scaler := range currentQueue {
			blog.Infof("delete scaler %s", scaler.GetUuid())
			auto.resourceMetrics.StopScalerMetrics(scaler)
			delete(auto.workQueue, k)
		}
		auto.Unlock()

		blog.V(3).Infof("ticker sync autoscaler queue done")
	}
}

func (auto *Autoscaler) tickerHandlerAutoscaler() {
	ticker := time.NewTicker(time.Second * time.Duration(auto.config.MetricsSyncPeriod))
	defer ticker.Stop()

	for {

		select {
		case <-ticker.C:
			blog.V(3).Infof("ticker handler autoscaler queue start...")
		}

		for uuid, scaler := range auto.workQueue {
			blog.Infof("ticker handler scaler %s start...", scaler.GetUuid())
			targetRef := scaler.Spec.ScaleTargetRef

			var application *schedtypes.Application
			var err error

			//get scaler target ref object
			if targetRef.Kind == commtypes.AutoscalerTargetRefApplication {
				application, err = auto.store.FetchApplicationInfo(targetRef.Namespace, targetRef.Name)
				if err != nil {
					blog.Errorf("fetch scaler %s target ref application(%s:%s) error %s", uuid,
						targetRef.Namespace, targetRef.Name, err.Error())
					continue
				}
			} else if targetRef.Kind == commtypes.AutoscalerTargetRefDeployment {
				deploy, err := auto.store.FetchDeploymentInfo(targetRef.Namespace, targetRef.Name)
				if err != nil {
					blog.Errorf("fetch scaler %s target ref deployment(%s:%s) error %s", uuid,
						targetRef.Namespace, targetRef.Name, err.Error())
					continue
				}

				if deploy.Status != schedtypes.DEPLOYMENT_STATUS_RUNNING {
					blog.Errorf("scaler %s targetref deployment(%s:%s) status %s, and can't scale it",
						uuid, targetRef.Namespace, targetRef.Name, deploy.Status)
					scaler.Status.TargetRefStatus = deploy.Status
					err = auto.store.UpdateAutoscaler(scaler)
					if err != nil {
						blog.Errorf("store scaler %s error %s", scaler.GetUuid(), err.Error())
					}
					continue
				}

				application, err = auto.store.FetchApplicationInfo(targetRef.Namespace, deploy.Application.ApplicationName)
				if err != nil {
					blog.Errorf("fetch scaler %s target ref application(%s:%s) error %s", uuid,
						targetRef.Namespace, deploy.Application.ApplicationName, err.Error())
					continue
				}
			} else {
				blog.Errorf("scaler %s targetRef kind %s is invalid", scaler.GetUuid(), targetRef.Kind)
				continue
			}

			//if app.status is not running or abnormal, then can't scale it
			if application.Status != schedtypes.APP_STATUS_RUNNING && application.Status != schedtypes.APP_STATUS_ABNORMAL {
				blog.Errorf("scaler %s targetref application(%s:%s) status %s, and can't scale it",
					uuid, targetRef.Namespace, targetRef.Name, application.Status)
				scaler.Status.TargetRefStatus = application.Status
				err = auto.store.UpdateAutoscaler(scaler)
				if err != nil {
					blog.Errorf("store scaler %s error %s", scaler.GetUuid(), err.Error())
				}
				continue
			}

			//update scaler status info
			scaler.Status.TargetRefStatus = application.Status
			scaler.Status.CurrentInstance = uint(application.Instances)
			if scaler.Status.DesiredInstance == 0 {
				scaler.Status.DesiredInstance = uint(application.Instances)
			}
			if scaler.Spec.MinInstance == 0 {
				scaler.Spec.MinInstance = DefaultMinContainerInstance
			}
			if scaler.Spec.MaxInstance == 0 {
				scaler.Spec.MaxInstance = DefaultMaxContainerInstance
			}

			blog.Infof("scaler %s status is ok, and update current metrics", scaler.GetUuid())
			//update scaler metric current's value
			err = auto.updateScalerCurrentMetrics(scaler)
			if err != nil {
				blog.Errorf("update scaler %s current metrics error %s", uuid, err.Error())
				continue
			}
			err = auto.store.UpdateAutoscaler(scaler)
			if err != nil {
				blog.Errorf("store scaler %s error %s", scaler.GetUuid(), err.Error())
				continue
			}

			//compute scaler desired Instance
			desiredInstance, operator, err := auto.computeScalerDesiredInstance(scaler)
			if err != nil {
				blog.Errorf("compute scaler %s desired instance error %s", scaler.GetUuid(), err.Error())
				continue
			}

			if operator == commtypes.AutoscalerOperatorNone {
				blog.V(3).Infof("scaler %s autoscaler operator is %s", scaler.GetUuid(), operator)
				continue
			}

			// The period for which autoscaler will look backwards and
			// not scale up below any recommendation it made during that period
			if operator == commtypes.AutoscalerOperatorScaleUp &&
				scaler.Status.LastScaleOPeratorType == commtypes.AutoscalerOperatorScaleUp &&
				(time.Now().Unix()-scaler.Status.LastScaleTime.Unix()) < auto.config.UpscaleStabilization {
				blog.Infof("scaler %s at time %s last operator %s, then continue", scaler.GetUuid(),
					scaler.Status.LastScaleTime.Format("2006-01-02 15:04:05"), scaler.Status.LastScaleOPeratorType)
				continue
			}

			// The period for which autoscaler will look backwards and
			// not scale down below any recommendation it made during that period
			if operator == commtypes.AutoscalerOperatorScaleDown &&
				(time.Now().Unix()-scaler.Status.LastScaleTime.Unix()) < auto.config.DownscaleStabilization {
				blog.Infof("scaler %s at time %s last operator %s, then continue", scaler.GetUuid(),
					scaler.Status.LastScaleTime.Format("2006-01-02 15:04:05"), scaler.Status.LastScaleOPeratorType)
				continue
			}

			blog.Infof("scaler %s %s target ref(%s:%s) to instance %d", scaler.GetUuid(),
				operator, targetRef.Namespace, targetRef.Name, desiredInstance)
			switch operator {
			case commtypes.AutoscalerOperatorScaleUp:
				err = auto.scaleScalerTargetRef(desiredInstance, scaler)

			case commtypes.AutoscalerOperatorScaleDown:
				err = auto.scaleScalerTargetRef(desiredInstance, scaler)
			}

			if err != nil {
				blog.Errorf("scale operator %s scaler %s error %s", operator, scaler.GetUuid(), err.Error())
				continue
			}

			blog.Infof("autoscale scaler %s operator %s desired instance %d success", scaler.GetUuid(), operator, desiredInstance)
			//update scaler status info
			scaler.Status.LastScaleOPeratorType = operator
			scaler.Status.DesiredInstance = desiredInstance
			scaler.Status.LastScaleTime = time.Now()
			scaler.Status.ScaleNumber++
			err = auto.store.UpdateAutoscaler(scaler)
			if err != nil {
				blog.Errorf("store scaler %s error %s", scaler.GetUuid(), err.Error())
			}
		}
	}
}

//scale scaler target ref deployment, application
func (auto *Autoscaler) scaleScalerTargetRef(desiredInstance uint, scaler *commtypes.BcsAutoscaler) error {
	targetRef := scaler.Spec.ScaleTargetRef

	var err error
	switch targetRef.Kind {
	case commtypes.AutoscalerTargetRefDeployment:
		err = auto.scalerController.ScaleDeployment(targetRef.Namespace, targetRef.Name, desiredInstance)

	case commtypes.AutoscalerTargetRefApplication:
		err = auto.scalerController.ScaleApplication(targetRef.Namespace, targetRef.Name, desiredInstance)
	}

	return err
}

func (auto *Autoscaler) checkScalerIsValid(scaler *commtypes.BcsAutoscaler) error {
	//check the target ref kind
	scaler.Spec.ScaleTargetRef.Kind = strings.ToLower(scaler.Spec.ScaleTargetRef.Kind)
	if scaler.Spec.ScaleTargetRef.Kind != commtypes.AutoscalerTargetRefApplication &&
		scaler.Spec.ScaleTargetRef.Kind != commtypes.AutoscalerTargetRefDeployment {
		return fmt.Errorf("scaler %s TargetRef.Kind %s is invalid", scaler.GetUuid(), scaler.Spec.ScaleTargetRef.Kind)
	}

	//check
	for _, metrics := range scaler.Spec.MetricsTarget {
		if metrics.Type != commtypes.ResourceMetricSourceType && metrics.Type != commtypes.TaskgroupsMetricSourceType &&
			metrics.Type != commtypes.ExternalMetricSourceType {
			return fmt.Errorf("scaler %s metrics %s type %s is invalid", scaler.GetUuid(), metrics.Name, metrics.Type)
		}
	}

	return nil
}
