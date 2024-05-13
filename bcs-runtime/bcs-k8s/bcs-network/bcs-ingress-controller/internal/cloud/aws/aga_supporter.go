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

package aws

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
)

// AgaMappingInfo aggregate port map info of aga
type AgaMappingInfo struct {
	Arn     string `json:"arn"`
	Enabled *bool  `json:"enabled"`

	DNSName *string                    `json:"DNSName"`
	IPSets  []*globalaccelerator.IpSet `json:"IPSets"`

	SubnetID                *string `json:"subnetID,omitempty"`
	SubnetRegion            *string `json:"subnetRegion,omitempty"`
	DestinationTrafficState *string `json:"destinationTrafficState,omitempty"`

	PortMappings []*portMapping `json:"portMappings,omitempty"`
}

type portMapping struct {
	CloudStartPort int64 `json:"cloudStartPort,omitempty"`
	CloudEndPort   int64 `json:"cloudEndPort,omitempty"`

	LocalStartPort int64 `json:"localStartPort,omitempty"`
	LocalEndPort   int64 `json:"localEndPort,omitempty"`
}

type portPair struct {
	CloudPort int64
	LocalPort int64
}

// AgaSupporter help handle aga
type AgaSupporter struct {
	sessionMap sync.Map
	secretID   string
	secretKey  string
}

// NewAgaSupporter return new aga supporter
func NewAgaSupporter() *AgaSupporter {
	sup := &AgaSupporter{}
	sup.loadEnv()
	return sup
}

func (a *AgaSupporter) loadEnv() {
	if len(a.secretID) == 0 {
		a.secretID = os.Getenv(EnvNameAWSAccessKeyID)
	}
	if len(a.secretKey) == 0 {
		a.secretKey = os.Getenv(EnvNameAWSAccessKey)
	}
}

func (a *AgaSupporter) getRegionSession(region string) (*session.Session, error) {
	iSession, ok := a.sessionMap.Load(region)
	if !ok {
		newSession, err := session.NewSession(&aws.Config{
			Region:      aws.String(region), // 根据需要更改区域
			Credentials: credentials.NewStaticCredentials(a.secretID, a.secretKey, ""),
		})
		if err != nil {
			blog.Errorf("create aws session for region %s failed, err %s", region, err.Error())
			return nil, fmt.Errorf("create aws session for region %s failed, err %s", region, err.Error())
		}
		a.sessionMap.Store(region, newSession)
		return newSession, nil
	}
	sess, ok := iSession.(*session.Session)
	if !ok {
		blog.Errorf("unknown type store in sessionMap, value: %v", iSession)
		return nil, fmt.Errorf("unknown type store in sessionMap, value: %v", iSession)
	}

	return sess, nil
}

// ListCustomRoutingByDefinition param
func (a *AgaSupporter) ListCustomRoutingByDefinition(agaHostRegion, region, instanceID,
	destAddress string) ([]*AgaMappingInfo,
	error) {
	blog.Infof("ListCustomRoutingByDefinition req[region=%s, instanceID=%s, destAddress=%s]", region, instanceID,
		destAddress)
	sess, err := a.getRegionSession(region)
	if err != nil {
		return nil, err
	}

	// aga host can only be called in us-west region now, use a differentregion here
	agaSess, err := a.getRegionSession(agaHostRegion)
	if err != nil {
		return nil, err
	}

	subnetID, err := a.getSubnetIDOfInstance(sess, instanceID)
	if err != nil {
		return nil, err
	}

	portMappingResp, err := a.listCustomRoutingPortMappingsByDest(agaSess, subnetID, destAddress)
	if err != nil {
		return nil, err
	}

	resp, err := a.mergeCustomRouting(agaSess, portMappingResp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (a *AgaSupporter) getSubnetIDOfInstance(sess *session.Session, instanceID string) (string, error) {
	blog.V(4).Infof("DescribeInstance req[%s]", instanceID)
	ecSvc := ec2.New(sess)
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID),
		},
	}
	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDKEC2,
			"DescribeInstances", ret, startTime)
	}
	result, err := ecSvc.DescribeInstances(input)
	if err != nil {
		blog.Errorf("DescribeInstance '%s' failed, err: %s", instanceID, err.Error())
		mf(metrics.LibCallStatusErr)
		return "", errors.Wrapf(err, "DescribeInstance '%s' failed", instanceID)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		blog.Errorf("Invalid instance result[req:%s]: %s", instanceID, common.ToJsonString(result))
		mf(metrics.LibCallStatusErr)
		return "", errors.Errorf("Invalid instance result[req:%s]: %s", instanceID, common.ToJsonString(result))
	}

	subnetID := result.Reservations[0].Instances[0].SubnetId
	if subnetID == nil {
		blog.Errorf("unknown error: instance[%s] subnetID is empty", instanceID)
		mf(metrics.LibCallStatusErr)
		return "", errors.Errorf("unknown error: instance[%s] subnetID is empty", instanceID)
	}
	mf(metrics.LibCallStatusOK)
	blog.V(4).Infof("DescribeInstance [%s] related subnetID: %s", instanceID, *subnetID)
	return aws.StringValue(subnetID), nil
}

