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

// Package app xxx
package app

import (
	"fmt"
	"io"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/app/options"
	bcsstorage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage"
)

const (
	serviceName = "bcs-storage"
)

// App app
type App struct {
	op *options.StorageOptions
}

// NewApp create app
func NewApp(op *options.StorageOptions) *App {
	return &App{
		op: op,
	}
}

// Run the bcs-storage
func (app *App) Run() error {
	// init cert
	initCert(app.op)

	// init tracing
	closer, err := initTracingInstance(app.op)
	if err != nil {
		blog.Errorf("initTracingInstance failed: %v", err)
		return err
	}

	if closer != nil {
		defer closer.Close()
	}

	// 创建 StorageServer
	storage, err := bcsstorage.NewStorageServer(app.op)
	if err != nil {
		blog.Error("fail to create storage server. err:%s", err.Error())
		return err
	}

	if err = common.SavePid(app.op.ProcessConfig); err != nil {
		blog.Warn("fail to save pid. err:%s", err.Error())
	}

	// 初始化
	if err = storage.Init(); err != nil {
		return err
	}

	return storage.Start()
}

func initTracingInstance(op *options.StorageOptions) (io.Closer, error) {
	opts := make([]tracing.Option, 0)
	if op.Tracing.TracingSwitch != "" {
		opts = append(opts, tracing.TracerSwitch(op.Tracing.TracingSwitch))
	}
	if op.Tracing.TracingType != "" {
		opts = append(opts, tracing.TracerType(tracing.TraceType(op.Tracing.TracingType)))
	}
	if op.Tracing.RPCMetrics {
		opts = append(opts, tracing.RPCMetrics(op.Tracing.RPCMetrics))
	}
	if op.Tracing.ReportMetrics {
		opts = append(opts, tracing.ReportMetrics(op.Tracing.ReportMetrics))
	}

	// init reporter
	if op.Tracing.ReportLog {
		opts = append(opts, tracing.ReportLog(op.Tracing.ReportLog))
	}
	if op.Tracing.AgentFromEnv {
		opts = append(opts, tracing.AgentFromEnv(op.Tracing.AgentFromEnv))
	}
	if op.Tracing.AgentHostPort != "" {
		opts = append(opts, tracing.AgentHostPort(op.Tracing.AgentHostPort))
	}
	// init sampler
	if op.Tracing.SampleType != "" {
		opts = append(opts, tracing.SampleType(op.Tracing.SampleType),
			tracing.SampleParameter(op.Tracing.SampleParameter))
	}
	if op.Tracing.SampleFromEnv {
		opts = append(opts, tracing.SampleFromEnv(op.Tracing.SampleFromEnv))
	}
	if op.Tracing.SamplingServerURL != "" {
		opts = append(opts, tracing.SamplingServerURL(op.Tracing.SamplingServerURL))
	}

	tracer, err := tracing.NewInitTracing(serviceName, opts...)
	if err != nil {
		blog.Errorf("failed to init tracing factory, err: %v", err)
		return nil, err
	}
	closer, err := tracer.Init()
	if err != nil {
		blog.Errorf("failed to init tracing system, err: %v", err)
		return nil, err
	}

	blog.Infof("bcs-tracing switch: %s", op.Tracing.TracingSwitch)
	return closer, nil
}

func initCert(op *options.StorageOptions) {
	// server
	op.ServerCert.CertFile = op.CertConfig.ServerCertFile
	op.ServerCert.KeyFile = op.CertConfig.ServerKeyFile
	op.ServerCert.CAFile = op.CertConfig.CAFile

	if op.ServerCert.CertFile != "" && op.ServerCert.KeyFile != "" {
		op.ServerCert.IsSSL = true
	}

	// client
	op.ClientCert.CertFile = op.CertConfig.ClientCertFile
	op.ClientCert.KeyFile = op.CertConfig.ClientKeyFile
	op.ClientCert.CAFile = op.CertConfig.CAFile

	if op.ClientCert.CertFile != "" && op.ClientCert.KeyFile != "" {
		op.ClientCert.IsSSL = true
	}

	if op.IPv6Address == "" {
		fmt.Printf("op.IPv6Address is empty\n")
	} else {
		// 初始化IPv6Address字段
		op.ServiceConfig.InitIPv6AddressFiled()
	}
}
