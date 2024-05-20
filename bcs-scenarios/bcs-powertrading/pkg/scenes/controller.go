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

// Package scenes xxx
package scenes

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/bkcc"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/bksops"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/clustermgr"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/cr"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/job"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/resourcemgr"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/storage"
)

// Controller interface
type Controller interface {
	Init(opts ...Option) error
	Options() *Options
	Run(cxt context.Context)
}

// Option function for better injection
type Option func(o *Options)

// Options controller options
type Options struct {
	// interval for one logic loop, unit is second
	Interval int
	// Concurrency for handle cluster client request
	Concurrency    int
	BKsopsCli      bksops.Client
	JobCli         job.Client
	Storage        storage.Storage
	BkccCli        bkcc.Client
	BkCrCli        cr.Client
	ClusterMgrCli  clustermgr.Client
	ResourceMgrCli resourcemgr.Client
}
