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

package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go-micro.dev/v4/codec"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
)

func TestSkipClientWildcardDoesNotCoverInstalledMeshes(t *testing.T) {
	origin := options.GlobalOptions
	t.Cleanup(func() { options.GlobalOptions = origin })
	options.GlobalOptions = &options.MeshManagerOptions{
		Auth: options.AuthConfig{
			ClientPermissions: `{"admin":["*"],"billing":["MeshManagerInternal.ListInstalledMeshes"]}`,
		},
	}

	require.False(t, SkipClient(context.Background(), authRequest{method: common.MeshManagerListInstalledMeshes}, "admin"))
	require.True(t, SkipClient(context.Background(), authRequest{method: common.MeshManagerListInstalledMeshes}, "billing"))
	require.True(t, SkipClient(context.Background(), authRequest{method: common.MeshManagerListIstio}, "admin"))
	require.False(t, SkipClient(context.Background(), authRequest{method: common.MeshManagerListIstio}, "billing"))
}

type authRequest struct {
	method string
}

func (r authRequest) Service() string           { return "meshmanager" }
func (r authRequest) Method() string            { return r.method }
func (r authRequest) Endpoint() string          { return r.method }
func (r authRequest) ContentType() string       { return "application/grpc" }
func (r authRequest) Header() map[string]string { return nil }
func (r authRequest) Body() interface{}         { return nil }
func (r authRequest) Read() ([]byte, error)     { return nil, nil }
func (r authRequest) Codec() codec.Reader       { return nil }
func (r authRequest) Stream() bool              { return false }
