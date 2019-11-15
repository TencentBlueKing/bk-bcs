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

package cbs

import (
	"context"
	"net"
	"net/url"
	"os"
	"path"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"google.golang.org/grpc"
)

var (
	DriverName     = "com.tencent.cloud.csi.cbs"
	DriverVerision = "0.1.0"
)

type Driver struct {
	region    string
	zone      string
	secretId  string
	secretKey string
}

func NewDriver(region string, zone string, secretId string, secretKey string) (*Driver, error) {
	driver := Driver{
		zone:      zone,
		region:    region,
		secretId:  secretId,
		secretKey: secretKey,
	}

	return &driver, nil
}

func (drv *Driver) Run(endpoint *url.URL, cbsUrl string) error {
	controller, err := newCbsController(drv.secretId, drv.secretKey, drv.region, drv.zone, cbsUrl)
	if err != nil {
		return err
	}

	identity, err := newCbsIdentity()
	if err != nil {
		return err
	}

	node, err := newCbsNode(drv.secretId, drv.secretKey, drv.region)
	if err != nil {
		return err
	}

	logGRPC := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		glog.Infof("GRPC call: %s, request: %+v", info.FullMethod, req)
		resp, err := handler(ctx, req)
		if err != nil {
			glog.Errorf("GRPC error: %v", err)
		} else {
			glog.Infof("GRPC error: %v, response: %+v", err, resp)
		}
		return resp, err
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(logGRPC),
	}

	srv := grpc.NewServer(opts...)

	csi.RegisterControllerServer(srv, controller)
	csi.RegisterIdentityServer(srv, identity)
	csi.RegisterNodeServer(srv, node)

	if endpoint.Scheme == "unix" {
		sockPath := path.Join(endpoint.Host, endpoint.Path)
		if _, err := os.Stat(sockPath); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		} else {
			if err := os.Remove(sockPath); err != nil {
				return err
			}
		}
	}

	listener, err := net.Listen(endpoint.Scheme, path.Join(endpoint.Host, endpoint.Path))
	if err != nil {
		return err
	}

	return srv.Serve(listener)
}
