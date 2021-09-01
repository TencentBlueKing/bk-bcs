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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/go-ini/ini"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/klog"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"

	"github.com/bk-bcs/bcs-k8s/bcs-cluster-autoscaler/cloudprovider/bcs/clustermanager"
)

const (
	intervalTimeDetach = 5 * time.Second
)

// Config bcs config
type Config struct {
	Region     string `json:"region"`
	RegionName string `json:"regionName"`
	ClusterID  string
}

var config Config

func readConfig(cfg io.Reader) error {
	if cfg == nil {
		err := fmt.Errorf("No cloud provider config given")
		return err
	}

	if err := json.NewDecoder(cfg).Decode(&config); err != nil {
		klog.Errorf("Couldn't parse config: %v", err)
		return err
	}

	clusterInfo, err := ini.Load("/etc/kubernetes/config")
	if err != nil {
		klog.Errorf("read clusterId from /etc/kubernetes/config failed, %s", err.Error())
		return err
	}

	section := clusterInfo.Section("")
	if !section.HasKey("KUBE_CLUSTER") {
		return fmt.Errorf("KUBE_CLUSTER not found")
	}
	config.ClusterID = section.Key("KUBE_CLUSTER").String()
	klog.Infof("read clusterId from /etc/kubernetes/config, clusterId : %s", config.ClusterID)
	klog.Infof("bcs config %+v", config)

	return nil
}

// GrpcTokenAuth grpc token
type GrpcTokenAuth struct {
	Token string
}

// NewTokenAuth implementations of grpc credentials interface
func NewTokenAuth(t string) *GrpcTokenAuth {
	return &GrpcTokenAuth{
		Token: t,
	}
}

// GetRequestMetadata convert http Authorization for grpc key
func (t GrpcTokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", t.Token),
	}, nil
}

// RequireTransportSecurity RequireTransportSecurity
func (GrpcTokenAuth) RequireTransportSecurity() bool {
	return false
}

// CreateNodeGroupCache constructs node group cache object.
func CreateNodeGroupCache(configReader io.Reader) (*NodeGroupCache, clustermanager.NodePoolClientInterface, error) {
	if configReader != nil {
		err := readConfig(configReader)
		if err != nil {
			return nil, nil, err
		}
	}

	var opts []grpc.DialOption

	endpoint := os.Getenv("BcsApiAddress")
	re := regexp.MustCompile("https?://")
	s := re.Split(endpoint, 2)
	endpoint = s[len(s)-1]
	if len(endpoint) == 0 {
		klog.Errorf("Can not get Endpoint")
	}
	token := os.Getenv("BcsToken")
	if len(token) == 0 {
		klog.Errorf("Can not get BcsToekn")
	}
	tlsFile := os.Getenv("TlsFile")
	if len(tlsFile) == 0 {
		klog.Infof("Can not get tls file configuration, build grpc client without credentials")
		opts = append(opts, grpc.WithInsecure())
	} else {
		klog.Infof("Build grpc client with credentials")
		tls, err := credentials.NewClientTLSFromFile(tlsFile, "")
		if err != nil {
			klog.Errorf("Can not load tls file")
			return nil, nil, fmt.Errorf("Load TLS file %s failed", tlsFile)
		}
		opts = append(opts, grpc.WithTransportCredentials(tls))
	}

	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
		"authorization":  fmt.Sprintf("Bearer %s", token),
	}
	md := metadata.New(header)
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	opts = append(opts, grpc.WithPerRPCCredentials(NewTokenAuth(token)))
	var client clustermanager.NodePoolClientInterface
	client, err := clustermanager.NewNodePoolClient(endpoint, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("Can not build NodePoolClient")
	}

	return newNodeGroupCache(func(ng string) ([]*clustermanager.Node, error) {
		return client.GetNodes(ng)
	}), client, nil
}

// InstanceRef contains a reference to some entity.
type InstanceRef struct {
	Name string
	IP   string
}

// InstanceRefFromProviderID creates InstanceConfig object from provider id which
// must be in format: qcloud:///100003/ins-3ven36lk
func InstanceRefFromProviderID(id string) (*InstanceRef, error) {
	validIDRegex := regexp.MustCompile(`^qcloud\:\/\/\/[-0-9a-z]*\/[-0-9a-z]*$`)
	if validIDRegex.FindStringSubmatch(id) == nil {
		return nil, fmt.Errorf("Wrong id: expected format qcloud:///zoneid/ins-<name>, got %v", id)
	}
	splitted := strings.Split(id[10:], "/")
	return &InstanceRef{
		Name: splitted[1],
	}, nil
}

// buildNodeGroupFromSpec builds node group with value format: <min>:<max>:<nodeGroupID>
func buildNodeGroupFromSpec(value string, client clustermanager.NodePoolClientInterface) (*NodeGroup, error) {
	spec, err := dynamic.SpecFromString(value, true)

	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}

	group := buildNodeGroup(client, spec.MinSize, spec.MaxSize, spec.Name)

	return group, nil
}

func buildNodeGroup(client clustermanager.NodePoolClientInterface, minSize int, maxSize int, name string) *NodeGroup {
	return &NodeGroup{
		client:      client,
		scalingType: ScalingTypeClassic,
		minSize:     minSize,
		maxSize:     maxSize,
		nodeGroupID: name,
		InstanceRef: InstanceRef{
			Name: name,
		},
	}
}

func buildGenericLabels(template *nodeTemplate, nodeName string) map[string]string {
	result := make(map[string]string)
	// TODO: extract it somehow
	result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
	result[kubeletapis.LabelOS] = cloudprovider.DefaultOS
	result[apiv1.LabelInstanceType] = template.InstanceType
	result[apiv1.LabelZoneRegion] = template.Region
	result[apiv1.LabelHostname] = nodeName
	return result
}

func convertResource(lc *clustermanager.LaunchConfiguration) map[apiv1.ResourceName]resource.Quantity {
	resources := map[apiv1.ResourceName]resource.Quantity{}
	resources[apiv1.ResourceCPU] = *resource.NewQuantity(int64(lc.CPU), resource.DecimalSI)
	resources[apiv1.ResourceMemory] = *resource.NewQuantity(int64(lc.Mem), resource.DecimalSI)
	resources["gpu"] = *resource.NewQuantity(int64(lc.GPU), resource.DecimalSI)
	return resources
}
