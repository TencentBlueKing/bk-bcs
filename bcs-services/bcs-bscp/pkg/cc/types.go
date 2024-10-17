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
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	etcd3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm/logger"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/version"
)

const (
	// RedisStandaloneMode 单节点 redis
	RedisStandaloneMode = "standalone"
	// RedisSentinelMode 哨兵模式 redis，哨兵实例
	RedisSentinelMode = "sentinel"
	// RedisClusterMode 集群模式 redis
	RedisClusterMode = "cluster"
)

// FeatureFlags 特性配置
type FeatureFlags struct {
	// BizView 业务白名单
	BizView FeatureBizView `json:"biz_view" yaml:"BIZ_VIEW"`
	// ResourceLimit 业务资源限制
	ResourceLimit FeatureResourceLimit `json:"resource_limit" yaml:"RESOURCE_LIMIT"`
}

// FeatureBizView 业务白名单
type FeatureBizView struct {
	Default *bool `yaml:"default"`
	// map[bizID]true/false
	Spec map[string]*bool `yaml:"spec"`
}

// FeatureResourceLimit 业务资源限制
type FeatureResourceLimit struct {
	Default ResourceLimit `json:"default" yaml:"default"`
	// map[bizID]ResourceLimit
	Spec map[string]ResourceLimit `json:"spec" yaml:"spec"`
}

// ResourceLimit 资源限制配置项
type ResourceLimit struct {
	// MaxFileSize 配置文件大小上限，单位 MB，默认为200MB
	MaxFileSize uint `json:"maxFileSize" yaml:"maxFileSize"`
	// AppConfigCnt 单个app下允许创建的配置数（模版+非模版），默认为2000
	AppConfigCnt uint `yaml:"appConfigCnt"`
	// TmplSetTmplCnt 单个模版套餐下允许创建的模版数，默认为2000
	TmplSetTmplCnt uint `yaml:"tmplSetTmplCnt"`
	// MaxUploadContentLength 最大内容长度
	MaxUploadContentLength uint `yaml:"maxUploadContentLength"`
}

// validate if the feature resource limit is valid or not.
func (f FeatureFlags) validate() error {
	for bizID := range f.BizView.Spec {
		if _, err := strconv.Atoi(bizID); err != nil {
			return fmt.Errorf("invalid featureFlags.BIZ_VIEW.spec.{bizID} value %s, "+
				"biz id should be an interger", bizID)
		}
	}

	for bizID := range f.ResourceLimit.Spec {
		if _, err := strconv.Atoi(bizID); err != nil {
			return fmt.Errorf("invalid featureFlags.RESOURCE_LIMIT.spec.{bizID} value %s, "+
				"biz id should be an interger", bizID)
		}
	}

	return nil
}

const (
	// DefaultBizView is default biz view
	DefaultBizView = true
	// DefaultMaxFileSize is default max file size, unit is MB
	DefaultMaxFileSize = 200
	// DefaultAppConfigCnt is default app's config count
	DefaultAppConfigCnt = 2000
	// DefaultTmplSetTmplCnt is default template set's template count
	DefaultTmplSetTmplCnt = 2000
	// DefaultMaxUploadContentLength 默认最大内容长度(2048MB)
	DefaultMaxUploadContentLength = 2 * 1024
	// DefaultMaxUploadSingleContentLength 默认最大单个内容长度(200MB)
	DefaultMaxUploadSingleContentLength = 200
)

// trySetDefault try set the default value of feature flag
func (f *FeatureFlags) trySetDefault() {
	if f.BizView.Default == nil {
		bizView := DefaultBizView
		f.BizView.Default = &bizView
	}

	if f.ResourceLimit.Default.MaxFileSize == 0 {
		f.ResourceLimit.Default.MaxFileSize = DefaultMaxFileSize
	}

	if f.ResourceLimit.Default.AppConfigCnt == 0 {
		f.ResourceLimit.Default.AppConfigCnt = DefaultAppConfigCnt
	}

	if f.ResourceLimit.Default.TmplSetTmplCnt == 0 {
		f.ResourceLimit.Default.TmplSetTmplCnt = DefaultTmplSetTmplCnt
	}

	if f.ResourceLimit.Default.MaxUploadContentLength == 0 {
		f.ResourceLimit.Default.MaxUploadContentLength = DefaultMaxUploadContentLength
	}
}

// Service defines Setting related runtime.
type Service struct {
	Etcd Etcd `yaml:"etcd"`
}

// trySetDefault set the Setting default value if user not configured.
func (s *Service) trySetDefault() {
	s.Etcd.trySetDefault()
}

// validate Setting related runtime.
func (s Service) validate() error {
	if err := s.Etcd.validate(); err != nil {
		return err
	}

	return nil
}

// Etcd defines etcd related runtime
type Etcd struct {
	// Endpoints is a list of URLs.
	Endpoints []string `yaml:"endpoints"`
	// DialTimeoutMS is the timeout seconds for failing
	// to establish a connection.
	DialTimeoutMS uint `yaml:"dialTimeoutMS"`
	// Username is a user's name for authentication.
	Username string `yaml:"username"`
	// Password is a password for authentication.
	Password string    `yaml:"password"`
	TLS      TLSConfig `yaml:"tls"`
}

// trySetDefault set the etcd default value if user not configured.
func (es *Etcd) trySetDefault() {
	if len(es.Endpoints) == 0 {
		es.Endpoints = []string{"127.0.0.1:2379"}
	}

	if es.DialTimeoutMS == 0 {
		es.DialTimeoutMS = 200
	}
}

