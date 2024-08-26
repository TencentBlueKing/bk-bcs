/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

const (
	// ResourceTypeLaunchTemplate is a ResourceType enum value
	ResourceTypeLaunchTemplate = "launch-template"
)

var (
	// DefaultUserDataHeader default user data header for creating launch template userdata
	DefaultUserDataHeader = "MIME-Version: 1.0\nContent-Type: multipart/mixed; boundary=\"==MYBOUNDARY==\"\n\n" +
		"--==MYBOUNDARY==\nContent-Type: text/x-shellscript; charset=\"us-ascii\"\n\n"
	// DefaultUserDataTail default user data tail for creating launch template userdata
	DefaultUserDataTail = "\n\n--==MYBOUNDARY==--"
)

var (
	// DeviceName device name list
	DeviceName = []string{"/dev/xvdb", "/dev/xvdc", "/dev/xvdd", "/dev/xvde", "/dev/xvdf", "/dev/xvdg", "/dev/xvdh",
		"/dev/xvdi", "/dev/xvdj", "/dev/xvdk", "/dev/xvdl", "/dev/xvdm", "/dev/xvdn", "/dev/xvdo", "/dev/xvdp",
		"/dev/xvdq", "/dev/xvdr", "/dev/xvds", "/dev/xvdt", "/dev/xvdu", "/dev/xvdv", "/dev/xvdw", "/dev/xvdx",
		"/dev/xvdy", "/dev/xvdz", "/dev/xvdba", "/dev/xvdbb", "/dev/xvdbc", "/dev/xvdbd", "/dev/xvdbe", "/dev/xvdbf",
		"/dev/xvdbg", "/dev/xvdbh", "/dev/xvdbi", "/dev/xvdbj", "/dev/xvdbk", "/dev/xvdbl", "/dev/xvdbm", "/dev/xvdbn",
		"/dev/xvdbo", "/dev/xvdbp", "/dev/xvdbq", "/dev/xvdbr", "/dev/xvdbs", "/dev/xvdbt", "/dev/xvdbu", "/dev/xvdbv",
		"/dev/xvdbw", "/dev/xvdbx", "/dev/xvdby", "/dev/xvdbz"}
)

// AWSClientSet aws client set
type AWSClientSet struct {
	*AutoScalingClient
	*EC2Client
	*EksClient
	*IAMClient
}

// NewSession generates a new aws session
func NewSession(opt *cloudprovider.CommonOption) (*session.Session, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}

	awsConf := &aws.Config{}
	awsConf.Region = aws.String(opt.Region)
	awsConf.Credentials = credentials.NewStaticCredentials(opt.Account.SecretID, opt.Account.SecretKey, "")

	sess, err := session.NewSession(awsConf)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// NewAWSClientSet creates a aws client set
func NewAWSClientSet(opt *cloudprovider.CommonOption) (*AWSClientSet, error) {
	sess, err := NewSession(opt)
	if err != nil {
		return nil, err
	}

	return &AWSClientSet{
		&AutoScalingClient{asClient: autoscaling.New(sess)},
		&EC2Client{ec2Client: ec2.New(sess)},
		&EksClient{eksClient: eks.New(sess)},
		&IAMClient{iamClient: iam.New(sess)},
	}, nil
}

// GenerateAwsRestConf generate aws rest config
func GenerateAwsRestConf(opt *cloudprovider.CommonOption, cluster *eks.Cluster) (*rest.Config, error) {
	sess, err := NewSession(opt)
	if err != nil {
		return nil, err
	}

	generator, err := token.NewGenerator(false, false)
	if err != nil {
		return nil, err
	}

	awsToken, err := generator.GetWithOptions(&token.GetTokenOptions{
		Session:   sess,
		ClusterID: *cluster.Name,
	})
	if err != nil {
		return nil, err
	}

	decodedCA, err := base64.StdEncoding.DecodeString(*cluster.CertificateAuthority.Data)
	if err != nil {
		return nil, err
	}

	restConfig := &rest.Config{
		Host: *cluster.Endpoint,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: decodedCA,
		},
		BearerToken: awsToken.Token,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	return restConfig, nil
}

// GetClusterKubeConfig constructs the cluster kubeConfig
func GetClusterKubeConfig(opt *cloudprovider.CommonOption, cluster *eks.Cluster) (string, error) {
	restConfig, err := GenerateAwsRestConf(opt, cluster)
	if err != nil {
		return "", fmt.Errorf("GetClusterKubeConfig GenerateAwsRestConf failed, cluster=%s: %v",
			*cluster.Name, err)
	}

	cert, err := base64.StdEncoding.DecodeString(*cluster.CertificateAuthority.Data)
	if err != nil {
		return "", fmt.Errorf("GetClusterKubeConfig invalid certificate failed, cluster=%s: %v",
			*cluster.Name, err)
	}

	saToken, err := utils.GenerateSATokenByRestConfig(context.Background(), restConfig)
	if err != nil {
		return "", fmt.Errorf("getClusterKubeConfig generate k8s serviceaccount token failed,cluster=%s: %v",
			*cluster.Name, err)
	}

	typesConfig := &types.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: []types.NamedCluster{
			{
				Name: *cluster.Name,
				Cluster: types.ClusterInfo{
					Server:                   *cluster.Endpoint,
					CertificateAuthorityData: cert,
				},
			},
		},
		AuthInfos: []types.NamedAuthInfo{
			{
				Name: *cluster.Name,
				AuthInfo: types.AuthInfo{
					Token: saToken,
				},
			},
		},
		Contexts: []types.NamedContext{
			{
				Name: *cluster.Name,
				Context: types.Context{
					Cluster:  *cluster.Name,
					AuthInfo: *cluster.Name,
				},
			},
		},
		CurrentContext: *cluster.Name,
	}

	configByte, err := json.Marshal(typesConfig)
	if err != nil {
		return "", fmt.Errorf("GetClusterKubeConfig marsh kubeconfig failed, %v", err)
	}

	return encrypt.Encrypt(nil, string(configByte))
}

