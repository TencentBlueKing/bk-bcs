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

package jaeger

import (
	"io"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
)

func InitTracingInstance(op *config.TracingConf) (io.Closer, error) {
	opts := []tracing.Option{}
	if op.TracingSwitch != "" {
		opts = append(opts, tracing.TracerSwitch(op.TracingSwitch))
	}
	if op.TracingType != "" {
		opts = append(opts, tracing.TracerType(tracing.TraceType(op.TracingType)))
	}
	if op.RPCMetrics {
		opts = append(opts, tracing.RPCMetrics(op.RPCMetrics))
	}
	if op.ReportMetrics {
		opts = append(opts, tracing.ReportMetrics(op.ReportMetrics))
	}
	// init reporter
	if op.ReportLog {
		opts = append(opts, tracing.ReportLog(op.ReportLog))
	}
	if op.AgentFromEnv {
		opts = append(opts, tracing.AgentFromEnv(op.AgentFromEnv))
	}
	if op.AgentHostPort != "" {
		opts = append(opts, tracing.AgentHostPort(op.AgentHostPort))
	}
	// init sampler
	if op.SampleType != "" {
		opts = append(opts, tracing.SampleType(op.SampleType),
			tracing.SampleParameter(op.SampleParameter))
	}
	if op.SampleFromEnv {
		opts = append(opts, tracing.SampleFromEnv(op.SampleFromEnv))
	}
	if op.SamplingServerURL != "" {
		opts = append(opts, tracing.SamplingServerURL(op.SamplingServerURL))
	}

	tracer, err := tracing.NewInitTracing(op.ServiceName, opts...)
	if err != nil {
		blog.Errorf("failed to init tracing factory, err: %v", err)
		return nil, err
	}
	closer, err := tracer.Init()
	if err != nil {
		blog.Errorf("failed to init tracing system, err: %v", err)
		return nil, err
	}

	blog.Infof("bcs-tracing switch: %s", op.TracingSwitch)
	return closer, nil
}
