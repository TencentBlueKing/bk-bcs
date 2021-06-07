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
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	jaegerclient "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/rpcmetrics"
	"github.com/uber/jaeger-lib/metrics"
	jprom "github.com/uber/jaeger-lib/metrics/prometheus"
)

// jaeger default errInfo
var (
	// errInitServer for jaeger server not inited
	errInitServer = errors.New("jaeger server not inited")
	// errSetServiceName for jaeger must set service name
	errSetServiceName = errors.New("jaeger server not set service name")
	// errSamplerType for jaeger not support sampler type
	errSamplerType = errors.New("jaeger server not support sampler, please input sample type: " +
		"const/remote/probabilistic/ratelimiting/lowerbound")
	// errInitJaegerFailed for jaeger init failed
	errInitJaegerFailed = errors.New("init jaeger server failed")
	// errSamplerRemoteInit for remote sampler conf
	errSamplerRemoteInit = errors.New("jaeger server init sampler failed: please input correct SamplingServerURL")
)

// Jaeger sampler strategy conf
const (
	// SamplerTypeConst is the type of sampler that always makes the same decision.
	SamplerTypeConst = "const"

	// SamplerTypeRemote is the type of sampler that polls Jaeger agent for sampling strategy.
	SamplerTypeRemote = "remote"

	// SamplerTypeProbabilistic is the type of sampler that samples traces
	// with a certain fixed probability.
	SamplerTypeProbabilistic = "probabilistic"

	// SamplerTypeRateLimiting is the type of sampler that samples
	// only up to a fixed number of traces per second.
	SamplerTypeRateLimiting = "ratelimiting"

	// SamplerTypeLowerBound is the type of sampler that samples
	// at least a fixed number of traces per second.
	SamplerTypeLowerBound = "lowerbound"
)

// Jaeger default Options conf
var (
	// defaultReportMetrics set default metrics conf false
	defaultReportMetrics = false
	// defaultRPCMetrics
	defaultRPCMetrics = false
	// defaultReportLog set default log report conf false
	defaultReportLog = false
	// defaultAgentHostPort set default agent host port
	defaultAgentHostPort = fmt.Sprintf("%s:%d", jaegerclient.DefaultUDPSpanServerHost, jaegerclient.DefaultUDPSpanServerPort)
	// defaultSampleType show sampler always to sample all
	defaultSampleType = SamplerTypeConst
	// defaultSampleParameter set default sample parameter for SamplerTypeConst
	defaultSampleParameter = 1
	// defaultSamplingServerURL is the default url to fetch sampling config from, via http
	defaultSamplingServerURL = fmt.Sprintf("http://127.0.0.1:%d/sampling", jaegerclient.DefaultSamplingServerPort)
)

const (
	// environment variable names
	envAgentHost        = "JAEGER_AGENT_HOST"
	envAgentPort        = "JAEGER_AGENT_PORT"
	envSamplingEndpoint = "JAEGER_SAMPLING_ENDPOINT"
)

// JaeOption init JaeOptions
type JaeOption func(op *JaeOptions)

// JaeOptions for jaeger system conf
type JaeOptions struct {
	ServiceName   string `json:"serviceName"`
	RPCMetrics    bool   `json:"rPCMetrics"`
	ReportMetrics bool   `json:"reportMetrics"`
	// reporter
	ReportLog     bool   `json:"reportLog"`
	FromEnv       bool   `json:"fromEnv"`
	AgentHostPort string `json:"agentHostPort"`
	// sampler
	Sampler SamplerConfig `json:"sampler"`
}

// SamplerConfig for sample decision config
type SamplerConfig struct {
	SampleType      string  `json:"sampleType"`
	SampleParameter float64 `json:"sampleParameter"`
	// SamplerConfig get SamplingServerURL by env(JAEGER_SAMPLING_ENDPOINT) when set FromEnv, else by SamplingServerURL directly
	FromEnv           bool   `json:"fromEnv"`
	SamplingServerURL string `json:"samplingServerURL"`
}