func taintTransEffect(ori string) string {
	switch ori {
	case "NoSchedule":
		return eks.TaintEffectNoSchedule
	case "PreferNoSchedule":
		return eks.TaintEffectPreferNoSchedule
	case "NoExecute":
		return eks.TaintEffectNoExecute
	}

	return ori
}

// MapToTaints converts a map of string-string to a slice of Taint
func MapToTaints(taints []*proto.Taint) []*Taint {
	result := make([]*Taint, 0)
	for _, v := range taints {
		result = append(result, &Taint{
			Key:    aws.String(v.Key),
			Value:  aws.String(v.Value),
			Effect: aws.String(taintTransEffect(v.Effect))})
	}

	return result
}

// MapToAwsTaints converts a map of string-string to a slice of aws Taint
func MapToAwsTaints(taints []*proto.Taint) []*eks.Taint {
	result := make([]*eks.Taint, 0)
	for _, v := range taints {
		key := v.Key
		value := v.Value
		effect := v.Effect
		result = append(result, &eks.Taint{Key: &key, Value: &value, Effect: &effect})
	}
	return result
}

func generateAwsCreateNodegroupInput(input *CreateNodegroupInput) *eks.CreateNodegroupInput {
	newInput := &eks.CreateNodegroupInput{
		AmiType:       input.AmiType,
		ClusterName:   input.ClusterName,
		CapacityType:  input.CapacityType,
		NodeRole:      input.NodeRole,
		NodegroupName: input.NodegroupName,
		ScalingConfig: func(c *NodegroupScalingConfig) *eks.NodegroupScalingConfig {
			if c == nil {
				return nil
			}
			return &eks.NodegroupScalingConfig{
				DesiredSize: c.DesiredSize,
				MaxSize:     c.MaxSize,
				MinSize:     c.MinSize,
			}
		}(input.ScalingConfig),
		Subnets: input.Subnets,
	}
	if input.LaunchTemplate != nil {
		newInput.LaunchTemplate = &eks.LaunchTemplateSpecification{
			Id:      input.LaunchTemplate.Id,
			Version: input.LaunchTemplate.Version,
		}
	}
	if len(input.Labels) != 0 {
		newInput.Labels = input.Labels
	}
	if len(input.Tags) != 0 {
		newInput.Tags = input.Tags
	}
	newInput.Taints = make([]*eks.Taint, 0)
	for _, v := range input.Taints {
		newInput.Taints = append(newInput.Taints, &eks.Taint{
			Key:    v.Key,
			Value:  v.Value,
			Effect: v.Effect,
		})
	}

	return newInput
}

// CreateTagSpecs creates tag specs
func CreateTagSpecs(instanceTags map[string]*string) []*ec2.LaunchTemplateTagSpecificationRequest {
	if len(instanceTags) == 0 {
		return nil
	}

	tags := make([]*ec2.Tag, 0)
	for key, value := range instanceTags {
		tags = append(tags, &ec2.Tag{Key: aws.String(key), Value: value})
	}
	return []*ec2.LaunchTemplateTagSpecificationRequest{
		{
			ResourceType: aws.String(ec2.ResourceTypeInstance),
			Tags:         tags,
		},
	}
}

// generateAwsCreateLaunchTemplateInput generate Aws CreateLaunchTemplateInput
func generateAwsCreateLaunchTemplateInput(input *CreateLaunchTemplateInput) *ec2.CreateLaunchTemplateInput {
	awsInput := &ec2.CreateLaunchTemplateInput{
		LaunchTemplateName: input.LaunchTemplateName,
		TagSpecifications:  generateAwsTagSpecs(input.TagSpecifications),
	}
	if input.LaunchTemplateData != nil {
		awsInput.LaunchTemplateData = &ec2.RequestLaunchTemplateData{
			KeyName:      input.LaunchTemplateData.KeyName,
			InstanceType: input.LaunchTemplateData.InstanceType,
			UserData:     input.LaunchTemplateData.UserData,
		}
	}
	return awsInput
}

func generateAwsTagSpecs(tagSpecs []*TagSpecification) []*ec2.TagSpecification {
	if tagSpecs == nil {
		return nil
	}
	awsTagSpecs := make([]*ec2.TagSpecification, 0)
	for _, t := range tagSpecs {
		awsTagSpecs = append(awsTagSpecs, &ec2.TagSpecification{
			ResourceType: t.ResourceType,
			Tags: func(t []*Tag) []*ec2.Tag {
				awsTags := make([]*ec2.Tag, 0)
				for _, v := range t {
					awsTags = append(awsTags, &ec2.Tag{Key: v.Key, Value: v.Value})
				}
				return awsTags
			}(t.Tags),
		})
	}
	return awsTagSpecs
}
