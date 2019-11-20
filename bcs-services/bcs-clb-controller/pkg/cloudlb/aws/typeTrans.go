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

package aws

import (
	"fmt"
	"strings"

	"bk-bcs/bcs-common/common/blog"
	cloudlbType "bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1"
)

const (
	AWS_LOADBALANCE_NETWORK_PUBLIC   = "internet-facing"
	AWS_LOADBALANCE_NETWORK_PRIVATE  = "internal"
	AWS_LOADBALANCE_TYPE_APPLICATION = "application"
	AWS_LOADBALANCE_TYPE_NETWORK     = "network"
	AWS_LOADBALANCE_PROTOCOL_HTTP    = "HTTP"
	AWS_LOADBALANCE_PROTOCOL_HTTPS   = "HTTPS"
	AWS_LOADBALANCE_PROTOCOL_TCP     = "TCP"
)

//from bcs type to aws type
func networkTypeBcs2Aws(bcsType string) (string, error) {
	if bcsType == cloudlbType.ClbNetworkTypePrivate {
		return AWS_LOADBALANCE_NETWORK_PRIVATE, nil
	} else if bcsType == cloudlbType.ClbNetworkTypePublic {
		return AWS_LOADBALANCE_NETWORK_PUBLIC, nil
	}
	blog.Errorf("unsupported network type %s", bcsType)
	return "", fmt.Errorf("unsupported network type %s", bcsType)
}

func protocolTypeBcs2Aws(bcsType string) (string, error) {
	if bcsType == cloudlbType.ClbListenerProtocolHTTP || bcsType == cloudlbType.ClbListenerProtocolHTTPS ||
		bcsType == cloudlbType.ClbListenerProtocolTCP {
		return strings.ToUpper(bcsType), nil
	}
	blog.Errorf("unsupported protocol type %s", bcsType)
	return "", fmt.Errorf("unsupported protocol type %s", bcsType)
}

//from aws type to bcs type
func networkTypeAws2Bcs(awsType string) (string, error) {
	if awsType == AWS_LOADBALANCE_NETWORK_PRIVATE {
		return cloudlbType.ClbNetworkTypePrivate, nil
	} else if awsType == AWS_LOADBALANCE_NETWORK_PUBLIC {
		return cloudlbType.ClbNetworkTypePublic, nil
	}
	blog.Errorf("unsupported network type %s", awsType)
	return "", fmt.Errorf("unsupported network type %s", awsType)
}

func protocolTypeAws2Bcs(awsType string) (string, error) {
	if awsType == AWS_LOADBALANCE_PROTOCOL_HTTP || awsType == AWS_LOADBALANCE_PROTOCOL_HTTPS ||
		awsType == AWS_LOADBALANCE_PROTOCOL_TCP {
		return strings.ToLower(awsType), nil
	}
	blog.Errorf("unsupported protocol type %s", awsType)
	return "", fmt.Errorf("unsupported protocol type %s", awsType)
}
