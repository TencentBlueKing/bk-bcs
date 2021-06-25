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

// ServiceName set tracer serviceName
func ServiceName(name string) JaeOption {
	return func(opts *JaeOptions) {
		opts.ServiceName = name
	}
}

// RPCMetrics set tracer rpcMetrics
func RPCMetrics(rm bool) JaeOption {
	return func(opts *JaeOptions) {
		opts.RPCMetrics = rm
	}
}

// ReportMetrics set on prometheus report metrics
func ReportMetrics(rm bool) JaeOption {
	return func(opts *JaeOptions) {
		opts.ReportMetrics = rm
	}
}

// FromEnv set jaeger-agent hostPort from env by container deploy
func FromEnv(fe bool) JaeOption {
	return func(opts *JaeOptions) {
		opts.FromEnv = fe
	}
}

// AgentHostPort set jaeger-agent hostPort by idc deploy
func AgentHostPort(hp string) JaeOption {
	return func(opts *JaeOptions) {
		opts.AgentHostPort = hp
	}
}

// ReportLog set report tracer/span info
func ReportLog(rl bool) JaeOption {
	return func(opts *JaeOptions) {
		opts.ReportLog = rl
	}
}

// SamplerConfigInit set the jaeger sampler
func SamplerConfigInit(sc SamplerConfig) JaeOption {
	return func(opts *JaeOptions) {
		opts.Sampler = sc
	}
}
