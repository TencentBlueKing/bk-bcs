/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package datajob xxx
package datajob

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
)

// IDataJob dataJob interface
type IDataJob interface {
	DoPolicy(ctx context.Context)
	SetPolicy(Policy)
}

// Policy interface
type Policy interface {
	ImplementPolicy(ctx context.Context, opts *types.JobCommonOpts, clients *types.Clients)
}

// DataJob dataJob struct
type DataJob struct {
	Opts      types.JobCommonOpts
	jobPolicy Policy
	clients   *types.Clients
}

// DoPolicy do dataJob policy
func (j *DataJob) DoPolicy(ctx context.Context) {
	j.jobPolicy.ImplementPolicy(ctx, &j.Opts, j.clients)
}

// SetPolicy set dataJob policy
func (j *DataJob) SetPolicy(policy Policy) {
	j.jobPolicy = policy
}

// SetClient set dataJob clients
func (j *DataJob) SetClient(clients *types.Clients) {
	j.clients = clients
}