// ToConfig convert to etcd config.
func (es Etcd) ToConfig() (etcd3.Config, error) {
	var tlsC *tls.Config
	if es.TLS.Enable() {
		var err error
		tlsC, err = tools.ClientTLSConfVerify(es.TLS.InsecureSkipVerify, es.TLS.CAFile, es.TLS.CertFile,
			es.TLS.KeyFile, es.TLS.Password)
		if err != nil {
			return etcd3.Config{}, fmt.Errorf("init etcd tls config failed, err: %v", err)
		}
	}

	c := etcd3.Config{
		Endpoints:            es.Endpoints,
		AutoSyncInterval:     0,
		DialTimeout:          time.Duration(es.DialTimeoutMS) * time.Millisecond,
		DialKeepAliveTime:    0,
		DialKeepAliveTimeout: 0,
		MaxCallSendMsgSize:   0,
		MaxCallRecvMsgSize:   0,
		TLS:                  tlsC,
		Username:             es.Username,
		Password:             es.Password,
		RejectOldCluster:     false,
		DialOptions:          nil,
		Context:              nil,
		LogConfig:            nil,
		PermitWithoutStream:  false,
	}

	return c, nil
}

// validate etcd runtime
func (es Etcd) validate() error {
	if len(es.Endpoints) == 0 {
		return errors.New("etcd endpoints is not set")
	}

	if err := es.TLS.validate(); err != nil {
		return fmt.Errorf("etcd tls, %v", err)
	}

	return nil
}

// RedisCluster defines all the redis related runtime
type RedisCluster struct {
	// Endpoints is a seed list of host:port addresses of nodes.
	Endpoints []string `yaml:"endpoints"`
	// Username is a user's name for authentication.
	Username string `yaml:"username"`
	// Password is a password for authentication.
	Password       string `yaml:"password"`
	DialTimeoutMS  uint   `yaml:"dialTimeoutMS"`
	ReadTimeoutMS  uint   `yaml:"readTimeoutMS"`
	WriteTimeoutMS uint   `yaml:"writeTimeoutMS"`
	MinIdleConn    uint   `yaml:"minIdleConn"`
	DB             int    `yaml:"db"`
	Mode           string `yaml:"mode"` // 支持集群或者单例模式 可选项 standalone,cluster
	// PoolSize defines the connection pool size for
	// each node of the redis cluster.
	PoolSize uint      `yaml:"poolSize"`
	TLS      TLSConfig `yaml:"tls"`
	// MaxSlowLogLatencyMS defines the max tolerance in millisecond to execute
	// the redis command, if the cost time of execute have >= the MaxSlowLogLatencyMS
	// then this request will be logged.
	MaxSlowLogLatencyMS uint `yaml:"maxSlowLogLatencyMS"`
}

// trySetDefault set the redis's default value if user not configured.
func (rs *RedisCluster) trySetDefault() {
	if len(rs.Endpoints) == 0 {
		rs.Endpoints = []string{"127.0.0.1:6379"}
	}

	if rs.DialTimeoutMS == 0 {
		rs.DialTimeoutMS = 100
	}

	if rs.ReadTimeoutMS == 0 {
		rs.ReadTimeoutMS = 500
	}

	if rs.WriteTimeoutMS == 0 {
		rs.WriteTimeoutMS = 500
	}

	if rs.MinIdleConn == 0 {
		rs.MinIdleConn = 10
	}

	if rs.PoolSize == 0 {
		rs.PoolSize = 300
	}

	if rs.MaxSlowLogLatencyMS == 0 {
		rs.MaxSlowLogLatencyMS = 100
	}

}

// validate redis runtime
func (rs RedisCluster) validate() error {
	if len(rs.Endpoints) == 0 {
		return errors.New("redis endpoints is not set")
	}

	if (rs.DialTimeoutMS > 0 && rs.DialTimeoutMS < 50) || rs.DialTimeoutMS > 1000 {
		return errors.New("invalid redis dialTimeoutMS, should be in [50:1000]ms")
	}

	if (rs.ReadTimeoutMS > 0 && rs.ReadTimeoutMS < 10) || rs.ReadTimeoutMS > 500 {
		return errors.New("invalid redis readTimeoutMS, should be in [10:500]ms")
	}

	if (rs.WriteTimeoutMS > 0 && rs.WriteTimeoutMS < 10) || rs.WriteTimeoutMS > 500 {
		return errors.New("invalid redis writeTimeoutMS, should be in [10:500]ms")
	}

	if err := rs.TLS.validate(); err != nil {
		return fmt.Errorf("redis tls, %v", err)
	}
	return nil
}

// IAM defines all the iam related runtime.
type IAM struct {
	// IAM api url.
	APIURL string `yaml:"api_url"`
	// AppCode blueking belong to bscp's appcode.
	AppCode string `yaml:"appCode"`
	// AppSecret blueking belong to bscp app's secret.
	AppSecret string    `yaml:"appSecret"`
	TLS       TLSConfig `yaml:"tls"`
}

// validate iam runtime.
func (s IAM) validate() error {
	if len(s.APIURL) == 0 {
		return errors.New("iam api url is not set")
	}

	if len(s.AppCode) == 0 {
		return errors.New("iam appcode is not set")
	}

	if len(s.AppSecret) == 0 {
		return errors.New("iam app secret is not set")
	}

	if err := s.TLS.validate(); err != nil {
		return fmt.Errorf("iam tls, %v", err)
	}

	return nil
}

// StorageMode :
type StorageMode string

const (
	// BkRepo Type
	BkRepo StorageMode = "BKREPO"
	// S3 type
	S3 StorageMode = "S3"
)

// Repository defines all the repo related runtime.
type Repository struct {
	BaseRepo          `yaml:",inline"`
	RedisCluster      RedisCluster `yaml:"redisCluster"`
	EnableHA          bool         `yaml:"enableHA"`
	SyncPeriodSeconds uint         `yaml:"syncPeriodSeconds"`
	Slave             BaseRepo     `yaml:"slave"`
}

// BaseRepo 文件存储的基础部分
type BaseRepo struct {
	StorageType StorageMode   `yaml:"storageType"`
	S3          S3Storage     `yaml:"s3"`
	BkRepo      BkRepoStorage `yaml:"bkRepo"`
}

// BkRepoStorage BKRepo 存储类型
type BkRepoStorage struct {
	// Endpoints is a seed list of host:port addresses of repo nodes.
	Endpoints []string `yaml:"endpoints"`
	// Project bscp project name in repo.
	Project string `yaml:"project"`
	// User basic auth username.
	Username string `yaml:"username"`
	// Password basic auth password.
	Password string `yaml:"password"`
	// TLS defines the tls config for repo.
	TLS TLSConfig `yaml:"tls"`
}

