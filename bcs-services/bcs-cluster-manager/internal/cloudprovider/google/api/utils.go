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

package api

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"

	errs "github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	bcsNamespace              = "bcs-system"
	clusterAdmin              = "cluster-admin"
	bcsCusterManager          = "bcs-cluster-manager"
	newClusterRoleBindingName = "bcs-system-cluster-manager-clusterRoleBinding"
)

// GkeServiceAccount for GKE service account
type GkeServiceAccount struct {
	Type                string `json:"type"`
	ProjectID           string `json:"project_id"`
	PrivateKeyID        string `json:"private_key_id"`
	PrivateKey          string `json:"private_key"`
	ClientEmail         string `json:"client_email"`
	ClientID            string `json:"client_id"`
	AuthURI             string `json:"auth_uri"`
	TokenURI            string `json:"token_uri"`
	AuthProviderCertURL string `json:"auth_provider_x509_cert_url"`
	ClientCertURL       string `json:"client_x509_cert_url"`
}

// GetTokenSource gets token source from provided sa credential
func GetTokenSource(ctx context.Context, credential string) (oauth2.TokenSource, error) {
	ts, err := google.CredentialsFromJSON(ctx, []byte(credential), container.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("GetTokenSource failed: %v", err)
	}
	return ts.TokenSource, nil
}

// GenerateSAToken generates a serviceAccountToken
func GenerateSAToken(ctx context.Context, restConfig *rest.Config) (string, error) {
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return "", fmt.Errorf("GenerateSAToken create clientset failed: %v", err)
	}

	return GenerateServiceAccountToken(ctx, clientSet)
}

func createNamespace(ctx context.Context, clientset kubernetes.Interface) error {
	_, err := clientset.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: bcsNamespace,
		},
	}, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func createServiceAccount(ctx context.Context, clientset kubernetes.Interface) (*v1.ServiceAccount, error) {
	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: bcsCusterManager,
		},
	}

	_, err := clientset.CoreV1().ServiceAccounts(bcsNamespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("GenerateServiceAccountToken creating service account failed: %v", err)
	}
	return serviceAccount, nil
}

func createClusterRole(ctx context.Context, clientset kubernetes.Interface) (*rbacv1.ClusterRole, error) {
	adminRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterAdmin,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			{
				NonResourceURLs: []string{"*"},
				Verbs:           []string{"*"},
			},
		},
	}
	clusterAdminRole, err := clientset.RbacV1().ClusterRoles().Get(ctx, clusterAdmin, metav1.GetOptions{})
	if err != nil {
		clusterAdminRole, err = clientset.RbacV1().ClusterRoles().Create(ctx, adminRole, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("GenerateServiceAccountToken create admin role failed: %v", err)
		}
	}
	return clusterAdminRole, nil
}

func createClusterRoleBinding(ctx context.Context, clientset kubernetes.Interface, sa *v1.ServiceAccount,
	cr *rbacv1.ClusterRole) error {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: newClusterRoleBindingName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      sa.Name,
				Namespace: bcsNamespace,
				APIGroup:  v1.GroupName,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     cr.Name,
			APIGroup: rbacv1.GroupName,
		},
	}
	if _, err := clientset.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding,
		metav1.CreateOptions{}); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("GenerateServiceAccountToken create role bindings failed: %v", err)
	}
	return nil
}

// GenerateServiceAccountToken generates a serviceAccountToken for clusterAdmin given a rest clientset
func GenerateServiceAccountToken(ctx context.Context, clientset kubernetes.Interface) (string, error) {
	err := createNamespace(ctx, clientset)
	if err != nil {
		return "", err
	}

	serviceAccount, err := createServiceAccount(ctx, clientset)
	if err != nil {
		return "", err
	}

	clusterAdminRole, err := createClusterRole(ctx, clientset)
	if err != nil {
		return "", err
	}

	err = createClusterRoleBinding(ctx, clientset, serviceAccount, clusterAdminRole)
	if err != nil {
		return "", err
	}

	start := time.Millisecond * 250
	for i := 0; i < 5; i++ {
		time.Sleep(start)
		if serviceAccount, err = clientset.CoreV1().ServiceAccounts(bcsNamespace).Get(ctx,
			serviceAccount.Name, metav1.GetOptions{}); err != nil {
			return "", fmt.Errorf("GenerateServiceAccountToken get service account failed: %v", err)
		}
		secret, err := CreateSecretForServiceAccount(ctx, clientset, serviceAccount)
		if err != nil {
			return "", fmt.Errorf("GenerateServiceAccountToken create secret for service account failed: %v", err)
		}
		if token, ok := secret.Data["token"]; ok {
			return string(token), nil
		}
		start *= 2
	}

	return "", errs.New("GenerateServiceAccountToken fetch serviceAccountToken failed")
}

