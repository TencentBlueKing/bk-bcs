/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package clientset NOTES
package clientset

import (
	"fmt"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/logs"
	pbas "bscp.io/pkg/protocol/auth-server"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/serviced"
	"bscp.io/pkg/tools"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ClientSet defines configure server's all the depends api client.
type ClientSet struct {
	// DS data service's client api
	DS pbds.DataClient
	AS pbas.AuthClient
	// authorizer auth related operations.
	Authorizer auth.Authorizer
}

// NewClientSet : new clientset
func NewClientSet(sd serviced.Discover, tls cc.TLSConfig) (*ClientSet, error) {
	logs.Infof("start initialize the client set.")

	opts := make([]grpc.DialOption, 0)

	// add dial load balancer.
	opts = append(opts, sd.LBRoundRobin())

	if !tls.Enable() {
		// dial without ssl
		opts = append(opts, grpc.WithInsecure())
	} else {
		// dial with ssl.
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return nil, fmt.Errorf("init client set tls config failed, err: %v", err)
		}

		cred := credentials.NewTLS(tlsC)
		opts = append(opts, grpc.WithTransportCredentials(cred))
	}

	// connect data service.
	dsConn, err := grpc.Dial(serviced.GrpcServiceDiscoveryName(cc.DataServiceName), opts...)
	if err != nil {
		logs.Errorf("dial data service failed, err: %v", err)
		return nil, errf.New(errf.Unknown, fmt.Sprintf("dial data service failed, err: %v", err))
	}

	// connect data service.
	asConn, err := grpc.Dial(serviced.GrpcServiceDiscoveryName(cc.AuthServerName), opts...)
	if err != nil {
		logs.Errorf("dial data service failed, err: %v", err)
		return nil, errf.New(errf.Unknown, fmt.Sprintf("dial data service failed, err: %v", err))
	}

	authorizer, err := auth.NewAuthorizer(sd, cc.ConfigServer().Network.TLS)
	if err != nil {
		return nil, fmt.Errorf("new authorizer failed, err: %v", err)
	}

	cs := &ClientSet{
		DS:         pbds.NewDataClient(dsConn),
		AS:         pbas.NewAuthClient(asConn),
		Authorizer: authorizer,
	}

	logs.Infof("initialize the client set success.")
	return cs, nil
}