// S3Storage s3 存储类型
type S3Storage struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyID"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	UseSSL          bool   `yaml:"useSSL"`
	BucketName      string `yaml:"bucketName"`
}

// repoPollingAddrIndex repo request polling address index.
var repoPollingAddrIndex = 0

// OneEndpoint rotation training strategy, returns the request address of the repo.
func (s Repository) OneEndpoint() (string, error) {
	num := len(s.BkRepo.Endpoints)
	if num == 0 {
		return "", errors.New("no repo endpoints can be used")
	}

	var addr string
	if repoPollingAddrIndex < num-1 {
		repoPollingAddrIndex++
		addr = s.BkRepo.Endpoints[repoPollingAddrIndex]
	} else {
		repoPollingAddrIndex = 0
		addr = s.BkRepo.Endpoints[repoPollingAddrIndex]
	}

	return addr, nil
}

const defaultSyncPeriodSeconds = 24 * 3600 // default is 1 day
const minSyncPeriodSeconds = 3600          // min is 1 hour

func (s *Repository) trySetDefault() {
	if len(s.StorageType) == 0 {
		s.StorageType = BkRepo
	}
	s.RedisCluster.trySetDefault()
	if s.EnableHA {
		if len(s.Slave.StorageType) == 0 {
			s.Slave.StorageType = S3
		}
	}
	if s.SyncPeriodSeconds == 0 || s.SyncPeriodSeconds < minSyncPeriodSeconds {
		s.SyncPeriodSeconds = defaultSyncPeriodSeconds
	}
}

// validate repo runtime.
func (s Repository) validate() error {
	if err := s.BaseRepo.validate(); err != nil {
		return fmt.Errorf("repository master config err: %v", err)
	}

	if err := s.RedisCluster.validate(); err != nil {
		return fmt.Errorf("repository redis cluster config err: %v", err)
	}

	if s.EnableHA {
		if err := s.Slave.validate(); err != nil {
			return fmt.Errorf("repository slave config err: %v", err)
		}
	}

	return nil
}

// validate repo base part.
func (b BaseRepo) validate() error {
	switch strings.ToUpper(string(b.StorageType)) {
	case string(S3):
		if len(b.S3.Endpoint) == 0 {
			return errors.New("s3 endpoint is not set")
		}

		if len(b.S3.AccessKeyID) == 0 {
			return errors.New("s3 accessKeyID is not set")
		}

		if len(b.S3.SecretAccessKey) == 0 {
			return errors.New("s3 secretAccessKey is not set")
		}
		if len(b.S3.BucketName) == 0 {
			return errors.New("s3 bucketName is not set")
		}
	case string(BkRepo):
		if len(b.BkRepo.Endpoints) == 0 {
			return errors.New("bk_repo endpoints is not set")
		}

		if len(b.BkRepo.Username) == 0 {
			return errors.New("repo basic auth username is not set")
		}

		if len(b.BkRepo.Password) == 0 {
			return errors.New("repo basic auth password is not set")
		}

		if len(b.BkRepo.Project) == 0 {
			return errors.New("repo project is not set")
		}

		if err := b.BkRepo.TLS.validate(); err != nil {
			return fmt.Errorf("repo tls, %v", err)
		}
	default:
		return fmt.Errorf("unsupported storage type: %s", string(b.StorageType))
	}

	return nil
}

// Limiter defines the request limit options
type Limiter struct {
	// QPS should >=1
	QPS uint `yaml:"qps"`
	// Burst should >= 1;
	Burst uint `yaml:"burst"`
}

// validate if the limiter is valid or not.
func (lm Limiter) validate() error {
	if lm.QPS <= 0 {
		return errors.New("invalid QPS value, should >= 1")
	}

	if lm.Burst <= 0 {
		return errors.New("invalid Burst value, should >= 1")
	}

	return nil
}

// trySetDefault try set the default value of limiter
func (lm *Limiter) trySetDefault() {
	if lm.QPS == 0 {
		lm.QPS = 500
	}

	if lm.Burst == 0 {
		lm.Burst = 500
	}
}

// Sharding defines sharding related runtime
type Sharding struct {
	AdminDatabase Database `yaml:"adminDatabase"`
	// MaxSlowLogLatencyMS defines the max tolerance in millisecond to execute
	// the database command, if the cost time of execute have >= the MaxSlowLogLatencyMS
	// then this request will be logged.
	MaxSlowLogLatencyMS uint `yaml:"maxSlowLogLatencyMS"`
	// Limiter defines request's to ORM's limitation for each sharding, and
	// each sharding have the independent request limitation.
	Limiter *Limiter `yaml:"limiter"`
}

// trySetDefault set the sharding default value if user not configured.
func (s *Sharding) trySetDefault() {
	s.AdminDatabase.trySetDefault()

	if s.MaxSlowLogLatencyMS == 0 {
		s.MaxSlowLogLatencyMS = 100
	}

	if s.Limiter == nil {
		s.Limiter = new(Limiter)
	}

	s.Limiter.trySetDefault()
}

// validate sharding runtime
func (s Sharding) validate() error {

	if err := s.AdminDatabase.validate(); err != nil {
		return err
	}

	if s.MaxSlowLogLatencyMS <= 0 {
		return errors.New("invalid maxSlowLogLatencyMS")
	}

	if s.Limiter != nil {
		if err := s.Limiter.validate(); err != nil {
			return fmt.Errorf("sharding.limiter is invalid, %v", err)
		}
	}

	return nil
}

// FeatureFlagOption feature flag option
type FeatureFlagOption struct {
	// 1. if set enabled to true, it means all biz enable this feature,
	// only biz in list(blacklist) disable this feature
	// 2. if set enabled to false, it means only biz in list(whitelist) enable this feature
	Enabled bool `yaml:"enabled"`
	// List can be white list or black list, depends on enabled
	List []string `yaml:"list"`
}

