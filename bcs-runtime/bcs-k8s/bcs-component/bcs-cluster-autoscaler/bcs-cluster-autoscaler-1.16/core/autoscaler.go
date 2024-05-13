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

package core

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/expander/factory"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/backoff"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	cloudBuilder "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/builder"
	estimatorinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/estimator"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/scalingconfig"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/util"
)

// AutoscalerOptions is the whole set of options for configuring an autoscaler
type AutoscalerOptions struct {
	scalingconfig.Options
	KubeClient             kube_client.Interface
	EventsKubeClient       kube_client.Interface
	AutoscalingKubeClients *context.AutoscalingKubeClients
	CloudProvider          cloudprovider.CloudProvider
	PredicateChecker       *simulator.PredicateChecker
	ExpanderStrategy       expander.Strategy
	EstimatorBuilder       estimatorinternal.ExtendedEstimatorBuilder
	Processors             *ca_processors.AutoscalingProcessors
	Backoff                backoff.Backoff
}

// NewAutoscaler creates an autoscaler of an appropriate type according to the parameters
func NewAutoscaler(opts AutoscalerOptions) (core.Autoscaler, errors.AutoscalerError) {
	err := initializeDefaultOptions(&opts)
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	return NewBufferedAutoscaler(
		opts.Options,
		opts.PredicateChecker,
		opts.AutoscalingKubeClients,
		opts.Processors,
		opts.CloudProvider,
		opts.ExpanderStrategy,
		opts.EstimatorBuilder,
		opts.Backoff,
		opts.KubeClient), nil
}

// initializeDefaultOptions xxx
// Initialize default options if not provided.
func initializeDefaultOptions(opts *AutoscalerOptions) error {
	if opts.Processors == nil {
		opts.Processors = ca_processors.DefaultProcessors()
	}
	if opts.AutoscalingKubeClients == nil {
		opts.AutoscalingKubeClients = NewAutoscalingKubeClients(opts.Options.AutoscalingOptions, opts.KubeClient,
			opts.EventsKubeClient)
	}
	if opts.PredicateChecker == nil {
		predicateCheckerStopChannel := make(chan struct{})
		predicateChecker, err := simulator.NewPredicateChecker(opts.KubeClient, predicateCheckerStopChannel)
		if err != nil {
			return err
		}
		opts.PredicateChecker = predicateChecker
	}
	if opts.CloudProvider == nil {
		opts.CloudProvider = cloudBuilder.NewCloudProvider(opts.Options)
	}
	if opts.ExpanderStrategy == nil {
		expanderStrategy, err := factory.ExpanderStrategyFromString(opts.ExpanderName,
			opts.CloudProvider, opts.AutoscalingKubeClients, opts.KubeClient, opts.ConfigNamespace)
		if err != nil {
			return err
		}
		opts.ExpanderStrategy = expanderStrategy
	}
	if opts.EstimatorBuilder == nil {
		estimatorBuilder, err := estimatorinternal.NewEstimatorBuilder(opts.EstimatorName,
			opts.BufferedCPURatio, opts.BufferedMemRatio, opts.Options.BufferedResourceRatio)
		if err != nil {
			return err
		}
		opts.EstimatorBuilder = estimatorBuilder
	}
	if opts.Backoff == nil {
		opts.Backoff =
			backoff.NewIdBasedExponentialBackoff(clusterstate.InitialNodeGroupBackoffDuration,
				clusterstate.MaxNodeGroupBackoffDuration, clusterstate.NodeGroupBackoffResetTimeout)
	}

	return nil
}

// NewAutoscalingKubeClients builds AutoscalingKubeClients out of basic client.
func NewAutoscalingKubeClients(opts config.AutoscalingOptions, kubeClient,
	eventsKubeClient kube_client.Interface) *context.AutoscalingKubeClients {
	listerRegistryStopChannel := make(chan struct{})
	listerRegistry := util.NewListerRegistryWithDefaultListers(kubeClient, listerRegistryStopChannel)
	kubeEventRecorder := kube_util.CreateEventRecorder(eventsKubeClient)
	logRecorder, err := utils.NewStatusMapRecorder(kubeClient, opts.ConfigNamespace, kubeEventRecorder,
		opts.WriteStatusConfigMap)
	if err != nil {
		klog.Error("Failed to initialize status configmap, unable to write status events")
		// Get a dummy, so we can at least safely call the methods
		// DOTO(maciekpytel): recover from this after successful status configmap update?
		logRecorder, _ = utils.NewStatusMapRecorder(eventsKubeClient, opts.ConfigNamespace, kubeEventRecorder, false)
	}

	return &context.AutoscalingKubeClients{
		ListerRegistry: listerRegistry,
		ClientSet:      kubeClient,
		Recorder:       kubeEventRecorder,
		LogRecorder:    logRecorder,
	}
}
