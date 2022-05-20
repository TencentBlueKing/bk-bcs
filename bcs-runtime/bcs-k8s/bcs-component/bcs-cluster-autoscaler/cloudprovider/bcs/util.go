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
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-ini/ini"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/klog"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs/clustermanager"
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

var (
	config Config
	// EncryptionKey is aes key
	EncryptionKey string
)

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

	operator := os.Getenv("Operator")
	if len(operator) == 0 {
		return nil, nil, fmt.Errorf("Can not get Operator")
	}
	url := os.Getenv("BcsApiAddress")
	if len(url) == 0 {
		return nil, nil, fmt.Errorf("Can not get BcsApiAddress")
	}
	token := os.Getenv("BcsToken")
	if len(token) == 0 {
		return nil, nil, fmt.Errorf("Can not get BcsToken")
	}

	var err error
	encryption := os.Getenv("Encryption")
	if encryption == "yes" {
		if len(EncryptionKey) == 0 {
			return nil, nil, fmt.Errorf("Can not get EncryptionKey")
		}
		url, err = AesDecrypt(url, EncryptionKey)
		if err != nil {
			return nil, nil, fmt.Errorf("Can not decrypt BcsApiAddress: %s", err.Error())
		}
		token, err = AesDecrypt(token, EncryptionKey)
		if err != nil {
			return nil, nil, fmt.Errorf("Can not decrypt BcsToken: %s", err.Error())
		}
	}

	var client clustermanager.NodePoolClientInterface
	client, err = clustermanager.NewNodePoolClient(operator, url, token)
	if err != nil {
		return nil, nil, fmt.Errorf("Can not build NodePoolClient")
	}

	return NewNodeGroupCache(func(ng string) ([]*clustermanager.Node, error) {
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
	resources[apiv1.ResourceCPU] = resource.MustParse(fmt.Sprintf("%v", lc.CPU))
	resources[apiv1.ResourceMemory] = resource.MustParse(fmt.Sprintf("%v", lc.Mem) + "Gi")
	resources["gpu"] = resource.MustParse(fmt.Sprintf("%v", lc.GPU))
	return resources
}

// AesDecrypt decrypt message with key
func AesDecrypt(encrypted, key string) (string, error) {
	encrytedByte, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	k := []byte(key)

	block, err := aes.NewCipher(k)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	orig := make([]byte, len(encrytedByte))
	blockMode.CryptBlocks(orig, encrytedByte)
	orig = PKCS7UnPadding(orig)
	return string(orig), nil
}

// PKCS7UnPadding returns the unpadding text
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