// Database defines database related runtime.
type Database struct {
	// Endpoints is a seed list of host:port addresses of database nodes.
	Endpoints []string `yaml:"endpoints"`
	Database  string   `yaml:"database"`
	User      string   `yaml:"user"`
	Password  string   `yaml:"password"`
	// DialTimeoutSec is timeout in seconds to wait for a
	// response from the db server
	// all the timeout default value reference:
	// https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html
	DialTimeoutSec    uint      `yaml:"dialTimeoutSec"`
	ReadTimeoutSec    uint      `yaml:"readTimeoutSec"`
	WriteTimeoutSec   uint      `yaml:"writeTimeoutSec"`
	MaxIdleTimeoutMin uint      `yaml:"maxIdleTimeoutMin"`
	MaxOpenConn       uint      `yaml:"maxOpenConn"`
	MaxIdleConn       uint      `yaml:"maxIdleConn"`
	TLS               TLSConfig `yaml:"tls"`
}

// trySetDefault set the database's default value if user not configured.
func (ds *Database) trySetDefault() {
	if len(ds.Endpoints) == 0 {
		ds.Endpoints = []string{"127.0.0.1:3306"}
	}

	if ds.DialTimeoutSec == 0 {
		ds.DialTimeoutSec = 15
	}

	if ds.ReadTimeoutSec == 0 {
		ds.ReadTimeoutSec = 5
	}

	if ds.WriteTimeoutSec == 0 {
		ds.WriteTimeoutSec = 10
	}

	if ds.MaxOpenConn == 0 {
		ds.MaxOpenConn = 500
	}

	if ds.MaxIdleConn == 0 {
		ds.MaxIdleConn = 5
	}

	if ds.MaxIdleTimeoutMin == 0 {
		ds.MaxIdleTimeoutMin = 3
	}
}

// validate database runtime.
func (ds Database) validate() error {
	if len(ds.Endpoints) == 0 {
		return errors.New("database endpoints is not set")
	}

	if len(ds.Database) == 0 {
		return errors.New("database is not set")
	}

	if (ds.DialTimeoutSec > 0 && ds.DialTimeoutSec < 1) || ds.DialTimeoutSec > 60 {
		return errors.New("invalid database dialTimeoutMS, should be in [1:60]s")
	}

	if (ds.ReadTimeoutSec > 0 && ds.ReadTimeoutSec < 1) || ds.ReadTimeoutSec > 60 {
		return errors.New("invalid database readTimeoutMS, should be in [1:60]s")
	}

	if (ds.WriteTimeoutSec > 0 && ds.WriteTimeoutSec < 1) || ds.WriteTimeoutSec > 30 {
		return errors.New("invalid database writeTimeoutMS, should be in [1:30]s")
	}

	if err := ds.TLS.validate(); err != nil {
		return fmt.Errorf("database tls, %v", err)
	}

	return nil
}

// LogOption defines log's related configuration
type LogOption struct {
	LogDir           string `yaml:"logDir"`
	MaxPerFileSizeMB uint32 `yaml:"maxPerFileSizeMB"`
	MaxPerLineSizeKB uint32 `yaml:"maxPerLineSizeKB"`
	MaxFileNum       uint   `yaml:"maxFileNum"`
	LogAppend        bool   `yaml:"logAppend"`
	// log the log to std err only, it can not be used with AlsoToStdErr
	// at the same time.
	ToStdErr bool `yaml:"toStdErr"`
	// log the log to file and also to std err. it can not be used with ToStdErr
	// at the same time.
	AlsoToStdErr bool `yaml:"alsoToStdErr"`
	Verbosity    uint `yaml:"verbosity"`
}

// trySetDefault set the log's default value if user not configured.
func (log *LogOption) trySetDefault() {
	if len(log.LogDir) == 0 {
		log.LogDir = "./"
	}

	if log.MaxPerFileSizeMB == 0 {
		log.MaxPerFileSizeMB = 500
	}

	if log.MaxPerLineSizeKB == 0 {
		log.MaxPerLineSizeKB = 5
	}

	if log.MaxFileNum == 0 {
		log.MaxFileNum = 5
	}

}

// Logs convert it to logs.LogConfig.
func (log LogOption) Logs() logs.LogConfig {
	l := logs.LogConfig{
		LogDir:             log.LogDir,
		LogMaxSize:         log.MaxPerFileSizeMB,
		LogLineMaxSize:     log.MaxPerLineSizeKB,
		LogMaxNum:          log.MaxFileNum,
		RestartNoScrolling: log.LogAppend,
		ToStdErr:           log.ToStdErr,
		AlsoToStdErr:       log.AlsoToStdErr,
		Verbosity:          log.Verbosity,
	}

	return l
}

// Network defines all the network related options
type Network struct {
	// BindIP is the advertised IP address for service discovery.
	// if in ipv4 single stack mode, the BindIP would be ipv4 address.
	// if in ipv6 single stack mode, the BindIP would be ipv6 address.
	BindIP string `yaml:"bindIP"`
	// BindIPs is the advertised IP address for service expose.
	BindIPs []string `yaml:"bindIPs"`
	// RpcPort is port where server listen to rpc port.
	RpcPort uint `yaml:"rpcPort"`
	// HttpPort is port where server listen to http port.
	HttpPort uint `yaml:"httpPort"`
	// GwHttpPort  is port where server listen to grpc-gateway http port.
	GwHttpPort uint      `yaml:"gwHttpPort"`
	TLS        TLSConfig `yaml:"tls"`
}

// trySetFlagBindIP try set flag bind ip, bindIP only can set by one of the flag or configuration file.
func (n *Network) trySetFlagBindIP(ip net.IP) error {
	if len(ip) != 0 {
		if len(n.BindIP) != 0 {
			return errors.New("bind ip only can set by one of the flags or configuration file")
		}

		n.BindIP = ip.String()
		return nil
	}

	return nil
}