// CreateSecretForServiceAccount creates a service-account-token Secret for the provided Service Account.
// If the secret already exists, the existing one is returned.
func CreateSecretForServiceAccount(ctx context.Context, clientSet kubernetes.Interface, sa *v1.ServiceAccount) (
	*v1.Secret, error) {
	secretName := ServiceAccountSecretName(sa)
	secretClient := clientSet.CoreV1().Secrets(sa.Namespace)
	secret, err := secretClient.Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		sc := SecretTemplate(sa)
		secret, err = secretClient.Create(ctx, sc, metav1.CreateOptions{})
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return nil, err
			}
			secret, err = secretClient.Get(ctx, secretName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
		}
	}
	if len(secret.Data[v1.ServiceAccountTokenKey]) > 0 {
		return secret, nil
	}
	blog.Errorf("CreateSecretForServiceAccount: waiting for secret [%s] to be populated with token", secretName)
	for {
		if len(secret.Data[v1.ServiceAccountTokenKey]) > 0 {
			return secret, nil
		}
		time.Sleep(2 * time.Second)
		secret, err = secretClient.Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}
}

// SecretTemplate generate a template of service-account-token Secret for the provided Service Account.
func SecretTemplate(sa *v1.ServiceAccount) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceAccountSecretName(sa),
			Namespace: sa.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "v1",
					Kind:       "ServiceAccount",
					Name:       sa.Name,
					UID:        sa.UID,
				},
			},
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": sa.Name,
			},
		},
		Type: v1.SecretTypeServiceAccountToken,
	}
}

// ServiceAccountSecretName returns the secret name for the given Service Account.
func ServiceAccountSecretName(sa *v1.ServiceAccount) string {
	return SafeConcatName(sa.Name, "token")
}

// SafeConcatName for safe concat name
func SafeConcatName(name ...string) string {
	fullPath := strings.Join(name, "-")
	if len(fullPath) < 64 {
		return fullPath
	}
	digest := sha256.Sum256([]byte(fullPath))
	// since we cut the string in the middle, the last char may not be compatible with what is expected in k8s
	// we are checking and if necessary removing the last char
	c := fullPath[56]
	if 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return fullPath[0:57] + "-" + hex.EncodeToString(digest[0:])[0:5]
	}

	return fullPath[0:56] + "-" + hex.EncodeToString(digest[0:])[0:6]
}

// GetClusterKubeConfig get cloud cluster's kube config
func GetClusterKubeConfig(ctx context.Context, saSecret, gkeProjectID, region, clusterName string) (string, error) {
	client, err := GetContainerServiceClient(ctx, saSecret)
	if err != nil {
		return "", err
	}
	// Get the kube cluster in given project.
	parent := "projects/" + gkeProjectID + "/locations/" + region + "/clusters/" + clusterName
	gkeCluster, err := client.Projects.Locations.Clusters.Get(parent).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("GetClusterKubeConfig list clusters failed, project=%s: %w", gkeProjectID, err)
	}
	name := fmt.Sprintf("%s_%s_%s", gkeProjectID, gkeCluster.Location, gkeCluster.Name)
	cert, err := base64.StdEncoding.DecodeString(gkeCluster.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return "", fmt.Errorf("GetClusterKubeConfig invalid certificate failed, cluster=%s: %w", name, err)
	}

	restConfig := &rest.Config{
		TLSClientConfig: rest.TLSClientConfig{
			CAData: cert,
		},
		Host: "https://" + gkeCluster.Endpoint,
		AuthProvider: &api.AuthProviderConfig{
			Name: GoogleAuthPlugin,
			Config: map[string]string{
				"scopes":      "https://www.googleapis.com/auth/cloud-platform",
				"credentials": saSecret,
			},
		},
	}

	var saToken string
	saToken, err = GenerateSAToken(ctx, restConfig)
	if err != nil {
		return "", fmt.Errorf("getClusterKubeConfig generate k8s serviceaccount token failed, project=%s cluster=%s: %w",
			gkeProjectID, clusterName, err)
	}

	typesConfig := &types.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: []types.NamedCluster{
			{
				Name: name,
				Cluster: types.ClusterInfo{
					Server:                   "https://" + gkeCluster.Endpoint,
					CertificateAuthorityData: cert,
				},
			},
		},
		AuthInfos: []types.NamedAuthInfo{
			{
				Name: name,
				AuthInfo: types.AuthInfo{
					Token: saToken,
				},
			},
		},
		Contexts: []types.NamedContext{
			{
				Name: name,
				Context: types.Context{
					Cluster:  name,
					AuthInfo: name,
				},
			},
		},
		CurrentContext: name,
	}

	configByte, err := json.Marshal(typesConfig)
	if err != nil {
		return "", fmt.Errorf("GetClusterKubeConfig marsh kubeconfig failed, %v", err)
	}
	return base64.StdEncoding.EncodeToString(configByte), nil
}
