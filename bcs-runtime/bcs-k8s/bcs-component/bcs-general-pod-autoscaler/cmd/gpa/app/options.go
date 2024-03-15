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

package app

import (
	"time"

	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/config/v1alpha1"
)

// RunOptions run options
type RunOptions struct {
	KubeconfigPath       string
	MasterUrl            string
	QPS                  int
	Burst                int
	Resync               time.Duration
	ElectionName         string
	ElectionNamespace    string
	ElectionResourceLock string
	*v1alpha1.GPAControllerConfiguration
}

// NewServerRunOptions new server run options
func NewServerRunOptions() *RunOptions {
	options := &RunOptions{GPAControllerConfiguration: &v1alpha1.GPAControllerConfiguration{}}
	options.addKubeFlags()
	options.addElectionFlags()
	options.addGPAFlags()
	RecommendDefaultGPAControllerConfig(options.GPAControllerConfiguration)
	return options
}

func (s *RunOptions) addKubeFlags() {
	pflag.DurationVar(&s.Resync, "resync", 10*time.Minute, "Time to resync from apiserver.")
	pflag.StringVar(&s.KubeconfigPath,
		"kubeconfig-path", "", "Absolute path to the kubeconfig file.")
	pflag.StringVar(&s.MasterUrl, "master", "", "Master url.")
	pflag.IntVar(&s.QPS, "qps", 100, "qps of auto scaler.")
	pflag.IntVar(&s.Burst, "burst", 200, "burst of auto scaler.")
}

func (s *RunOptions) addElectionFlags() {
	pflag.StringVar(&s.ElectionName, "election-name", "general-podautoscaler", "election name.")
	pflag.StringVar(&s.ElectionNamespace,
		"election-namespace", "bcs-system", "election namespace.")
	pflag.StringVar(&s.ElectionResourceLock,
		"election-resource-lock",
		"leases",
		"election resource type, support endoints, leases, configmaps and so on.")
}

// addGPAFlags addFlags adds flags related to GPAController for controller manager to the specified FlagSet
// nolint o should be consistent with previous receiver name s
func (o *RunOptions) addGPAFlags() {
	if o == nil {
		return
	}

	pflag.DurationVar(&o.GeneralPodAutoscalerSyncPeriod.Duration,
		"general-pod-autoscaler-sync-period", o.GeneralPodAutoscalerSyncPeriod.Duration,
		"The period for syncing the number of pods in general pod autoscaler.")
	pflag.DurationVar(&o.GeneralPodAutoscalerUpscaleForbiddenWindow.Duration,
		"general-pod-autoscaler-upscale-delay", o.GeneralPodAutoscalerUpscaleForbiddenWindow.Duration,
		"The period since last upscale, before another upscale can be performed in general pod autoscaler.")
	pflag.DurationVar(&o.GeneralPodAutoscalerDownscaleStabilizationWindow.Duration,
		"general-pod-autoscaler-downscale-stabilization",
		o.GeneralPodAutoscalerDownscaleStabilizationWindow.Duration,
		"The period for which autoscaler will look backwards and " + 
		"not scale down below any recommendation it made during that period.")
	pflag.DurationVar(
		&o.GeneralPodAutoscalerDownscaleForbiddenWindow.Duration,
		"general-pod-autoscaler-downscale-delay",
		o.GeneralPodAutoscalerDownscaleForbiddenWindow.Duration,
		"The period since last downscale, before another downscale can be performed in general pod autoscaler.")
	pflag.Float64Var(&o.GeneralPodAutoscalerTolerance,
		"general-pod-autoscaler-tolerance", o.GeneralPodAutoscalerTolerance,
		"The minimum change (from 1.0) in the desired-to-actual metrics ratio for " + 
		"the general pod autoscaler to consider scaling.")
	pflag.BoolVar(&o.GeneralPodAutoscalerUseRESTClients,
		"general-pod-autoscaler-use-rest-clients",
		o.GeneralPodAutoscalerUseRESTClients,
		"If set to true, causes the general pod autoscaler controller to use REST clients through the kube-aggregator,"+
			" instead of using the legacy metrics client through the API server proxy."+
			" This is required for custom metrics support in the general pod autoscaler.")
	pflag.DurationVar(&o.GeneralPodAutoscalerCPUInitializationPeriod.Duration,
		"general-pod-autoscaler-cpu-initialization-period",
		o.GeneralPodAutoscalerCPUInitializationPeriod.Duration,
		"The period after pod start when CPU samples might be skipped.")
	pflag.DurationVar(&o.GeneralPodAutoscalerInitialReadinessDelay.Duration,
		"general-pod-autoscaler-initial-readiness-delay",
		o.GeneralPodAutoscalerInitialReadinessDelay.Duration,
		"The period after pod start during which readiness changes will be treated as initial readiness.")
	pflag.IntVar(&o.GeneralPodAutoscalerWorkers,
		"general-pod-autoscaler-workers",
		o.GeneralPodAutoscalerWorkers,
		"The number for parallel process worker.")
}

// NewConfig new config
func (s *RunOptions) NewConfig() (*rest.Config, error) {
	var (
		config *rest.Config
		err    error
	)
	config, err = rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags(s.MasterUrl, s.KubeconfigPath)
		if err != nil {
			return nil, err
		}
	}
	config.Burst = s.Burst
	config.QPS = float32(s.QPS)
	return config, nil
}