// trySetFlagPort set http and grpc port
func (n *Network) trySetFlagPort(port, grpcPort int) error {
	if n.HttpPort == 0 && port > 0 {
		n.HttpPort = uint(port)
	}
	if n.RpcPort == 0 && grpcPort > 0 {
		n.RpcPort = uint(grpcPort)
	}

	return nil
}

// trySetDefault set the network's default value if user not configured.
func (n *Network) trySetDefault() {
	ip, ips := tools.GetIPsFromEnv()
	if len(n.BindIP) == 0 {
		if ip != "" {
			n.BindIP = ip
		} else {
			n.BindIP = "127.0.0.1"
		}
	}
	if len(n.BindIPs) == 0 {
		if len(ips) > 0 {
			n.BindIPs = ips
		} else {
			n.BindIPs = []string{n.BindIP}
		}
	}
	if n.GwHttpPort == 0 {
		n.GwHttpPort = 80
	}
}

// validate network options
func (n Network) validate() error {

	if len(n.BindIP) == 0 {
		return errors.New("network bindIP is not set")
	}

	if net.ParseIP(n.BindIP) == nil {
		return fmt.Errorf("invalid network bindIP: %s", n.BindIP)
	}

	for _, ip := range n.BindIPs {
		if net.ParseIP(ip) == nil {
			return fmt.Errorf("invalid network bindIPs: %s", ip)
		}
	}

	if err := n.TLS.validate(); err != nil {
		return fmt.Errorf("network tls, %v", err)
	}

	return nil
}

// TLSConfig defines tls related options.
type TLSConfig struct {
	// Server should be accessed without verifying the TLS certificate.
	// For testing only.
	InsecureSkipVerify bool `yaml:"insecureSkipVerify"`
	// Server requires TLS client certificate authentication
	CertFile string `yaml:"certFile"`
	// Server requires TLS client certificate authentication
	KeyFile string `yaml:"keyFile"`
	// Trusted root certificates for server
	CAFile string `yaml:"caFile"`
	// the password to decrypt the certificate
	Password string `yaml:"password"`
}

// Enable test tls if enable.
func (tls TLSConfig) Enable() bool {
	if len(tls.CertFile) == 0 &&
		len(tls.KeyFile) == 0 &&
		len(tls.CAFile) == 0 {
		return false
	}

	return true
}

// validate tls configs
func (tls TLSConfig) validate() error {
	if !tls.Enable() {
		return nil
	}

	if len(tls.CertFile) == 0 {
		return errors.New("enabled tls, but cert file is not configured")
	}

	if len(tls.KeyFile) == 0 {
		return errors.New("enabled tls, but key file is not configured")
	}

	if len(tls.CAFile) == 0 {
		return errors.New("enabled tls, but ca file is not configured")
	}

	return nil
}

// SysOption is the system's normal option, which is parsed from
// flag commandline.
type SysOption struct {
	ConfigFiles []string
	// BindIP Setting startup bind ip.
	BindIP   net.IP
	Port     int
	GRPCPort int
	// Versioned Setting if show current version info.
	Versioned bool
}

// CheckV check if show current version info.
func (s SysOption) CheckV() {
	if s.Versioned {
		version.ShowVersion("", version.Row)
		os.Exit(0)
	}
}

// Esb defines the esb related runtime.
type Esb struct {
	// Endpoints is a seed list of host:port addresses of esb nodes.
	Endpoints []string `yaml:"endpoints"`
	// AppCode is the blueking app code of bscp to request esb.
	AppCode string `yaml:"appCode"`
	// AppSecret is the blueking app secret of bscp to request esb.
	AppSecret string `yaml:"appSecret"`
	// User is the blueking user of bscp to request esb.
	User     string    `yaml:"user"`
	TLS      TLSConfig `yaml:"tls"`
	BscpHost string    `yaml:"bscpHost"`
}

// validate esb runtime.
func (s Esb) validate() error {
	if len(s.Endpoints) == 0 {
		return errors.New("esb endpoints is not set")
	}

	if len(s.AppCode) == 0 {
		return errors.New("esb app code is not set")
	}

	if len(s.AppSecret) == 0 {
		return errors.New("esb app secret is not set")
	}

	if len(s.User) == 0 {
		return errors.New("esb user is not set")
	}

	if err := s.TLS.validate(); err != nil {
		return fmt.Errorf("validate esb tls failed, err: %v", err)
	}

	return nil
}

