/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
查看哪些数据需要迁移
*/

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"

	"github.com/golang/protobuf/jsonpb"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/parnurzeal/gorequest"
	"google.golang.org/protobuf/runtime/protoiface"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

var (
	clusterIDs []string = []string{}
)

var (
	gatewayHost string
	authToken   string
)

func parseFlags() {

	// bcs gateway
	flag.StringVar(&gatewayHost, "gateway_host", "", "bcs gateway host")
	flag.StringVar(&authToken, "auth_token", "", "bcs gateway auth token")

	flag.Parse()
}

func main() {
	parseFlags()
	// r := gorequest.New().Get(gatewayHost)
	// r.TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	// if authToken == "" {
	// 	panic("empty token")
	// }
	// r.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	projectCodes := []string{}
	clientset.SetClientGroup(gatewayHost, authToken)
	cg := clientset.GetClientGroup()
	for _, clusterID := range clusterIDs {
		cli, err := cg.Client(clusterID)
		if err != nil {
			panic(err)
		}
		nsList, err := cli.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			panic(err)
		}
		for _, ns := range nsList.Items {
			annotations := ns.GetAnnotations()
			for k, v := range annotations {
				if k == "io.tencent.bcs.projectcode" {
					projectCodes = append(projectCodes, v)
				}
			}
		}
	}

	fmt.Println(projectCodes)
	results := []string{}
	for _, projectCode := range projectCodes {
		r := gorequest.New().Get(gatewayHost + "/bcsapi/v4/bcsproject/v1/projects/" + projectCode)
		r.TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		r.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		resp, _, errs := r.End()
		if len(errs) != 0 {
			panic(errs[0])
		}
		if resp.StatusCode != 200 {
			panic(fmt.Sprintf("response error, status: %d", resp.StatusCode))
		}
		rresp := &proto.ProjectResponse{}
		err := UnmarshalPB(resp.Body, rresp)
		if err != nil {
			panic(err)
		}
		if rresp.GetCode() != 0 {
			panic(fmt.Sprintf("resp error, code: %d", rresp.GetCode()))
		}
		if rresp.GetData().GetBusinessID() == "" {
			if !stringx.StringInSlice(rresp.GetData().GetProjectCode(), results) {
				results = append(results, rresp.GetData().GetProjectCode())
			}
		}
	}
	fmt.Println(results)
}

func UnmarshalPB(r io.Reader, m protoiface.MessageV1) error {
	marshaler := &jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}
	return marshaler.Unmarshal(r, m)
}
