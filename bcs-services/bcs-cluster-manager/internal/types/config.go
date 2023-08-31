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

package types

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
)

// Config holds the information needed to build connect to remote kubernetes clusters as a given user
type Config struct {
	// Legacy field from pkg/api/types.go TypeMeta.
	// +optional
	Kind string `json:"kind,omitempty"`
	// Legacy field from pkg/api/types.go TypeMeta.
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`
	// Preferences holds general information to be use for cli interactions
	Preferences Preferences `json:"preferences"`
	// Clusters is a map of referencable names to cluster configs
	Clusters []NamedCluster `json:"clusters"`
	// AuthInfos is a map of referencable names to user configs
	AuthInfos []NamedAuthInfo `json:"users"`
	// Contexts is a map of referencable names to context configs
	Contexts []NamedContext `json:"contexts"`
	// CurrentContext is the name of the context that you would like to use by default
	CurrentContext string `json:"current-context"`
	// Extensions holds additional information. This is useful for extenders so that reads and writes don't clobber unknown fields
	// +optional
	Extensions []NamedExtension `json:"extensions,omitempty"`
}

// Preferences holds general information to be use for cli interactions
type Preferences struct {
	// +optional
	Colors bool `json:"colors,omitempty"`
	// Extensions holds additional information. This is useful for extenders
	// so that reads and writes don't clobber unknown fields
	Extensions []NamedExtension `json:"extensions,omitempty"`
}

// ClusterInfo contains information about how to communicate with a kubernetes cluster
type ClusterInfo struct {
	// Server is the address of the kubernetes cluster (https://hostname:port).
	Server string `json:"server"`
	// InsecureSkipTLSVerify skips the validity check for the server's certificate.
	//	This will make your HTTPS connections insecure.
	InsecureSkipTLSVerify bool `json:"insecure-skip-tls-verify,omitempty"`
	// CertificateAuthority is the path to a cert file for the certificate authority.
	CertificateAuthority string `json:"certificate-authority,omitempty"`
	// CertificateAuthorityData contains PEM-encoded certificate authority certificates. Overrides CertificateAuthority
	CertificateAuthorityData []byte `json:"certificate-authority-data,omitempty"`
	// Extensions holds additional information. This is useful for extenders
	// so that reads and writes don't clobber unknown fields
	Extensions []NamedExtension `json:"extensions,omitempty"`
}

// AuthInfo contains information that describes identity information.  This is use to tell the kubernetes cluster who you are.
type AuthInfo struct {
	// ClientCertificate is the path to a client cert file for TLS.
	// +optional
	ClientCertificate string `json:"client-certificate,omitempty"`
	// ClientCertificateData contains PEM-encoded data from a client cert file for TLS. Overrides ClientCertificate
	// +optional
	ClientCertificateData []byte `json:"client-certificate-data,omitempty"`
	// ClientKey is the path to a client key file for TLS.
	// +optional
	ClientKey string `json:"client-key,omitempty"`
	// ClientKeyData contains PEM-encoded data from a client key file for TLS. Overrides ClientKey
	// +optional
	ClientKeyData []byte `json:"client-key-data,omitempty"`
	// Token is the bearer token for authentication to the kubernetes cluster.
	// +optional
	Token string `json:"token,omitempty"`
	// TokenFile is a pointer to a file that contains a bearer token (as described above).
	// If both Token and TokenFile are present, Token takes precedence.
	TokenFile string `json:"tokenFile,omitempty"`
	// Impersonate is the username to imperonate.  The name matches the flag.
	// +optional
	Impersonate string `json:"as,omitempty"`
	// ImpersonateGroups is the groups to imperonate.
	// +optional
	ImpersonateGroups []string `json:"as-groups,omitempty"`
	// ImpersonateUserExtra contains additional information for impersonated user.
	// +optional
	ImpersonateUserExtra map[string][]string `json:"as-user-extra,omitempty"`
	// Username is the username for basic authentication to the kubernetes cluster.
	// +optional
	Username string `json:"username,omitempty"`
	// Password is the password for basic authentication to the kubernetes cluster.
	// +optional
	Password string `json:"password,omitempty"`
	// AuthProvider specifies a custom authentication plugin for the kubernetes cluster.
	// +optional
	AuthProvider *AuthProviderConfig `json:"auth-provider,omitempty"`
	// Exec specifies a custom exec-based authentication plugin for the kubernetes cluster.
	// +optional
	Exec *ExecConfig `json:"exec,omitempty"`
	// Extensions holds additional information. This is useful for extenders so that reads
	// and writes don't clobber unknown fields
	Extensions []NamedExtension `json:"extensions,omitempty"`
}

// Context is a tuple of references to a cluster
type Context struct {
	// Cluster is the name of the cluster for this context
	Cluster string `json:"cluster"`
	// AuthInfo is the name of the authInfo for this context
	AuthInfo string `json:"user"`
	// Namespace is the default namespace to use on unspecified requests
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// Extensions holds additional information. This is useful for extenders
	// so that reads and writes don't clobber unknown fields
	Extensions []NamedExtension `json:"extensions,omitempty"`
}

// NamedCluster relates nicknames to cluster information
type NamedCluster struct {
	// Name is the nickname for this Cluster
	Name string `json:"name"`
	// ClusterInfo holds the cluster information
	Cluster ClusterInfo `json:"cluster"`
}

// NamedContext relates nicknames to context information
type NamedContext struct {
	// Name is the nickname for this Context
	Name string `json:"name"`
	// Context holds the context information
	Context Context `json:"context"`
}

// NamedAuthInfo relates nicknames to auth information
type NamedAuthInfo struct {
	// Name is the nickname for this AuthInfo
	Name string `json:"name"`
	// AuthInfo holds the auth information
	AuthInfo AuthInfo `json:"user"`
}

// NamedExtension relates nicknames to extension information
type NamedExtension struct {
	// Name is the nickname for this Extension
	Name string `json:"name"`
	// Extension holds the extension information
	Extension interface{} `json:"extension"`
}

// AuthProviderConfig holds the configuration for a specified auth provider.
type AuthProviderConfig struct {
	Name   string            `json:"name"`
	Config map[string]string `json:"config"`
}

// ExecConfig specifies a command to provide client credentials. The command is exec'd
// and outputs structured stdout holding credentials.
//
// See the client.authentiction.k8s.io API group for specifications of the exact input
// and output format
type ExecConfig struct {
	// Command to execute.
	Command string `json:"command"`
	// Arguments to pass to the command when executing it.
	// +optional
	Args []string `json:"args"`
	// Env defines additional environment variables to expose to the process. These
	// are unioned with the host's environment, as well as variables client-go uses
	// to pass argument to the plugin.
	// +optional
	Env []ExecEnvVar `json:"env"`

	// Preferred input version of the ExecInfo. The returned ExecCredentials MUST use
	// the same encoding version as the input.
	APIVersion string `json:"apiVersion,omitempty"`
}

// ExecEnvVar is used for setting environment variables when executing an exec-based
// credential plugin.
type ExecEnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// YamlInput yaml content
type YamlInput struct {
	// FileName yaml file
	FileName string
	// YamlContent yaml content
	YamlContent string
}

// GetKubeConfigFromYAMLBody get kubeConfig from YAML file
func GetKubeConfigFromYAMLBody(isFile bool, input YamlInput) (*Config, error) {
	var (
		kubeConfig []byte
		err        error
	)

	if isFile {
		kubeConfig, err = ioutil.ReadFile(input.FileName)
		if err != nil {
			return nil, err
		}
	} else {
		kubeConfig = []byte(input.YamlContent)
	}

	// Init an empty Config as unmarshal layout template
	c := Config{}
	if err := yaml.Unmarshal(kubeConfig, &c); err != nil {
		return nil, err
	}

	return &c, nil
}
