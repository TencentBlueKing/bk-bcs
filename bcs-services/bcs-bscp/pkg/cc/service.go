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

package cc

import (
	"net"
	"sync"
)

var (
	initOnce sync.Once

	// serviceName is the runtime service's name.
	serviceName Name
)

// InitService set the initial service.
func InitService(sn Name) {
	initOnce.Do(func() {
		serviceName = sn
	})
}

// ServiceName return the current runtime service's name.
func ServiceName() Name {
	return serviceName
}

// Name is the name of the service
type Name string

const (
	// APIServerName is api server's name
	APIServerName Name = "api-server"
	// DataServiceName is data service's name
	DataServiceName Name = "data-service"
	// CacheServiceName is cache service's name
	CacheServiceName Name = "cache-service"
	// ConfigServerName is the config server's service name
	ConfigServerName Name = "config-server"
	// FeedServerName is the feed server's service name
	FeedServerName Name = "feed-server"
	// AuthServerName is the auth server's service name
	AuthServerName Name = "auth-server"
	// VaultServerName is the vault server's service name
	VaultServerName Name = "vault-server"
	// UIName is the ui service name
	UIName Name = "ui"
)

// Setting defines all service Setting interface.
type Setting interface {
	trySetFlagBindIP(ip net.IP) error
	trySetFlagPort(port, grpcPort int) error
	trySetDefault()
	Validate() error
}