// FSLocalCache defines feed server's local cache related runtime.
type FSLocalCache struct {
	// AppCacheSize defines how many app can be cached.
	AppCacheSize uint `yaml:"appCacheSize"`
	// AppCacheTTLSec defines how long will this app can be cached in seconds.
	AppCacheTTLSec uint `yaml:"appCacheTTLSec"`

	// ReleasedInstanceCacheSize defines how many released instance can be cached.
	ReleasedInstanceCacheSize uint `yaml:"releasedInstanceCacheSize"`
	// ReleasedInstanceCacheTTLSec defines how long will this released instance can be cached in seconds.
	// the large of the value, the longer it will take for the published instance take effected. should <= 120.
	ReleasedInstanceCacheTTLSec uint `yaml:"releasedInstanceCacheTTLSec"`

	// ReleasedCICacheSize defines how many released configuration items can be cached.
	ReleasedCICacheSize uint `yaml:"releasedCICacheSize"`
	// ReleasedCICacheTTLSec defines how long will this released configuration items can be cached in seconds.
	ReleasedCICacheTTLSec uint `yaml:"releasedCICacheTTLSec"`

	// ReleasedKvCacheSize defines how many released kvs can be cached.
	ReleasedKvCacheSize uint `yaml:"releasedKvCacheSize"`
	// ReleasedKvCacheTTLSec defines how long will this released kvs can be cached in seconds.
	ReleasedKvCacheTTLSec uint `yaml:"releasedKvCacheTTLSec"`

	// PublishedStrategyCacheSize defines how many published strategies can be cached.
	PublishedStrategyCacheSize uint `yaml:"publishedStrategyCacheSize"`
	// PublishedStrategyCacheTTLSec defines how long will this published strategy can be cached in seconds.
	// the large of value, the longer it takes for the published app strategy take effected. should <= 120.
	PublishedStrategyCacheTTLSec uint `yaml:"publishedStrategyCacheTTLSec"`

	// ReleasedGroupCacheSize defines how many released groups can be cached.
	ReleasedGroupCacheSize uint `yaml:"releasedGroupCacheSize"`
	// ReleasedGroupCacheTTLSec defines how long will this released group can be cached in seconds.
	ReleasedGroupCacheTTLSec uint `yaml:"releasedGroupCacheTTLSec"`

	// AuthCacheSize defines how many auth results can be cached.
	AuthCacheSize uint `yaml:"authCacheSize"`
	// AuthCacheTTLSec defines how long this auth result with permission can be cached in seconds.
	AuthCacheTTLSec uint `yaml:"authCacheTTLSec"`

	// CredentialCacheSize defines how many credentials can be cached.
	CredentialCacheSize uint `yaml:"credentialCacheSize"`
	// CredentialCacheTTLSec defines how long this credential can be cached in seconds.
	CredentialCacheTTLSec uint `yaml:"credentialCacheTTLSec"`

	// ReleasedHookCacheSize defines how many released hooks can be cached.
	ReleasedHookCacheSize uint `yaml:"releasedHookCacheSize"`
	// ReleasedHookCacheTTLSec defines how long will this released hooks can be cached in seconds.
	ReleasedHookCacheTTLSec uint `yaml:"releasedHookCacheTTLSec"`

	// AsyncDownloadCacheSize defines how many async download task can be cached.
	AsyncDownloadCacheSize uint `yaml:"asyncDownloadCacheSize"`
	// AsyncDownloadCacheTTLSec defines how long will this async download task can be cached in seconds.
	AsyncDownloadCacheTTLSec uint `yaml:"asyncDownloadCacheTTLSec"`
}

// validate if the feed server's local cache runtime is valid or not.
func (fc FSLocalCache) validate() error {

	if fc.ReleasedInstanceCacheTTLSec > 600 {
		return errors.New("invalid fsLocalCache.releasedInstanceCacheTTLSec value, should <= 600")
	}

	if fc.PublishedStrategyCacheTTLSec > 600 {
		return errors.New("invalid fsLocalCache.publishedStrategyCacheTTLSec value, should <= 600")
	}

	return nil
}

// trySetDefault try set the feed server's local cache default runtime if it's not set by user.
func (fc *FSLocalCache) trySetDefault() {
	if fc.AppCacheSize == 0 {
		fc.AppCacheSize = 100
	}

	if fc.AppCacheTTLSec == 0 {
		fc.AppCacheTTLSec = 1800
	}

	if fc.ReleasedInstanceCacheSize == 0 {
		fc.ReleasedInstanceCacheSize = 200
	}

	if fc.ReleasedInstanceCacheTTLSec == 0 {
		fc.ReleasedInstanceCacheTTLSec = 60
	}

	if fc.ReleasedCICacheSize == 0 {
		fc.ReleasedCICacheSize = 100
	}

	if fc.ReleasedCICacheTTLSec == 0 {
		fc.ReleasedCICacheTTLSec = 120
	}

	if fc.ReleasedKvCacheSize == 0 {
		fc.ReleasedKvCacheSize = 100
	}

	if fc.ReleasedKvCacheTTLSec == 0 {
		fc.ReleasedKvCacheTTLSec = 120
	}

	if fc.PublishedStrategyCacheSize == 0 {
		fc.PublishedStrategyCacheSize = 100
	}

	if fc.PublishedStrategyCacheTTLSec == 0 {
		fc.PublishedStrategyCacheTTLSec = 120
	}

	if fc.ReleasedGroupCacheSize == 0 {
		fc.ReleasedGroupCacheSize = 100
	}

	if fc.ReleasedGroupCacheTTLSec == 0 {
		fc.ReleasedGroupCacheTTLSec = 120
	}

	if fc.AuthCacheSize == 0 {
		fc.AuthCacheSize = 1000
	}

	if fc.AuthCacheTTLSec == 0 {
		fc.AuthCacheTTLSec = 300
	}

	if fc.CredentialCacheSize == 0 {
		fc.CredentialCacheSize = 5000
	}

	if fc.CredentialCacheTTLSec == 0 {
		fc.CredentialCacheTTLSec = 1
	}

	if fc.ReleasedHookCacheSize == 0 {
		fc.ReleasedHookCacheSize = 100
	}

	if fc.ReleasedHookCacheTTLSec == 0 {
		fc.ReleasedHookCacheTTLSec = 120
	}

	if fc.AsyncDownloadCacheSize == 0 {
		fc.AsyncDownloadCacheSize = 5000
	}

	if fc.AsyncDownloadCacheTTLSec == 0 {
		fc.AsyncDownloadCacheTTLSec = 600
	}
}

// Downstream define feed server downStream related settings.
type Downstream struct {
	// BounceIntervalHour the maximum grpc connection time from sidecar to the upstream feed server. feed
	// server will send this parameter to sidecar. if the connection between sidecar and feed server
	// reaches this interval, sidecar will re-select the feed server instance to establish the connection.
	// unit is hour, the minimum maxWatchTimeMin is 1, the maximum BounceIntervalHour is 48, and the
	// default BounceIntervalHour is 1.
	BounceIntervalHour uint `yaml:"bounceIntervalHour"`
	// NotifyMaxLimit is the concurrent number of goroutines which are used to broadcast app release messages to
	// sidecars, which are connnected to one feed server, when new app releases are published. The larger of it,
	// the more CPU and Mem will be costed.the minimum notifyMaxLimit is 10, the default notifyMaxLimit is 50.
	NotifyMaxLimit uint `yaml:"notifyMaxLimit"`
}

