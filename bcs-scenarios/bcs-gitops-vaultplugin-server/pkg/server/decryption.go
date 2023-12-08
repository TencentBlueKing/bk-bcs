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

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/auth/vault"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/config"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/kube"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/types"
	argoappv1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/pkg/argoplugin"
)

type decryptionManifestRequest struct {
	Project   string   `json:"project"`
	Manifests []string `json:"manifests"`
}

type decryptionManifestResponse struct {
	Manifests []string `json:"manifests"`
	Message   string   `json:"message"`
}

func (s *Server) routerDecryptManifest(w http.ResponseWriter, r *http.Request) {
	req := new(decryptionManifestRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.responseError(r, w, http.StatusBadRequest,
			errors.Wrapf(err, "decryption manifest decode request body failed"))
		return
	}
	if req.Manifests == nil || req.Project == "" {
		s.responseError(r, w, http.StatusBadRequest,
			errors.Errorf("request 'manifests' or 'project' required"))
		return
	}
	projectName := req.Project
	secretKey, err := s.getSecretKeyForProject(r.Context(), projectName)
	if err != nil {
		s.responseError(r, w, http.StatusBadRequest,
			errors.Errorf("get secret key for project '%s' failed", projectName))
		return
	}
	res, err := s.decryptVaultSecret(r.Context(), projectName, secretKey, req.Manifests)
	if err != nil {
		s.responseError(r, w, http.StatusInternalServerError,
			errors.Wrapf(err, "descrypt vault secret for project '%s/%s' failed", projectName, secretKey))
		return
	}
	resp := &decryptionManifestResponse{
		Manifests: res,
	}
	// nolint
	bs, _ := json.Marshal(resp)
	s.responseDirect(w, bs)
}

func (s *Server) getSecretKeyForProject(ctx context.Context, project string) (string, error) {
	argoProj, err := s.argoStore.GetProject(ctx, project)
	if err != nil {
		return "", errors.Wrapf(err, "get project '%s' failed", project)
	}
	if argoProj == nil {
		return "", errors.Errorf("project '%s' not exist", project)
	}
	annot := argoProj.GetAnnotations()
	if annot == nil {
		return "", errors.Errorf("project '%s' annotation is empty", project)
	}
	secretKey, ok := annot[common.SecretKey]
	if !ok {
		return "", errors.Errorf("project '%s' not have secret key", project)
	}
	return secretKey, nil
}

func (s *Server) buildArgoVaultPlugin(projectName, secretName string) (*argoplugin.VaultArgoPlugin, error) {
	apiClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, errors.Wrapf(err, "create vault client failed")
	}
	v := viper.New()
	cmdConfig, err := config.New(v, &config.Options{
		SecretName: secretName,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "create argo vault plugin config failed")
	}
	argoVaultPlugin := argoplugin.NewVaultArgoPluginBackend(&vault.TokenAuth{}, apiClient,
		v.GetString(types.EnvAvpKvVersion),
		s.secretManager, projectName)
	cmdConfig.Backend = argoVaultPlugin
	if err = cmdConfig.Backend.Login(); err != nil {
		return nil, errors.Wrapf(err, "argo vault plugin login failed")
	}
	return argoVaultPlugin, nil
}

func (s *Server) decryptVaultSecret(ctx context.Context, projectname string, secretName string,
	manifestsstring []string) ([]string, error) {
	argoVaultPlugin, err := s.buildArgoVaultPlugin(projectname, secretName)
	if err != nil {
		return nil, errors.Wrapf(err, "build argo vault plugin failed")
	}

	v := viper.New()
	var pathValidation *regexp.Regexp
	if rexp := v.GetString(types.EnvPathValidation); rexp != "" {
		pathValidation, err = regexp.Compile(rexp)
		if err != nil {
			return nil, errors.Errorf("%s is not a valid regular expression: %s", rexp, err)
		}
	}
	res := make([]string, 0, len(manifestsstring))
	for _, m := range manifestsstring {
		var obj *unstructured.Unstructured
		obj, err = argoappv1.UnmarshalToUnstructured(m)
		if err != nil {
			return nil, errors.Wrapf(err, "unmashal manifest '%s' failed", m)
		}
		manifest := *obj

		var template *kube.Template
		template, err = kube.NewTemplate(manifest, argoVaultPlugin, pathValidation)
		if err != nil {
			return nil, errors.Wrapf(err, "create template for manifest '%s' failed", m)
		}

		annotations := manifest.GetAnnotations()
		if annotations == nil || annotations[types.AVPIgnoreAnnotation] != "true" {
			if err = template.Replace(); err != nil {
				return nil, errors.Wrapf(err, "template replace for manifest %s.%s failed",
					manifest.GetNamespace(), manifest.GetName())
			}
		} else {
			blog.Warnf("RequestID[%s] skipping %s.%s because annotation '%s' is true",
				requestID(ctx), manifest.GetNamespace(), manifest.GetName(), types.AVPIgnoreAnnotation)
		}
		var jsonData []byte
		jsonData, err = json.Marshal(template.TemplateData)
		if err != nil {
			return nil, errors.Wrapf(err, "marshal template for manifest %s.%s failed",
				manifest.GetNamespace(), manifest.GetName())
		}
		res = append(res, string(jsonData))
	}
	return res, nil
}
