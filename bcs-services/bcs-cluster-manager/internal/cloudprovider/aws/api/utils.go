package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"time"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

const (
	// ResourceTypeLaunchTemplate is a ResourceType enum value
	ResourceTypeLaunchTemplate = "launch-template"
)

var (
	// DefaultUserDataHeader default user data header for creating launch template userdata
	DefaultUserDataHeader = "MIME-Version: 1.0\nContent-Type: multipart/mixed; boundary=\"==MYBOUNDARY==\"\n\n" +
		"-==MYBOUNDARY==\nContent-Type: text/x-shellscript; charset=\"us-ascii\"\n\n"
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

// GetClusterKubeConfig constructs the cluster kubeConfig
func GetClusterKubeConfig(opt *cloudprovider.CommonOption, cluster *eks.Cluster) (string, error) {
	sess, err := NewSession(opt)
	if err != nil {
		return "", err
	}
	generator, err := token.NewGenerator(false, false)
	if err != nil {
		return "", err
	}

	awsToken, err := generator.GetWithOptions(&token.GetTokenOptions{
		Session:   sess,
		ClusterID: *cluster.Name,
	})
	if err != nil {
		return "", err
	}

	decodedCA, err := base64.StdEncoding.DecodeString(*cluster.CertificateAuthority.Data)
	if err != nil {
		return "", err
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

	cert, err := base64.StdEncoding.DecodeString(*cluster.CertificateAuthority.Data)
	if err != nil {
		return "", fmt.Errorf("GetClusterKubeConfig invalid certificate failed, cluster=%s: %w", *cluster.Name, err)
	}

	saToken, err := cloudprovider.GenerateSAToken(restConfig)
	if err != nil {
		return "", fmt.Errorf("getClusterKubeConfig generate k8s serviceaccount token failed,cluster=%s: %w",
			*cluster.Name, err)
	}

	typesConfig := &types.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: []types.NamedCluster{
			{
				Name: *cluster.Name,
				Cluster: types.ClusterInfo{
					Server:                   "https://" + *cluster.Endpoint,
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

	return base64.StdEncoding.EncodeToString(configByte), nil
}

// MapToTaints converts a map of string-string to a slice of Taint
func MapToTaints(taints []*proto.Taint) []*Taint {
	result := make([]*Taint, 0)
	for _, v := range taints {
		key := v.Key
		value := v.Value
		effect := v.Effect
		result = append(result, &Taint{Key: &key, Value: &value, Effect: &effect})
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
		ClusterName:   input.ClusterName,
		NodegroupName: input.NodegroupName,
		//AmiType:       input.AmiType,
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
		NodeRole:     input.NodeRole,
		Labels:       input.Labels,
		Tags:         input.Tags,
		CapacityType: input.CapacityType,
	}
	newInput.Taints = make([]*eks.Taint, 0)
	for _, v := range input.Taints {
		newInput.Taints = append(newInput.Taints, &eks.Taint{
			Key:    v.Key,
			Value:  v.Value,
			Effect: v.Effect,
		})
	}
	if input.LaunchTemplate != nil {
		newInput.LaunchTemplate = &eks.LaunchTemplateSpecification{
			Id:      input.LaunchTemplate.Id,
			Name:    input.LaunchTemplate.Name,
			Version: input.LaunchTemplate.Version,
		}
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

func generateAwsCreateLaunchTemplateInput(input *CreateLaunchTemplateInput) *ec2.CreateLaunchTemplateInput {
	awsInput := &ec2.CreateLaunchTemplateInput{
		LaunchTemplateName: input.LaunchTemplateName,
		TagSpecifications:  generateAwsTagSpecs(input.TagSpecifications),
	}
	if input.LaunchTemplateData != nil {
		awsInput.LaunchTemplateData = &ec2.RequestLaunchTemplateData{
			ImageId:      input.LaunchTemplateData.ImageId,
			InstanceType: input.LaunchTemplateData.InstanceType,
			KeyName:      input.LaunchTemplateData.KeyName,
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
