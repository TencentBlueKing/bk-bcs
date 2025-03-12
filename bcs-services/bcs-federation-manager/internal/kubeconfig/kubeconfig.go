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

// Package kubeconfig ...
package kubeconfig

import (
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	// DefaultKubeconfigUserToken default user token which is useless for now,
	// but cannot be empty for register to cluster manager
	DefaultKubeconfigUserToken = "xxxxxxx"
)

// NewConfigForRegister create a new kubeconfig for register federation cluster
func NewConfigForRegister(address string) *Config {
	return newConfig(true, address, DefaultKubeconfigUserToken, "fed-cluster", "fed-context", "admin")
}

// NewConfigForProvider create a new kubeconfig for clusternet provider
func NewConfigForProvider(address string, token string, clusterId string) *Config {
	clusterId = strings.ToLower(clusterId)
	return newConfig(false, address, token, clusterId, "master-"+clusterId, "clusternet-computing")
}

func newConfig(skipTLS bool, address, token, clusterName, contextName, userName string) *Config {
	config := &Config{
		APIVersion:     "v1",
		Clusters:       make([]Cluster, 0),
		Contexts:       make([]Context, 0),
		CurrentContext: "",
		Kind:           "Config",
		Preferences:    struct{}{},
		Users:          make([]User, 0),
	}

	config.AddCluster(
		&Cluster{
			Name: clusterName,
			Cluster: ClusterInfo{
				InsecureSkipTLSVerify: skipTLS,
				Server:                address,
			},
		},
	).AddContext(
		&Context{
			Name: contextName,
			Context: ContextInfo{
				Cluster: clusterName,
				User:    userName,
			},
		},
	).AddUser(
		&User{
			Name: userName,
			User: UserInfo{
				Token: token,
			},
		},
	).SetCurrentContext(contextName)

	return config
}

// Config kubeconfig for import cluster
type Config struct {
	APIVersion     string    `yaml:"apiVersion"`
	Clusters       []Cluster `yaml:"clusters"`
	Contexts       []Context `yaml:"contexts"`
	CurrentContext string    `yaml:"current-context"`
	Kind           string    `yaml:"kind"`
	Preferences    struct{}  `yaml:"preferences"`
	Users          []User    `yaml:"users"`
}

// Yaml return the yaml format string
func (c *Config) Yaml() string {
	result, _ := yaml.Marshal(c)
	return string(result)
}

// AddCluster add a cluster to kubeconfig
func (c *Config) AddCluster(cls *Cluster) *Config {
	if c.Clusters == nil {
		c.Clusters = make([]Cluster, 0)
	}
	c.Clusters = append(c.Clusters, *cls)
	return c
}

// AddContext add a context to kubeconfig
func (c *Config) AddContext(ctx *Context) *Config {
	if c.Contexts == nil {
		c.Contexts = make([]Context, 0)
	}
	c.Contexts = append(c.Contexts, *ctx)
	return c
}

// AddUser add a user to kubeconfig
func (c *Config) AddUser(usr *User) *Config {
	if c.Users == nil {
		c.Users = make([]User, 0)
	}
	c.Users = append(c.Users, *usr)
	return c
}

// SetCurrentContext set the current context
func (c *Config) SetCurrentContext(context string) *Config {
	c.CurrentContext = context
	return c
}

// Cluster represents a cluster configuration
type Cluster struct {
	Cluster ClusterInfo `yaml:"cluster"`
	Name    string      `yaml:"name"`
}

// ClusterInfo represents the details of a cluster
type ClusterInfo struct {
	InsecureSkipTLSVerify bool   `yaml:"insecure-skip-tls-verify,omitempty"`
	Server                string `yaml:"server"`
}

// Context represents a context configuration
type Context struct {
	Context ContextInfo `yaml:"context"`
	Name    string      `yaml:"name"`
}

// ContextInfo represents the details of a context
type ContextInfo struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

// User represents a user configuration
type User struct {
	Name string   `yaml:"name"`
	User UserInfo `yaml:"user"`
}

// UserInfo represents the details of a user
type UserInfo struct {
	Token string `yaml:"token"`
}
