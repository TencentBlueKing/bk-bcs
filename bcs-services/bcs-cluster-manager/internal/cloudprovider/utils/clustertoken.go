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

// Package utils xxx
package utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	errs "github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
)

const (
	bcsNamespace           = "bcs-system"
	clusterAdmin           = "cluster-admin"
	bcsClusterManager      = "bcs-cluster-manager"
	clusterRoleBindingName = "bcs-system:cluster-manager"
)

// CheckClusterConnect check cluster connect by kubeConfig
func CheckClusterConnect(ctx context.Context, kubeConfig string) error {
	clientSet, err := clusterops.NewKubeClient(kubeConfig)
	if err != nil {
		return fmt.Errorf("CheckClusterConnect create clientset failed: %v", err)
	}

	version, err := clientSet.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("CheckClusterConnect serverVersion failed: %v", err)
	}

	blog.Infof("CheckClusterConnect %s", version.String())
	return nil
}

// GenerateSATokenByKubeConfig generates a serviceAccountToken
func GenerateSATokenByKubeConfig(ctx context.Context, kubeConfig string) (string, error) {
	clientSet, err := clusterops.NewKubeClient(kubeConfig)
	if err != nil {
		return "", fmt.Errorf("GenerateSATokenByKubeConfig create clientset failed: %v", err)
	}

	return GenerateServiceAccountToken(ctx, clientSet)
}

// GenerateSATokenByRestConfig generates a serviceAccountToken
func GenerateSATokenByRestConfig(ctx context.Context, config *rest.Config) (string, error) {
	clientSet, err := clusterops.NewKubeClientByRestConfig(config)
	if err != nil {
		return "", fmt.Errorf("GenerateSATokenByRestConfig create clientset failed: %v", err)
	}

	return GenerateServiceAccountToken(ctx, clientSet)
}

func createNamespace(ctx context.Context, clientSet kubernetes.Interface) error {
	_, err := clientSet.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: bcsNamespace,
		},
	}, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func createServiceAccount(ctx context.Context, clientSet kubernetes.Interface) (*v1.ServiceAccount, error) {
	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: bcsClusterManager,
		},
	}

	_, err := clientSet.CoreV1().ServiceAccounts(bcsNamespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("GenerateServiceAccountToken creating service account failed: %v", err)
	}

	return serviceAccount, nil
}

func createClusterRole(ctx context.Context, clientSet kubernetes.Interface) (*rbacv1.ClusterRole, error) {
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
	clusterAdminRole, err := clientSet.RbacV1().ClusterRoles().Get(ctx, clusterAdmin, metav1.GetOptions{})
	if err != nil {
		clusterAdminRole, err = clientSet.RbacV1().ClusterRoles().Create(ctx, adminRole, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("GenerateServiceAccountToken create admin role failed: %v", err)
		}
	}
	return clusterAdminRole, nil
}

func createClusterRoleBinding(ctx context.Context, clientSet kubernetes.Interface, sa *v1.ServiceAccount,
	cr *rbacv1.ClusterRole) error {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
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
	if _, err := clientSet.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding,
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

		secret, errCreate := CreateSecretForServiceAccount(ctx, clientset, serviceAccount)
		if errCreate != nil {
			return "", fmt.Errorf("GenerateServiceAccountToken create secret for service "+
				"account failed: %v", errCreate)
		}
		if token, ok := secret.Data["token"]; ok {
			return string(token), nil
		}
		start *= 2
	}

	return "", errs.New("GenerateServiceAccountToken fetch serviceAccountToken failed")
}

// GetServiceAccountToken get serviceAccount token
func GetServiceAccountToken(ctx context.Context, clientSet kubernetes.Interface, ns, name string) (string, error) {
	serviceAccount, err := clientSet.CoreV1().ServiceAccounts(ns).Get(ctx,
		name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("GetServiceAccountToken get service account failed: %v", err)
	}

	time.Sleep(5 * time.Second)
	// warning: secret may be empty when create serviceAccount successfully
	if len(serviceAccount.Secrets) == 0 {
		return "", fmt.Errorf("GetServiceAccountToken serviceAccount[%s] secret empty", name)
	}

	// get secret name
	secretName := serviceAccount.Secrets[0].Name
	secret, err := clientSet.CoreV1().Secrets(ns).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if len(secret.Data[v1.ServiceAccountTokenKey]) > 0 {
		return string(secret.Data[v1.ServiceAccountTokenKey]), nil
	}

	return "", fmt.Errorf("GetServiceAccountToken serviceAccount[%s] secret token empty", name)
}

// GetSecretForServiceAccount gets Secret for the provided Service Account
func GetSecretForServiceAccount(ctx context.Context, clientSet kubernetes.Interface, sa *v1.ServiceAccount) (*v1.Secret,
	error) {
	secretClient := clientSet.CoreV1().Secrets(sa.Namespace)
	if len(sa.Secrets) == 0 {
		return nil, errs.New("GetSecretForServiceAccount  serviceAccount secret is nil")
	}
	secret, err := secretClient.Get(ctx, sa.Secrets[0].Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return secret, nil
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("get secret token timeout: %v", ctx.Err())
		case <-time.Tick(2 * time.Second): // nolint
			if len(secret.Data[v1.ServiceAccountTokenKey]) > 0 {
				return secret, nil
			}

			secret, err = secretClient.Get(ctx, secretName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
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