// validate if the feed server's release service runtime is valid or not.
func (f Downstream) validate() error {
	if f.BounceIntervalHour < 1 && f.BounceIntervalHour != 0 {
		return errors.New("invalid downstream.bounceIntervalHour value, should >= 1")
	}

	if f.BounceIntervalHour > 48 {
		return errors.New("invalid downstream.bounceIntervalHour value, should <= 48(two days)")
	}

	if f.NotifyMaxLimit < 10 && f.NotifyMaxLimit != 0 {
		return errors.New("invalid downstream.notifyMaxLimit value, should >= 10")
	}

	return nil
}

// trySetDefault try set the feed server's release service default runtime if it's not set by user.
func (f *Downstream) trySetDefault() {
	if f.BounceIntervalHour == 0 {
		f.BounceIntervalHour = 1
	}

	if f.NotifyMaxLimit == 0 {
		f.NotifyMaxLimit = 50
	}
}

// MatchReleaseLimiter defines the request limit options for match release.
type MatchReleaseLimiter struct {
	// QPS should >=1
	QPS uint `yaml:"qps"`
	// Burst should >= 1;
	Burst uint `yaml:"burst"`
	// WaitTimeMil is request wait time.
	WaitTimeMil uint `yaml:"waitTimeMil"`
}

// validate if the limiter is valid or not.
func (lm MatchReleaseLimiter) validate() error {
	if lm.QPS <= 0 {
		return errors.New("invalid matchReleaseLimiter.qps value, should >= 1")
	}

	if lm.Burst <= 0 {
		return errors.New("invalid matchReleaseLimiter.burst value, should >= 1")
	}

	if lm.WaitTimeMil <= 0 {
		return errors.New("invalid matchReleaseLimiter.waitTimeMil value, should >= 1")
	}

	return nil
}

// trySetDefault try set the default value of limiter
func (lm *MatchReleaseLimiter) trySetDefault() {
	if lm.QPS == 0 {
		lm.QPS = 500
	}

	if lm.Burst == 0 {
		lm.Burst = 500
	}

	if lm.WaitTimeMil == 0 {
		lm.WaitTimeMil = 50
	}
}

// RateLimiter defines the rate limiter options for traffic control.
type RateLimiter struct {
	Enable          bool    `yaml:"enable"`
	ClientBandwidth uint    `yaml:"clientBandwidth"`
	Global          BasicRL `yaml:"global"`
	Biz             BizRLs  `yaml:"biz"`
}

// BizRLs defines the rate limiters for biz
type BizRLs struct {
	Default BasicRL          `yaml:"default"`
	Spec    map[uint]BasicRL `yaml:"spec"`
}

// BasicRL defines the basic options for rate limiter.
type BasicRL struct {
	Limit uint `yaml:"limit"`
	Burst uint `yaml:"burst"`
}

const (
	// DefaultClientBandwidth default client bandwidth
	DefaultClientBandwidth = 50 // 50MB/s = 400Mb/s
	// DefaultGlobalRateLimit default global rate limit
	DefaultGlobalRateLimit = 1000 // 1000MB/s = 8000Mb/s
	// DefaultGlobalRateBurst default global rate burst
	DefaultGlobalRateBurst = 2000 // 2000MB = 16000Mb
	// DefaultBizRateLimit default biz rate limit
	DefaultBizRateLimit = 100 // 100MB/s = 800Mb/s
	// DefaultBizRateBurst default biz rate burst
	DefaultBizRateBurst = 200 // 200MB = 1600Mb
)

// validate if the rate limiter is valid or not.
func (rl RateLimiter) validate() error {
	if rl.Biz.Default.Burst < rl.Biz.Default.Limit {
		return fmt.Errorf("invalid rateLimiter.biz.default.burst value %d, should >= rateLimiter.biz.default.limit "+
			"value %d", rl.Global.Burst, rl.Global.Limit)
	}

	if rl.Global.Limit < rl.Biz.Default.Limit {
		return fmt.Errorf("invalid rateLimiter.global.limit value %d, should >= rateLimiter.biz.default.limit value %d",
			rl.Global.Limit, rl.ClientBandwidth)
	}

	if rl.Global.Burst < rl.Biz.Default.Burst {
		return fmt.Errorf("invalid rateLimiter.global.burst value %d, should >= rateLimiter.biz.default.burst value %d",
			rl.Global.Burst, rl.Global.Limit)
	}

	for bizID, l := range rl.Biz.Spec {
		if l.Burst < l.Limit {
			return fmt.Errorf("invalid rateLimiter.biz.spec.%d.burst value %d, "+
				"should >= rateLimiter.biz.spec.%d.limit value %d", bizID, l.Burst, bizID, l.Limit)
		}
	}

	return nil
}

// trySetDefault try set the default value of rate limiter
func (rl *RateLimiter) trySetDefault() {
	if rl.ClientBandwidth == 0 {
		rl.ClientBandwidth = DefaultClientBandwidth
	}

	if rl.Global.Limit == 0 {
		rl.Global.Limit = DefaultGlobalRateLimit
	}

	if rl.Global.Burst == 0 {
		rl.Global.Burst = DefaultGlobalRateBurst
	}

	if rl.Biz.Default.Limit == 0 {
		rl.Biz.Default.Limit = DefaultBizRateLimit
	}

	if rl.Biz.Default.Burst == 0 {
		rl.Biz.Default.Burst = DefaultBizRateBurst
	}

	for bizID, l := range rl.Biz.Spec {
		if l.Limit == 0 {
			rl.Biz.Spec[bizID] = BasicRL{
				Limit: DefaultBizRateLimit,
				Burst: l.Burst,
			}
		}
		if l.Burst == 0 {
			rl.Biz.Spec[bizID] = BasicRL{
				Limit: rl.Biz.Spec[bizID].Limit,
				Burst: DefaultBizRateBurst,
			}
		}
	}
}

// Credential credential encryption algorithm and master key
type Credential struct {
	MasterKey           string `yaml:"master_key"`
	EncryptionAlgorithm string `yaml:"encryption_algorithm"`
}

// validate credential options
func (c Credential) validate() error {

	if len(c.MasterKey) == 0 {
		return errors.New("credential master key is not set")
	}

	if len(c.EncryptionAlgorithm) == 0 {
		return errors.New("credential Encryption Algorithm is not set")
	}

	return nil
}

