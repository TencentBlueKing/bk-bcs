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

package v1

import (
	"context"
	"fmt"
	"regexp"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/utils"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/api/proto/logmanager"
)

// LogManagerInterface defines the grpc client interfaces of LogManager
type LogManagerInterface interface {
	ObtainDataID(*proto.ObtainDataidReq) (int, error)
	CreateCleanStrategy(*proto.CreateCleanStrategyReq) error
	ListLogCollectionTask(*proto.ListLogCollectionTaskReq) ([]*proto.ListLogCollectionTaskRespItem, error)
	CreateLogCollectionTask(*proto.CreateLogCollectionTaskReq) error
	DeleteLogCollectionTask(*proto.DeleteLogCollectionTaskReq) error
}

// LogManager implements the logmanager's grpc client interface
type LogManager struct {
	ctx       context.Context
	client    proto.LogManagerClient
	bcsOption types.ClientOptions
	endpoint  string
}

// NewLogManager returns a LogManager grpc client interface
func NewLogManager(ctx context.Context, options types.ClientOptions) (LogManagerInterface, error) {
	re := regexp.MustCompile("https?://")
	s := re.Split(options.BcsApiAddress, 2)
	addr := s[len(s)-1]
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
		"authorization":  fmt.Sprintf("Bearer %s", options.BcsToken),
	}
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	opts = append(opts, grpc.WithPerRPCCredentials(utils.NewTokenAuth(options.BcsToken)))
	if options.ClientSSL != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(options.ClientSSL)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("Create log manager grpc client error: %s", err.Error())
	}
	client := proto.NewLogManagerClient(conn)
	blog.Infof("Create log manager grpc client success")
	return &LogManager{
		ctx:       ctx,
		bcsOption: options,
		client:    client,
		endpoint:  addr,
	}, nil
}
