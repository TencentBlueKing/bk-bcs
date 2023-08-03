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

package secret

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	vault "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/vaultplugin-server/common"
)

const (
	retryDelay = 3 * time.Second
	retryNum   = 4
)

// 递增重试
func incrementalBackoff(n uint, err error, config *retry.Config) time.Duration {
	return time.Duration(n+1) * time.Second
}

// VaultSecretManager vault client
type VaultSecretManager struct {
	option  *Options
	client  *vault.Client
	kclient *kubernetes.Clientset
}

// Init create vault k8s client
func (m *VaultSecretManager) Init() error {
	// init vault client
	config := vault.DefaultConfig()
	config.Address = m.option.Endpoints
	config.ConfigureTLS(&vault.TLSConfig{
		CAPath: common.GetVaultCa(),
	})
	client, err := vault.NewClient(config)
	if err != nil {
		return errors.Wrapf(err, "unable to initialize Vault client")
	}
	client.SetToken(m.option.Token)
	m.client = client

	// init k8s clientset
	kubernetesConfig, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrapf(err, "unable to get kubernetes config in cluster")
	}
	clientset, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		return errors.Wrapf(err, "unable to initialize kubernetes client")
	}
	m.kclient = clientset

	return nil
}

// Stop control interface
func (m *VaultSecretManager) Stop() {
}

func (m *VaultSecretManager) hasInitProject(project string) bool {
	ml, err := m.client.Sys().ListMounts()
	if err != nil {
		blog.Warnf("listMount failed when check initSecret, err: %s", err)
		return true
	}
	for mount := range ml {
		// listMount函数返回map的key为mount的name，同时会在后面加上/
		if mount == fmt.Sprintf("%s/", project) {
			return true
		}
	}
	return false
}

func (m *VaultSecretManager) deferInit(project string) {
	// check mount
	ml, err := m.client.Sys().ListMounts()
	if err != nil {
		blog.Infof("listMount failed when check initSecret, err: %s", err)
	}
	for mount := range ml {
		if mount == fmt.Sprintf("%s/", project) {
			err = m.client.Sys().Unmount(project)
			if err != nil {
				blog.Warnf("umount failed when deferInit, err: %s", err)
			}
		}
	}

	// check policy
	_, err = m.client.Sys().GetPolicy(project)
	if err != nil {
		blog.Infof("getPolicy failed when deferInit, if err is not found, skip... err: %s", err)
	}
	err = m.client.Sys().DeletePolicy(project)
	if err != nil {
		blog.Warnf("delPolicy failed when deferInit, err: %s", err)
	}
}

// InitProject mount, policy, token, secret
func (m *VaultSecretManager) InitProject(project string) error {
	if m.hasInitProject(project) {
		blog.Infof("Project mount [%s] exists, skip secrets init", project)
		return nil
	}

	// create kv, secrets volume for project root path
	kvMount := &vault.MountInput{
		Type:        "kv",
		Description: fmt.Sprintf("project [%s] for SecretServer", project),
		Options: map[string]string{
			"version": common.VaultVersion,
		},
	}
	err := m.client.Sys().Mount(project, kvMount)
	if err != nil {
		m.deferInit(project)
		return errors.Wrapf(err, "create mount --path=%s err", project)
	}

	err = m.client.Sys().PutPolicy(project, common.GetVaultProjectRule(project))
	if err != nil {
		// Atomic operations, rollback.
		blog.Infof("project[%s] initSecret failed, err: %s.", project, err)
		m.deferInit(project)
		return errors.Wrapf(err, "create policy %s err", project)
	}

	// create token
	ops := &vault.TokenCreateRequest{
		Policies: []string{project},
	}
	token, err := m.client.Auth().Token().Create(ops)
	if err != nil {
		blog.Infof("project[%s] initSecret failed, err: %s.", project, err)
		m.deferInit(project)
		return errors.Wrapf(err, "create token %s err", project)
	}

	// create secret
	ksecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.GetVaultSecretName(project),
			Namespace: common.GetVaultSecretNamespace(),
		},
		StringData: map[string]string{
			"VAULT_ADDR":    common.GetVaultAddr(),
			"VAULT_TOKEN":   token.Auth.ClientToken,
			"AVP_TYPE":      common.GetSecretType(),
			"AVP_AUTH_TYPE": "token",
			"VAULT_CACERT":  common.GetVaultCa(),
		},
	}
	_, err = m.kclient.CoreV1().Secrets(common.GetVaultSecretNamespace()).Create(context.Background(), ksecret, metav1.CreateOptions{})
	if err != nil {
		m.deferInit(project)
		blog.Errorf("project[%s] initSecret failed, err: %s.", project, err)
		return errors.Wrapf(err, "create Secret %s err", ksecret.Name)
	}

	return nil
}

