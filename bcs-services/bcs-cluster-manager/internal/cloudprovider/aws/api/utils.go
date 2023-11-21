package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"time"

	cutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

// GetClusterKubeConfig get eks cluster kebeconfig
func GetClusterKubeConfig(sess *session.Session, cluster *eks.Cluster) (string, error) {
	generator, err := token.NewGenerator(false, false)
	if err != nil {
		return "", err
	}

	awsToken, err := generator.GetWithOptions(&token.GetTokenOptions{
		Session:   sess,
		ClusterID: aws.StringValue(cluster.Name),
	})
	if err != nil {
		return "", err
	}

	decodedCA, err := base64.StdEncoding.DecodeString(aws.StringValue(cluster.CertificateAuthority.Data))
	if err != nil {
		return "", fmt.Errorf("GetClusterKubeConfig invalid certificate failed, cluster=%s: %w",
			aws.StringValue(cluster.Name), err)
	}

	restConfig := &rest.Config{
		Host: aws.StringValue(cluster.Endpoint),
		TLSClientConfig: rest.TLSClientConfig{
			CAData: decodedCA,
		},
		BearerToken: awsToken.Token,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	saToken, err := cutils.GenerateSATokenByRestConfig(context.Background(), restConfig)
	if err != nil {
		return "", fmt.Errorf("getClusterKubeConfig generate k8s serviceaccount token failed,cluster=%s: %w",
			aws.StringValue(cluster.Name), err)
	}

	typesConfig := &types.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: []types.NamedCluster{
			{
				Name: aws.StringValue(cluster.Name),
				Cluster: types.ClusterInfo{
					Server:                   aws.StringValue(cluster.Endpoint),
					CertificateAuthorityData: decodedCA,
				},
			},
		},
		AuthInfos: []types.NamedAuthInfo{
			{
				Name: aws.StringValue(cluster.Name),
				AuthInfo: types.AuthInfo{
					Token: saToken,
				},
			},
		},
		Contexts: []types.NamedContext{
			{
				Name: aws.StringValue(cluster.Name),
				Context: types.Context{
					Cluster:  aws.StringValue(cluster.Name),
					AuthInfo: aws.StringValue(cluster.Name),
				},
			},
		},
		CurrentContext: aws.StringValue(cluster.Name),
	}

	configByte, err := json.Marshal(typesConfig)
	if err != nil {
		return "", fmt.Errorf("GetClusterKubeConfig marsh kubeconfig failed, %v", err)
	}

	return base64.StdEncoding.EncodeToString(configByte), nil
}