// ApiServerSetting defines api server used setting options.
type ApiServerSetting struct {
	Network      Network      `yaml:"network"`
	Service      Service      `yaml:"service"`
	Log          LogOption    `yaml:"log"`
	Repo         Repository   `yaml:"repository"`
	BKNotice     BKNotice     `yaml:"bkNotice"`
	Esb          Esb          `yaml:"esb"`
	FeatureFlags FeatureFlags `yaml:"featureFlags"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *ApiServerSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetFlagPort set http and grpc port
func (s *ApiServerSetting) trySetFlagPort(port, grpcPort int) error {
	return s.Network.trySetFlagPort(port, grpcPort)
}

// trySetDefault set the ApiServerSetting default value if user not configured.
func (s *ApiServerSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()
	s.Repo.trySetDefault()
}

// Validate ApiServerSetting option.
func (s ApiServerSetting) Validate() error {

	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	if err := s.Repo.validate(); err != nil {
		return err
	}

	return nil
}

// AuthServerSetting defines auth server used setting options.
type AuthServerSetting struct {
	Network   Network           `yaml:"network"`
	Service   Service           `yaml:"service"`
	Log       LogOption         `yaml:"log"`
	LoginAuth LoginAuthSettings `yaml:"loginAuth"`
	IAM       IAM               `yaml:"iam"`
	Esb       Esb               `yaml:"esb"`
}

// LoginAuthSettings login conf
type LoginAuthSettings struct {
	Host      string `yaml:"host"`
	InnerHost string `yaml:"innerHost"`
	Provider  string `yaml:"provider"`
	UseESB    bool   `yaml:"useEsb"`
	GWPubKey  string `yaml:"gwPubkey"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *AuthServerSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetFlagPort set http and grpc port
func (s *AuthServerSetting) trySetFlagPort(port, grpcPort int) error {
	return s.Network.trySetFlagPort(port, grpcPort)
}

// trySetDefault set the AuthServerSetting default value if user not configured.
func (s *AuthServerSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()
}

// Validate AuthServerSetting option.
func (s AuthServerSetting) Validate() error {

	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	if err := s.IAM.validate(); err != nil {
		return err
	}

	return nil
}

// CacheServiceSetting defines cache service used setting options.
type CacheServiceSetting struct {
	Network Network   `yaml:"network"`
	Service Service   `yaml:"service"`
	Log     LogOption `yaml:"log"`

	Credential   Credential   `yaml:"credential"`
	Sharding     Sharding     `yaml:"sharding"`
	RedisCluster RedisCluster `yaml:"redisCluster"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *CacheServiceSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetFlagPort set http and grpc port
func (s *CacheServiceSetting) trySetFlagPort(port, grpcPort int) error {
	return s.Network.trySetFlagPort(port, grpcPort)
}

// trySetDefault set the CacheServiceSetting default value if user not configured.
func (s *CacheServiceSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()
	s.Sharding.trySetDefault()
	s.RedisCluster.trySetDefault()
}

// Validate CacheServiceSetting option.
func (s CacheServiceSetting) Validate() error {

	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	if err := s.Sharding.validate(); err != nil {
		return err
	}

	if err := s.RedisCluster.validate(); err != nil {
		return err
	}

	return nil
}

// ConfigServerSetting defines config server used setting options.
type ConfigServerSetting struct {
	Network    Network    `yaml:"network"`
	Service    Service    `yaml:"service"`
	Credential Credential `yaml:"credential"`
	Log        LogOption  `yaml:"log"`
	Repo       Repository `yaml:"repository"`
	Esb        Esb        `yaml:"esb"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *ConfigServerSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetFlagPort set http and grpc port
func (s *ConfigServerSetting) trySetFlagPort(port, grpcPort int) error {
	return s.Network.trySetFlagPort(port, grpcPort)
}

// trySetDefault set the ConfigServerSetting default value if user not configured.
func (s *ConfigServerSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()
}

// Validate ConfigServerSetting option.
func (s ConfigServerSetting) Validate() error {

	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	if err := s.Repo.validate(); err != nil {
		return err
	}

	if err := s.Credential.validate(); err != nil {
		return err
	}

	return nil
}

// DataServiceSetting defines cache service used setting options.
type DataServiceSetting struct {
	Network Network   `yaml:"network"`
	Service Service   `yaml:"service"`
	Log     LogOption `yaml:"log"`

	Credential Credential `yaml:"credential"`
	Sharding   Sharding   `yaml:"sharding"`
	Esb        Esb        `yaml:"esb"`
	Repo       Repository `yaml:"repository"`
	Vault      Vault      `yaml:"vault"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *DataServiceSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetFlagPort set http and grpc port
func (s *DataServiceSetting) trySetFlagPort(port, grpcPort int) error {
	return s.Network.trySetFlagPort(port, grpcPort)
}

// trySetDefault set the DataServiceSetting default value if user not configured.
func (s *DataServiceSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()
	s.Sharding.trySetDefault()
	s.Repo.trySetDefault()
	s.Vault.getConfigFromEnv()
}

// Validate DataServiceSetting option.
func (s DataServiceSetting) Validate() error {

	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	if err := s.Sharding.validate(); err != nil {
		return err
	}

	if err := s.Esb.validate(); err != nil {
		return err
	}

	if err := s.Repo.validate(); err != nil {
		return err
	}

	if err := s.Vault.validate(); err != nil {
		return err
	}

	return nil
}

// FeedServerSetting defines feed server used setting options.
type FeedServerSetting struct {
	Network Network   `yaml:"network"`
	Service Service   `yaml:"service"`
	Log     LogOption `yaml:"log"`

	Repository   Repository          `yaml:"repository"`
	Esb          Esb                 `yaml:"esb"`
	BCS          BCS                 `yaml:"bcs"`
	GSE          GSE                 `yaml:"gse"`
	FSLocalCache FSLocalCache        `yaml:"fsLocalCache"`
	Downstream   Downstream          `yaml:"downstream"`
	MRLimiter    MatchReleaseLimiter `yaml:"matchReleaseLimiter"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *FeedServerSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetFlagPort set http and grpc port
func (s *FeedServerSetting) trySetFlagPort(port, grpcPort int) error {
	return s.Network.trySetFlagPort(port, grpcPort)
}

// trySetDefault set the FeedServerSetting default value if user not configured.
func (s *FeedServerSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()
	s.FSLocalCache.trySetDefault()
	s.Downstream.trySetDefault()
	s.GSE.getFromEnv()
	s.MRLimiter.trySetDefault()
}

// Validate FeedServerSetting option.
func (s FeedServerSetting) Validate() error {

	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	if err := s.Repository.validate(); err != nil {
		return err
	}

	if err := s.FSLocalCache.validate(); err != nil {
		return err
	}

	if err := s.Downstream.validate(); err != nil {
		return err
	}

	if err := s.MRLimiter.validate(); err != nil {
		return err
	}

	if err := s.Esb.validate(); err != nil {
		return err
	}

	if err := s.GSE.validate(); err != nil {
		return err
	}

	return nil
}

// VaultServerSetting defines cache service used setting options.
type VaultServerSetting struct {
	Network    Network    `yaml:"network"`
	Service    Service    `yaml:"service"`
	Log        LogOption  `yaml:"log"`
	Credential Credential `yaml:"credential"`
	Sharding   Sharding   `yaml:"sharding"`
}

// trySetFlagBindIP try set flag bind ip.
func (s *VaultServerSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetFlagPort set http and grpc port
func (s *VaultServerSetting) trySetFlagPort(port, grpcPort int) error {
	return s.Network.trySetFlagPort(port, grpcPort)
}

// trySetDefault set the CacheServiceSetting default value if user not configured.
func (s *VaultServerSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Service.trySetDefault()
	s.Log.trySetDefault()
	s.Sharding.trySetDefault()
}

// Validate CacheServiceSetting option.
func (s VaultServerSetting) Validate() error {

	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Service.validate(); err != nil {
		return err
	}

	if err := s.Sharding.validate(); err != nil {
		return err
	}

	return nil
}
