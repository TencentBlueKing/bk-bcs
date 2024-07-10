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

package resourcemgr

import (
	"context"
	"crypto/tls"
	"fmt"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	impl "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/resourcemgr/proto"
)

func TestGetTaskByID(t *testing.T) {
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	header["Authorization"] = fmt.Sprintf("Bearer %s", "")
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}))) // nolint
	var conn *grpc.ClientConn
	conn, err := grpc.Dial("", opts...)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	client := impl.NewResourceManagerClient(conn)
	recordID := ""
	req := &impl.GetDeviceRecordReq{DeviceRecordID: &recordID}
	result, err := client.GetDeviceRecord(ctx, req)
	fmt.Println(err)
	fmt.Println(result)
	recordType := []int64{6}
	status := []int64{1}
	limit := int64(10000)
	listReq := &impl.ListDeviceRecordReq{
		Type:   recordType,
		Status: status,
		Limit:  &limit,
	}
	listResult, err := client.ListDeviceRecord(ctx, listReq)
	fmt.Println(err)
	fmt.Println(listResult)
	fmt.Println(listResult.Data)
}