func (a *AgaSupporter) listCustomRoutingPortMappingsByDest(sess *session.Session, subnetID,
	ipAddress string) (*globalaccelerator.ListCustomRoutingPortMappingsByDestinationOutput, error) {
	blog.V(4).Infof("ListCustomRoutingByDefinition req[subnetID='%s', IPAddr='%s']", subnetID, ipAddress)
	svc := globalaccelerator.New(sess)
	startTime := time.Now()

	// 统计API调用延时/状态
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDKAGA,
			"ListCustomRoutingPortMappingsByDestination", ret, startTime)
	}
	input := &globalaccelerator.ListCustomRoutingPortMappingsByDestinationInput{
		EndpointId:         aws.String(subnetID),
		DestinationAddress: aws.String(ipAddress),
	}

	result, err := svc.ListCustomRoutingPortMappingsByDestination(input)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		blog.Errorf("ListCustomRoutingByDefinition req[subnetID='%s', IPAddr='%s'] failed: %s", subnetID,
			ipAddress, err.Error())
		return nil, errors.Wrapf(err, "ListCustomRoutingByDefinition req[subnetID='%s', IPAddr='%s'] failed", subnetID,
			ipAddress)
	}
	mf(metrics.LibCallStatusOK)
	blog.V(4).Infof("ListCustomRoutingByDefinition req[subnetID='%s', IPAddr='%s'] success, resp: %s", subnetID,
		ipAddress, common.ToJsonString(result))
	return result, nil
}

func (a *AgaSupporter) describeAccelerator(sess *session.Session,
	agaArn string) (*globalaccelerator.DescribeCustomRoutingAcceleratorOutput, error) {
	blog.V(4).Infof("DescribeAccelerator req[arn='%s']", agaArn)

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDKAGA,
			"DescribeAccelerator", ret, startTime)
	}

	svc := globalaccelerator.New(sess)
	input := &globalaccelerator.DescribeCustomRoutingAcceleratorInput{AcceleratorArn: aws.String(agaArn)}
	result, err := svc.DescribeCustomRoutingAccelerator(input)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		blog.Errorf("DescribeAccelerator req[arn='%s'] failed, err: %s", agaArn, err.Error())
		return nil, errors.Wrapf(err, "DescribeAccelerator req[arn='%s'] failed", agaArn)
	}
	mf(metrics.LibCallStatusOK)
	blog.V(4).Infof("DescribeAccelerator req[arn='%s'] success", agaArn)
	return result, nil
}

func (a *AgaSupporter) mergeCustomRouting(sess *session.Session, portMappingResp *globalaccelerator.
	ListCustomRoutingPortMappingsByDestinationOutput) ([]*AgaMappingInfo, error) {
	agaMappingInfoList := make([]*AgaMappingInfo, 0)
	mappings := make(map[string][]portPair)

	for _, pm := range portMappingResp.DestinationPortMappings {
		arn := aws.StringValue(pm.AcceleratorArn)

		cloudAddress := pm.AcceleratorSocketAddresses[0]
		localAddress := pm.DestinationSocketAddress

		mappings[arn] = append(mappings[arn], portPair{CloudPort: aws.Int64Value(cloudAddress.Port),
			LocalPort: aws.Int64Value(localAddress.Port)})
	}

	for agaArn, portPairs := range mappings {
		agaInfo, err := a.describeAccelerator(sess, agaArn)
		if err != nil {
			return nil, err
		}

		portMappings := a.splitPortMappings(portPairs)

		agaMapInfo := &AgaMappingInfo{
			Arn:          agaArn,
			Enabled:      agaInfo.Accelerator.Enabled,
			DNSName:      agaInfo.Accelerator.DnsName,
			IPSets:       agaInfo.Accelerator.IpSets,
			PortMappings: portMappings,
		}

		agaMappingInfoList = append(agaMappingInfoList, agaMapInfo)
	}

	return agaMappingInfoList, nil
}

func (a *AgaSupporter) splitPortMappings(portPairs []portPair) []*portMapping {
	if len(portPairs) == 0 {
		blog.Warnf("get empty port pairs")
		return make([]*portMapping, 0)
	}
	// 根据cloudPort从小到大排序，便于后续分段
	sort.Slice(portPairs, func(i, j int) bool {
		return portPairs[i].CloudPort < portPairs[j].CloudPort
	})

	portMappings := make([]*portMapping, 0)

	var cloudStartPort int64
	var localStartPort int64
	for i := 0; i < len(portPairs); i++ {
		if cloudStartPort == 0 {
			cloudStartPort = portPairs[i].CloudPort
			localStartPort = portPairs[i].LocalPort
			continue
		}

		// 当端口不连续时(无论是cloud还是local)，认为是一个新的段
		// eg,
		// cloudPort 1080 1081 1082 1083 1090 1091 1092
		// localPort   80   81   85   86   90   91   92
		// 上例应该被分为三段（1080,1081）, (1082,1083), (1090,1092)
		if portPairs[i].CloudPort != portPairs[i-1].CloudPort+1 || portPairs[i].LocalPort != portPairs[i-1].
			LocalPort+1 {
			portMappings = append(portMappings, &portMapping{
				CloudStartPort: cloudStartPort,
				CloudEndPort:   portPairs[i-1].CloudPort,
				LocalStartPort: localStartPort,
				LocalEndPort:   portPairs[i-1].LocalPort,
			})

			cloudStartPort = portPairs[i].CloudPort
			localStartPort = portPairs[i].LocalPort
		}
	}

	// 最后一段
	portMappings = append(portMappings, &portMapping{
		CloudStartPort: cloudStartPort,
		CloudEndPort:   portPairs[len(portPairs)-1].CloudPort,
		LocalStartPort: localStartPort,
		LocalEndPort:   portPairs[len(portPairs)-1].LocalPort,
	})

	return portMappings
}