const (
	// VaultTokenEnv represents the environment variable key for the Vault token.
	VaultTokenEnv = "VAULT_TOKEN"
	// VaultAddressEnv represents the environment variable key for the Vault server address.
	VaultAddressEnv = "VAULT_ADDR"
)

// Vault Used to store the configuration information required for connecting to the Vault server.
type Vault struct {
	// Address is used to store the address of the Vault server
	Address string `yaml:"address"`
	// Token is used for accessing the Vault server
	Token string `yaml:"token"`
	// PluginDir is the directory where the Vault plugin is located
	PluginDir string `yaml:"pluginDir"`
	// UnsealKeys is used to unseal vault server
	UnsealKeys []string `yaml:"unsealKeys"`
	// SecretShares is the number of key shares to split the master key into
	SecretShares int `yaml:"secretShares"`
	// SecretThreshold is the number of key shares required to reconstruct the master key
	SecretThreshold int `yaml:"secretThreshold"`
}

// validate Vault options
func (v Vault) validate() error {

	if v.Address == "" {
		return errors.New("vault address is not set")
	}

	if v.Token == "" {
		return errors.New("vault token is not set")
	}

	return nil
}

// getConfigFromEnv Read configuration from environment variables
func (v *Vault) getConfigFromEnv() {

	if v.Token == "" {
		v.Token = os.Getenv(VaultTokenEnv)
	}
	if v.Address == "" {
		v.Address = os.Getenv(VaultAddressEnv)
	}
}

// BKNotice defines all the bk notice related runtime.
type BKNotice struct {
	Enable bool   `yaml:"enable"`
	Host   string `yaml:"host"`
}

// BCS defines all the bcs related runtime.
type BCS struct {
	Host  string `yaml:"host"`
	Token string `yaml:"token"`
}

// GSE defines all the gse related runtime.
type GSE struct {
	// Enabled is the flag to enable gse p2p download.
	Enabled bool `yaml:"enabled"`
	// Host is the gse bk api gateway address.
	Host string `yaml:"host"`
	// NodeAgentID is the node's agent id where feed server deployded, it might be different in different instance,
	// so recommend to get it from the environment variable.
	NodeAgentID string `yaml:"nodeAgentID"`
	// ClusterID is the cluster id where feed server deployded.
	ClusterID string `yaml:"clusterID"`
	// PodID is the pod's id where feed server deployded, it must be different in different instance,
	// so must get it from the environment variable.
	PodID string `yaml:"podID"`
	// ContainerName is the container's name of the feed server
	ContainerName string `yaml:"containerName"`
	// AgentUser is the user exists in the feed server container/node.
	AgentUser string `yaml:"agentUser"`
	// CacheDir is the directory where the source file download to and cached.
	CacheDir string `yaml:"cacheDir"`
	// CacheSizeGB is the cache size of the source file in the feed server.
	CacheSizeGB uint `yaml:"cacheSizeGB"`
	// CacheRetentionRate is the cache retention rate of the source file in the feed server.
	CacheRetentionRate float64 `yaml:"cacheRetentionRate"`
}

func (g *GSE) trySetDefault() {
	if g.AgentUser == "" {
		g.AgentUser = "root"
	}
	if g.CacheDir == "" {
		g.CacheDir = "/data/bscp/gse/cache"
	}
	if g.CacheSizeGB == 0 {
		g.CacheSizeGB = 10
	}
	if g.CacheRetentionRate == 0 {
		g.CacheRetentionRate = 0.5
	}
}

func (g *GSE) getFromEnv() {
	if len(g.NodeAgentID) == 0 {
		g.NodeAgentID = os.Getenv("NODE_AGENT_ID")
	}

	if len(g.ClusterID) == 0 {
		g.ClusterID = os.Getenv("CLUSTER_ID")
	}

	if len(g.PodID) == 0 {
		g.PodID = os.Getenv("POD_ID")
	}

	if len(g.ContainerName) == 0 {
		g.ContainerName = os.Getenv("CONTAINER_NAME")
	}
}

// validate gse runtime
func (g GSE) validate() error {
	if !g.Enabled {
		return nil
	}
	if g.Host == "" {
		return errors.New("gse host is not set")
	}
	if g.NodeAgentID == "" && (g.ClusterID == "" || g.PodID == "" || g.ContainerName == "") {
		return errors.New("to enable p2p download, either agent id must be set or cluster id, " +
			"pod id, container name must all be set")
	}
	return nil
}

// Gorm defines the grom related settings.
type Gorm struct {
	LogLevel GormLogLevel `yaml:"logLevel"`
}

// GormLogLevel is gorm log level type
type GormLogLevel string

const (
	// GormLogSilent used for gorm log level
	GormLogSilent GormLogLevel = "silent"
	// GormLogError used for gorm log level
	GormLogError GormLogLevel = "error"
	// GormLogWarn used for gorm log level
	GormLogWarn GormLogLevel = "warn"
	// GormLogInfo used for gorm log level
	GormLogInfo GormLogLevel = "info"
)

// GetLogLevel get log level for gorm.
func (g Gorm) GetLogLevel() logger.LogLevel {
	switch strings.ToLower(string(g.LogLevel)) {
	case string(GormLogSilent):
		return logger.Silent
	case string(GormLogError):
		return logger.Error
	case string(GormLogWarn):
		return logger.Warn
	case string(GormLogInfo):
		return logger.Info
	default:
		return logger.Info
	}
}

// validate if the gorm is valid or not.
func (g Gorm) validate() error {
	switch strings.ToLower(string(g.LogLevel)) {
	case string(GormLogSilent):
	case string(GormLogError):
	case string(GormLogWarn):
	case string(GormLogInfo):
	default:
		return fmt.Errorf("unsopported log level: %s", g.LogLevel)
	}
	return nil
}

// trySetDefault try set the default value of gorm
func (g *Gorm) trySetDefault() {
	if g.LogLevel == "" {
		g.LogLevel = GormLogInfo
	}
}