// NewJaegerServer for jaeger system init
func NewJaegerServer(opts ...JaeOption) (*Jaeger, error) {
	defaultJaeOptions := &JaeOptions{
		ServiceName:   "",
		RPCMetrics:    defaultRPCMetrics,
		ReportMetrics: defaultReportMetrics,
		ReportLog:     defaultReportLog,
		FromEnv:       false,
		AgentHostPort: defaultAgentHostPort,

		Sampler: SamplerConfig{
			SampleType:        defaultSampleType,
			SampleParameter:   float64(defaultSampleParameter),
			FromEnv:           false,
			SamplingServerURL: defaultSamplingServerURL,
		},
	}

	for _, opt := range opts {
		opt(defaultJaeOptions)
	}

	err := validateJaeOptions(defaultJaeOptions)
	if err != nil {
		blog.Errorf("NewJaegerServer failed: %v", err)
		return nil, err
	}

	return &Jaeger{
		Opts: defaultJaeOptions,
	}, nil
}

func validateJaeOptions(opt *JaeOptions) error {
	if len(opt.ServiceName) == 0 {
		return errSetServiceName
	}

	switch opt.Sampler.SampleType {
	case SamplerTypeConst, SamplerTypeLowerBound, SamplerTypeProbabilistic, SamplerTypeRateLimiting, SamplerTypeRemote:
	default:
		return errSamplerType
	}

	if opt.Sampler.SampleType == SamplerTypeRemote {
		if !opt.Sampler.FromEnv && len(opt.Sampler.SamplingServerURL) == 0 {
			return errSamplerRemoteInit
		}
	}

	return nil
}

// Jaeger will enable tracing system
type Jaeger struct {
	Opts *JaeOptions
}

// Init init jaeger tracing system
func (j *Jaeger) Init() (io.Closer, error) {
	if j == nil {
		return nil, errInitServer
	}

	cfg := jaegercfg.Configuration{
		Sampler:  &jaegercfg.SamplerConfig{},
		Reporter: &jaegercfg.ReporterConfig{},
	}
	// set serviceName
	cfg.ServiceName = j.Opts.ServiceName

	// set sampler
	cfg.Sampler.Type = j.Opts.Sampler.SampleType
	cfg.Sampler.Param = j.Opts.Sampler.SampleParameter
	if j.Opts.Sampler.SampleType == SamplerTypeRemote {
		if j.Opts.Sampler.FromEnv {
			if e := os.Getenv(envSamplingEndpoint); e != "" {
				cfg.Sampler.SamplingServerURL = e
			}
		} else {
			cfg.Sampler.SamplingServerURL = j.Opts.Sampler.SamplingServerURL
		}
	}

	// set reporter
	var (
		host, port = "", ""
	)
	// conf hostPort
	if !j.Opts.FromEnv {
		cfg.Reporter.LocalAgentHostPort = j.Opts.AgentHostPort
	} else {
		// env hostPort
		if e := os.Getenv(envAgentHost); e != "" {
			host = e
		}
		if p := os.Getenv(envAgentPort); p != "" {
			port = p
		}
		if host != "" && port != "" {
			cfg.Reporter.LocalAgentHostPort = fmt.Sprintf("%s:%s", host, port)
		}
	}

	metricsFactory := jprom.New().Namespace(metrics.NSOptions{Name: cfg.ServiceName, Tags: nil})
	metricsFactory = metricsFactory.Namespace(metrics.NSOptions{Name: cfg.ServiceName, Tags: nil})

	jaeOpts := []jaegercfg.Option{}
	if j.Opts.ReportMetrics {
		blog.Info("Using Prometheus as metrics backend")
		jaeOpts = append(jaeOpts, jaegercfg.Metrics(metricsFactory))
	}

	if j.Opts.ReportLog {
		blog.Info("report tracer and span Info to log")
		cfg.Reporter.LogSpans = true
		jaeOpts = append(jaeOpts, jaegercfg.Logger(jaegerLoggerAdapter{}))
	}

	if j.Opts.RPCMetrics {
		blog.Info("report tracer and span RPCMetrics")
		jaeOpts = append(jaeOpts, jaegercfg.Observer(rpcmetrics.NewObserver(metricsFactory, rpcmetrics.DefaultNameNormalizer)))
	}

	closer, err := cfg.InitGlobalTracer(cfg.ServiceName, jaeOpts...)
	if err != nil {
		blog.Errorf("Could not initialize jaeger tracer: %s", err.Error())
		return nil, errInitJaegerFailed
	}

	return closer, nil
}

type jaegerLoggerAdapter struct{}

// Error log span err
func (l jaegerLoggerAdapter) Error(msg string) {
	blog.Error(msg)
}

// Infof log span info
func (l jaegerLoggerAdapter) Infof(msg string, args ...interface{}) {
	blog.Infof(msg, args)
}
