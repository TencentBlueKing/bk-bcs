/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package model

// Ing Ingress 表单化建模
type Ing struct {
	Metadata Metadata `structs:"metadata"`
	Spec     IngSpec  `structs:"spec"`
}

// IngSpec xxx
type IngSpec struct {
	RuleConf       IngRuleConf       `structs:"ruleConf"`
	DefaultBackend IngDefaultBackend `structs:"defaultBackend"`
	Cert           IngCert           `structs:"cert"`
}

// IngRuleConf xxx
type IngRuleConf struct {
	Rules []IngRule `structs:"rules"`
}

// IngRule xxx
type IngRule struct {
	Domain string    `structs:"domain"`
	Paths  []IngPath `structs:"paths"`
}

// IngPath xxx
type IngPath struct {
	Type      string `structs:"type"`
	Path      string `structs:"path"`
	TargetSVC string `structs:"targetSVC"`
	Port      int64  `structs:"port"`
}

// IngDefaultBackend xxx
type IngDefaultBackend struct {
	TargetSVC string `structs:"targetSVC"`
	Port      int64  `structs:"port"`
}

// IngCert xxx
type IngCert struct {
	TLS []IngTLS `structs:"tls"`
}

// IngTLS xxx
type IngTLS struct {
	SecretName string   `structs:"secretName"`
	Hosts      []string `structs:"hosts"`
}

// SVC Service 表单化建模
type SVC struct {
	Metadata Metadata `structs:"metadata"`
	Spec     SVCSpec  `structs:"spec"`
}

// SVCSpec xxx
type SVCSpec struct {
	PortConf        SVCPortConf     `structs:"portConf"`
	Selector        SVCSelector     `structs:"selector"`
	SessionAffinity SessionAffinity `structs:"sessionAffinity"`
	IP              IPConf          `structs:"ip"`
}

// SVCPortConf xxx
type SVCPortConf struct {
	Type  string    `structs:"type"`
	Ports []SVCPort `structs:"ports"`
}

// SVCPort xxx
type SVCPort struct {
	Name       string `structs:"name"`
	Port       int64  `structs:"port"`
	Protocol   string `structs:"protocol"`
	TargetPort int64  `structs:"targetPort"`
	NodePort   int64  `structs:"nodePort"`
}

// SVCSelector xxx
type SVCSelector struct {
	Labels []LabelSelector `structs:"labels"`
}

// SessionAffinity xxx
type SessionAffinity struct {
	Type       string `structs:"type"`
	StickyTime int64  `structs:"stickyTime"`
}

// IPConf xxx
type IPConf struct {
	Address  string   `structs:"address"`
	External []string `structs:"external"`
}

// EP Endpoint 表单化建模
type EP struct {
	Metadata Metadata `structs:"metadata"`
	Spec     EPSpec   `structs:"spec"`
}

// EPSpec xxx
type EPSpec struct {
	Address []string `structs:"address"`
	Ports   []EPPort `structs:"ports"`
}

// EPPort xxx
type EPPort struct {
	Name     string `structs:"name"`
	Port     int64  `structs:"port"`
	Protocol string `structs:"protocol"`
}
