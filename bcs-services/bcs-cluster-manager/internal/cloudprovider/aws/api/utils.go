package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
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
		ClusterID: *cluster.Name,
	})
	if err != nil {
		return "", err
	}

	decodedCA, err := base64.StdEncoding.DecodeString(*cluster.CertificateAuthority.Data)
	if err != nil {
		return "", fmt.Errorf("GetClusterKubeConfig invalid certificate failed, cluster=%s: %w", *cluster.Name, err)
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
					CertificateAuthorityData: decodedCA,
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