// GetSecretAnnotation get init secret info for gitops-manager
func (m *VaultSecretManager) GetSecretAnnotation(project string) string {
	return common.GetVaultSecForProAnno(project)
}

// GetSecret interface for get secret
func (m *VaultSecretManager) GetSecret(ctx context.Context, req *SecretRequest) (map[string]interface{}, error) {
	var sec *vault.KVSecret
	var err error

	err = retry.Do(
		func() error {
			sec, err = m.client.KVv2(req.Project).Get(ctx, req.Path)
			if err != nil {
				return err
			}
			return nil
		},
		retry.Attempts(retryNum),
		retry.DelayType(incrementalBackoff),
	)
	if err != nil {
		return nil, err
	}
	return sec.Data, nil
}

// GetMetadata interface for get metadata
func (m *VaultSecretManager) GetMetadata(ctx context.Context, req *SecretRequest) (*SecretMetadata, error) {
	meta, err := m.client.KVv2(req.Project).GetMetadata(ctx, req.Path)
	if err != nil {
		return nil, err
	}

	s, err := json.Marshal(meta)
	if err != nil {
		return nil, errors.Wrapf(err, "json marshal failed.")
	}
	data := &SecretMetadata{}
	if err := json.Unmarshal(s, data); err != nil {
		blog.Infof("GetMetadata meta []byte is: %s", s)
		return nil, err
	}

	return data, nil
}

// ListSecret interface for list secret
func (m *VaultSecretManager) ListSecret(ctx context.Context, req *SecretRequest) ([]string, error) {
	pathToRead := fmt.Sprintf("%s/metadata/%s", req.Project, req.Path)

	secrets, err := m.client.Logical().ListWithContext(ctx, pathToRead)
	if err != nil {
		return nil, err
	}

	// 不存在的项目或者目录会返回nil
	if secrets == nil {
		return nil, nil
	}

	fs := make([]string, 0, len(secrets.Data))
	for _, val := range secrets.Data["keys"].([]interface{}) {
		fs = append(fs, val.(string))
	}
	return fs, nil
}

// CreateSecret interface
func (m *VaultSecretManager) CreateSecret(ctx context.Context, req *SecretRequest) error {
	_, err := m.client.KVv2(req.Project).Put(ctx, req.Path, req.Data)
	if err != nil {
		return err
	}
	return nil
}

// UpdateSecret interface
func (m *VaultSecretManager) UpdateSecret(ctx context.Context, req *SecretRequest) error {
	_, err := m.client.KVv2(req.Project).Put(ctx, req.Path, req.Data)
	if err != nil {
		return err
	}
	return nil
}

// DeleteSecret interface
func (m *VaultSecretManager) DeleteSecret(ctx context.Context, req *SecretRequest) error {
	return m.client.KVv2(req.Project).DeleteMetadata(ctx, req.Path)
}

// GetSecretWithVersion interface
func (m *VaultSecretManager) GetSecretWithVersion(ctx context.Context, req *SecretRequest, version int) (map[string]interface{}, error) {
	secret, err := m.client.KVv2(req.Project).GetVersion(ctx, req.Path, version)
	if err != nil {
		return nil, err
	}
	return secret.Data, nil
}

// GetVersionsAsList interface
func (m *VaultSecretManager) GetVersionsAsList(ctx context.Context, req *SecretRequest) ([]*SecretVersion, error) {
	version, err := m.client.KVv2(req.Project).GetVersionsAsList(ctx, req.Path)
	if err != nil {
		return nil, err
	}

	s, err := json.Marshal(version)
	if err != nil {
		return nil, errors.Wrapf(err, "json marshal failed.")
	}
	var data []*SecretVersion
	if err = json.Unmarshal(s, &data); err != nil {
		blog.Infof("GetMetadata meta []byte is: %s", s)
		return nil, err
	}

	return data, nil
}

// Rollback interface
func (m *VaultSecretManager) Rollback(ctx context.Context, req *SecretRequest, version int) error {
	_, err := m.client.KVv2(req.Project).Rollback(ctx, req.Path, version)
	if err != nil {
		return err
	}
	return nil
}

// DeleteVersion interface
func (m *VaultSecretManager) DeleteVersion(ctx context.Context, req *SecretRequest, version []int) error {
	return m.client.KVv2(req.Project).DeleteVersions(ctx, req.Path, version)
}

// UnDeleteVersion interface
func (m *VaultSecretManager) UnDeleteVersion(ctx context.Context, req *SecretRequest, version []int) error {
	return m.client.KVv2(req.Project).Undelete(ctx, req.Path, version)
}

// DestroyVersion interface
func (m *VaultSecretManager) DestroyVersion(ctx context.Context, req *SecretRequest, version []int) error {
	return m.client.KVv2(req.Project).Destroy(ctx, req.Path, version)
}
