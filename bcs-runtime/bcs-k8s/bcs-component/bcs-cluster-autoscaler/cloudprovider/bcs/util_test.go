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

package bcs

import (
	"context"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs/clustermanager"
)

func TestNewTokenAuth(t *testing.T) {
	type args struct {
		t string
	}
	tests := []struct {
		name string
		args args
		want *GrpcTokenAuth
	}{
		{
			name: "test token auth",
			args: args{
				t: "xx",
			},
			want: &GrpcTokenAuth{
				Token: "xx",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTokenAuth(tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTokenAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrpcTokenAuth_GetRequestMetadata(t *testing.T) {
	type fields struct {
		Token string
	}
	type args struct {
		ctx context.Context
		in  []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "test get request meta data",
			fields: fields{
				Token: "xx",
			},
			args: args{
				ctx: context.TODO(),
			},
			want: map[string]string{
				"authorization": "Bearer xx",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := GrpcTokenAuth{
				Token: tt.fields.Token,
			}
			got, err := tr.GetRequestMetadata(tt.args.ctx, tt.args.in...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GrpcTokenAuth.GetRequestMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GrpcTokenAuth.GetRequestMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateNodeGroupCache(t *testing.T) {
	EncryptionKey = "abcdefghijklmnopqrstuvwx"
	os.Setenv("Operator", "bcs")
	os.Setenv("BcsApiAddress", "uGDbP6fO9fFUWldDCHd8wA==")
	os.Setenv("BcsToken", "uGDbP6fO9fFUWldDCHd8wA==")
	client, _ := clustermanager.NewNodePoolClient("bcs", "test", "test")

	type args struct {
		configReader io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *NodeGroupCache
		want1   clustermanager.NodePoolClientInterface
		wantErr bool
	}{
		{
			name: "create cache normal",
			args: args{},
			want: &NodeGroupCache{
				registeredGroups:       make([]*NodeGroup, 0),
				instanceToGroup:        make(map[InstanceRef]*NodeGroup),
				instanceToCreationType: make(map[InstanceRef]CreationType),
				getNodes: func(ng string) ([]*clustermanager.Node, error) {
					return client.GetNodes(ng)
				},
			},
			want1:   client,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := CreateNodeGroupCache(tt.args.configReader)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateNodeGroupCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				got.getNodes = nil
				tt.want.getNodes = nil
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreateNodeGroupCache() got = %v, want %v", got.getNodes, tt.want.getNodes)
				}
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("CreateNodeGroupCache() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_readConfig(t *testing.T) {
	opt := "util_test.go"
	config, fileErr := os.Open(opt)
	if fileErr != nil {
		t.Fatalf("Couldn't open cloud Provider configuration %s: %#v", opt, fileErr)
	}
	defer config.Close()

	type args struct {
		cfg io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "cfg is nil",
			args:    args{},
			wantErr: true,
		},
		{
			name: "",
			args: args{
				cfg: config,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := readConfig(tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("readConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
