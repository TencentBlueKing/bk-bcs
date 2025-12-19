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

// Package options xxx
package options

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	k8swatch "k8s.io/client-go/tools/watch"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/apiclient"
)

// ImageProxyOption defines the option of image-proxy
type ImageProxyOption struct {
	conf.FileConfig
	conf.LogConfig

	Address       string `json:"address" value:"0.0.0.0" usage:"the address"`
	HTTPPort      int64  `json:"httpPort" value:"2080" usage:"the http port"`
	HTTPSPort     int64  `json:"httpsPort" value:"2081" usage:"the https port"`
	MetricPort    int64  `json:"metricPort" value:"2082" usage:"the metric port"`
	TorrentPort   int64  `json:"torrentPort" value:"2083" usage:"the torrent port"`
	RedisAddress  string `json:"redisAddress" value:"" usage:""`
	RedisPassword string `json:"redisPassword" value:"" usage:"the password of redis"`

	// DisableTorrent 关闭 Torrent 传输
	DisableTorrent bool `json:"disableTorrent" value:"false" usage:"disable torrent transfer"`
	// EnableDockerd 开启 Dockerd 支持
	EnableDockerd bool `json:"enableDockerd" value:"false" usage:"enable dockerd"`
	// EnableContainerd 开启 Containerd 支持
	EnableContainerd bool `json:"enableContainerd" value:"false" usage:"enable containerd"`
	// TorrentThreshold Torrent 传输文件的阈值，超过才使用 Torrent 传输
	TorrentThreshold int64 `json:"torrentThreshold" value:"209715200" usage:"transfer by torrent if size exceeded the threshold"`
	// TorrentUploadLimit 种子上传速度限制，0 表示无限制
	TorrentUploadLimit int64 `json:"torrentUploadLimit" value:"0" usage:"upload limit"`
	// TorrentDownloadLimit 种子下载速度限制，0 表示无限制
	TorrentDownloadLimit int64 `json:"torrentDownloadLimit" value:"0" usage:"download limit"`

	// 用于从源仓库下载中 Layer 的存储目录，其下文件并不能保证完整性
	StoragePath string `json:"storagePath" value:"/data/bcs-image-proxy/storage" usage:"the path for download layer from remote original registry, just a temp storage"`
	// 用于 Torrent 下载的目录，文件不保证完整性
	TorrentPath string `json:"torrentPath" value:"/data/bcs-image-proxy/torrent" usage:"the path for torrent"`
	// 常规下载的 Layer 文件存储，其下文件保证完整性
	TransferPath string `json:"transferPath" value:"/data/bcs-image-proxy/transfer" usage:"the path for transfer file with bittorrent"`
	// 小文件，其下文件保证完整性
	SmallFilePath string `json:"smallFilePath" value:"/data/bcs-image-proxy/smallfile" usage:"the path of small files"`
	// 存储 docker/containerd 纳管的 Layer 缓存下来的文件，保证完整性
	OCIPath string `json:"ociPath" value:"/data/bcs-image-proxy/oci" usage:"the path of oci image"`

	EventFile       string `json:"eventFile" value:"/data/bcs-image-proxy/image.event" usage:"event record for image pull"`
	TorrentAnnounce string `json:"torrentAnnounce" value:"" usage:"the announce of torrent"`

	// PreferConfig 配置优先策略，用户可以指定 Master，以及指定某些节点作为优选的角色
	PreferConfig PreferConfig `json:"preferConfig" value:"" usage:"prefer config"`

	// CleanConfig 配置清理策略，支持用户配置清理时间、磁盘占用的阈值，以及保留最近多少天的数据
	CleanConfig CleanConfig `json:"cleanConfig" value:"" usage:"clean config"`

	ExternalConfigPath string         `json:"externalConfigPath" value:"" usage:"external config path"`
	ExternalConfig     ExternalConfig `json:"-"`
	k8sClient          *kubernetes.Clientset
}

// CleanConfig defines the clean config
type CleanConfig struct {
	Cron       string `json:"cron" value:"" usage:"the cron expression"`
	Threshold  int64  `json:"threshold" value:"100" usage:"the threshold of disk, unit: GB"`
	RetainDays int64  `json:"retainDays" value:"0" usage:"the day that need retain"`
}

// PreferConfig defines the prefer config
type PreferConfig struct {
	MasterIP    string            `json:"masterIP" value:"" usage:"manually specify the master node"`
	PreferNodes PreferNodesConfig `json:"preferNodes" value:"" usage:"assume the master role and download tasks"`
}

// PreferNodesConfig defines the prefer nodes config
type PreferNodesConfig struct {
	LabelSelectors string `json:"labelSelectors" usage:"the label selector to filter nodes"`
}

// ProxyType defines proxy type
type ProxyType string

const (
	// DomainProxy domain proxy
	DomainProxy ProxyType = "DomainProxy"
	// RegistryMirror registry mirror
	RegistryMirror ProxyType = "RegistryMirror"
)

// FilterRegistryMapping filter registry mapping
func (o *ImageProxyOption) FilterRegistryMapping(proxyHost string, proxyType ProxyType) *RegistryMapping {
	// 针对 ProxyHost 为空，设置其默认使用 docker.io
	if proxyHost == "" {
		return &o.ExternalConfig.DockerHubRegistry
	}
	for _, m := range o.ExternalConfig.RegistryMappings {
		switch proxyType {
		case RegistryMirror:
			// for containerd
			if proxyHost == m.OriginalHost {
				return m
			}
			// for dockerd
			if proxyHost == m.ProxyHost {
				return m
			}
		case DomainProxy:
			if proxyHost == m.ProxyHost {
				return m
			}
		}
	}
	return nil
}

// FilterRegistryMappingByOriginal filter registry mappings by original registry
func (o *ImageProxyOption) FilterRegistryMappingByOriginal(originalHost string) *RegistryMapping {
	if o.ExternalConfig.DockerHubRegistry.OriginalHost == originalHost {
		return &o.ExternalConfig.DockerHubRegistry
	}
	for _, m := range o.ExternalConfig.RegistryMappings {
		if originalHost == m.OriginalHost {
			return m
		}
	}
	return nil
}

// GetK8SClient get k8s client
func (o *ImageProxyOption) GetK8SClient() *kubernetes.Clientset {
	return o.k8sClient
}

func mapKeys(m map[string]struct{}) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func (o *ImageProxyOption) getServiceEndpoints(ns, name string) ([]string, error) {
	var preferNodes = make(map[string]struct{})
	// 获取到用户设置的 Prefer 节点信息列表
	if selectors := o.PreferConfig.PreferNodes.LabelSelectors; selectors != "" {
		nodeList, err := o.k8sClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{
			LabelSelector: selectors,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "list nodes by selector '%s' failed", selectors)
		}
		for i := range nodeList.Items {
			addresses := nodeList.Items[i].Status.Addresses
			for _, address := range addresses {
				if address.Type != corev1.NodeInternalIP {
					continue
				}
				preferNodes[address.Address] = struct{}{}
			}
		}
		if len(preferNodes) == 0 {
			blog.Warnf("[master-election] there not get any nodes by prefer-selectors: %s", selectors)
		}
	}

	// 获取到所有 Image-Proxy 的 Endpoint 节点列表
	eps, err := o.k8sClient.CoreV1().Endpoints(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "get k8s endpoint '%s/%s' failed", ns, name)
	}
	epMap := make(map[string]struct{})
	intersection := map[string]struct{}{}
	for i := range eps.Subsets {
		ep := &eps.Subsets[i]
		for j := range ep.Addresses {
			epIP := ep.Addresses[j].IP
			epMap[epIP] = struct{}{}
			if _, ok := preferNodes[epIP]; ok {
				intersection[epIP] = struct{}{}
			}
		}
	}
	result := make([]string, 0)
	if len(intersection) != 0 {
		blog.Infof("[master-election] get prefer-nodes with endpoints intersection: %v",
			mapKeys(intersection))
		// 如果存在交集，应该把交集返回回去
		for k := range intersection {
			result = append(result, k)
		}
	} else {
		// 如果无交集，应该把 Endpoints 返回回去
		for k := range epMap {
			result = append(result, k)
		}
	}
	if masterIP := o.PreferConfig.MasterIP; masterIP != "" {
		if _, ok := epMap[masterIP]; !ok {
			blog.Warnf("[master-election] preferConfig.masterIP is specified '%s', but not found", masterIP)
		} else {
			had := false
			// 对比结果是否有 MasterIP，如果没有需要加进去
			for _, ip := range result {
				if ip == masterIP {
					had = true
					break
				}
			}
			if !had {
				result = append(result, masterIP)
			}
		}
	}
	newResult := make([]string, 0, len(result))
	for _, ip := range result {
		newResult = append(newResult, fmt.Sprintf("%s:%d", ip, op.HTTPPort))
	}
	// blog.Infof("[master-election] get service endpoints: %d", len(newResult))
	return newResult, nil
}

func (o *ImageProxyOption) changeMaster(prevMaster, ns, name string) string {
	result, err := o.getServiceEndpoints(ns, name)
	if err != nil {
		blog.Errorf("get service endpoints failed: %s", err.Error())
	} else {
		o.ExternalConfig.LeaderConfig.Endpoints = result
		currentMaster := o.CurrentMaster()
		if prevMaster != o.CurrentMaster() {
			blog.Infof("current master: %s => %s", prevMaster, currentMaster)
			return currentMaster
		}
	}
	return prevMaster
}

// WatchK8sService watch k8s service
func (o *ImageProxyOption) WatchK8sService() error {
	ns := o.ExternalConfig.LeaderConfig.ServiceNamespace
	name := o.ExternalConfig.LeaderConfig.ServiceName
	result, err := o.getServiceEndpoints(ns, name)
	if err != nil {
		return err
	}
	o.ExternalConfig.LeaderConfig.Endpoints = result
	prevMaster := o.CurrentMaster()
	blog.Infof("current master: %s", prevMaster)

	ctx := context.Background()
	watcher, err := k8swatch.NewRetryWatcherWithContext(ctx, "endpoints", &cache.ListWatch{
		WatchFuncWithContext: func(ctx context.Context, options metav1.ListOptions) (watch.Interface, error) {
			return o.k8sClient.CoreV1().Endpoints(ns).Watch(ctx, metav1.ListOptions{
				ResourceVersion: "0",
				FieldSelector:   fmt.Sprintf("metadata.name=%s", name),
			})
		},
	})
	if err != nil {
		return errors.Wrapf(err, "create k8s watcher for endpoints failed")
	}
	go func() {
		defer watcher.Stop()
		blog.Infof("watching k8s endpoint '%s/%s'", ns, name)
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.ResultChan():
				if !ok {
					blog.Errorf("watch k8s endpoints channel interrupted")
					return
				}
				if event.Object == nil {
					blog.Errorf("watch k8s endpoints event.object is nil")
					continue
				}
				switch event.Type {
				case watch.Added, watch.Modified, watch.Deleted:
					prevMaster = o.changeMaster(prevMaster, ns, name)
				case watch.Error:
					fmt.Printf("watch k8s endpoints error occurred: %v\n", event.Object)
					return
				}
			}
		}
	}()
	return nil
}

// StringASCII return the acii of string
func StringASCII(str string) int64 {
	var result int64
	for i := range str {
		// take the index to do weighting
		result += int64(i) + int64(str[i])
	}
	return result
}

// CurrentMaster return the current master
func (o *ImageProxyOption) CurrentMaster() string {
	var currentASCII int64 = 0
	var currentEndpoint string
	masterIP := o.PreferConfig.MasterIP
	for i := range o.ExternalConfig.LeaderConfig.Endpoints {
		ep := o.ExternalConfig.LeaderConfig.Endpoints[i]
		if masterIP != "" && strings.HasPrefix(ep, masterIP+":") {
			return ep
		}
		ascii := StringASCII(ep)
		if currentASCII < ascii {
			currentASCII = ascii
			currentEndpoint = ep
		}
	}
	return currentEndpoint
}

// IsMaster return current is-master
func (o *ImageProxyOption) IsMaster() bool {
	master := o.CurrentMaster()
	return master == fmt.Sprintf("%s:%d", o.Address, o.HTTPPort)
}

// HTTPProxyTransport return the insecure-skip-verify transport
func (o *ImageProxyOption) HTTPProxyTransport() http.RoundTripper {
	netDialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	tp := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           netDialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// NOTE: insecure for original registry
		// NOCC:gas/tls(设计如此)
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if httpProxyUrl == nil {
		return tp
	}
	tp.Proxy = http.ProxyURL(httpProxyUrl)
	return tp
}

// LocalhostCert defines localhost proxy
const LocalhostCert = "localhost"

// ProxyKeyCert defines the key/cert for proxy host
type ProxyKeyCert struct {
	Key  string `json:"key"`
	Cert string `json:"cert"`
}

// ExternalConfig defines the external config
type ExternalConfig struct {
	Enable            bool                     `json:"enable"`
	HTTPProxy         string                   `json:"httpProxy"`
	BuiltInCerts      map[string]*ProxyKeyCert `json:"builtInCerts"`
	DockerHubRegistry RegistryMapping          `json:"dockerHubRegistry"`
	RegistryMappings  []*RegistryMapping       `json:"registryMappings"`
	LeaderConfig      LeaderConfig             `json:"leaderConfig,omitempty"`
}

// RegistryMapping defines the mapping for original registry with proxy. There also defines the
// username/password for registry when use RegistryMirror mode.
type RegistryMapping struct {
	ProxyHost    string `json:"proxyHost"`
	ProxyCert    string `json:"proxyCert"`
	ProxyKey     string `json:"proxyKey"`
	OriginalHost string `json:"originalHost"`

	Username string          `json:"username"`
	Password string          `json:"password"`
	Users    []*RegistryAuth `json:"users,omitempty"`
	// 用户多个用户名/密码，临时记录正确的内容
	CorrectUser string `json:"-"`
	CorrectPass string `json:"-"`
}

// RegistryAuth defines the user/pass for registry
type RegistryAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LeaderConfig defines the config of leader. There can specify service's namespace/name for kubernetes, also
// can specify endpoints for not kubernetes.
type LeaderConfig struct {
	ServiceNamespace string `json:"serviceNamespace"`
	ServiceName      string `json:"serviceName"`

	Endpoints []string `json:"endpoints"`
}

var (
	op               = new(ImageProxyOption)
	defaultDockerHub = "registry-1.docker.io"
	httpProxyUrl     *url.URL
)

func (o *ImageProxyOption) checkFilePath() error {
	if op.TorrentThreshold < apiclient.TwoHundredMB {
		op.TorrentThreshold = apiclient.TwoHundredMB
	}
	if op.TorrentUploadLimit > 0 && op.TorrentUploadLimit < 1048576 {
		return errors.Errorf("upload limit '%d' too small, must >= 1048576(1MB/s)", op.TorrentUploadLimit)
	}
	if op.TorrentDownloadLimit > 0 && op.TorrentDownloadLimit < 1048576 {
		return errors.Errorf("download limit '%d' too small, must >= 1048576(1MB/s)", op.TorrentUploadLimit)
	}

	if err := os.MkdirAll(op.TransferPath, 0600); err != nil {
		return errors.Wrapf(err, "create file-path '%s' failed", op.TransferPath)
	}
	if err := os.MkdirAll(op.StoragePath, 0600); err != nil {
		return errors.Wrapf(err, "create file-path '%s' failed", op.StoragePath)
	}
	if err := os.MkdirAll(op.SmallFilePath, 0600); err != nil {
		return errors.Wrapf(err, "create file-path '%s' failed", op.SmallFilePath)
	}
	// should remove torrentPath to avoid some cached files
	_ = os.RemoveAll(op.TorrentPath)
	if err := os.MkdirAll(op.TorrentPath, 0600); err != nil {
		return errors.Wrapf(err, "create file-path '%s' failed", op.TorrentPath)
	}
	if err := os.MkdirAll(op.OCIPath, 0600); err != nil {
		return errors.Wrapf(err, "create file-path '%s' failed", op.OCIPath)
	}
	return nil
}

func (o *ImageProxyOption) parseConExpression() error {
	if o.CleanConfig.Cron == "" {
		blog.Infof("clean-config not set, no-need auto clean")
		return nil
	}
	parser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)
	schedule, err := parser.Parse(o.CleanConfig.Cron)
	if err != nil {
		return errors.Wrapf(err, "parse cron expression '%s' failed", o.CleanConfig.Cron)
	}
	if o.CleanConfig.Threshold < 10 {
		o.CleanConfig.Threshold = 10
	}
	if o.CleanConfig.RetainDays < 0 {
		o.CleanConfig.RetainDays = 0
	}

	blog.Infof("clean-config is set '%s', retain: %d day, size: %d GB print the next ten execution times:",
		o.CleanConfig.Cron, o.CleanConfig.RetainDays, o.CleanConfig.Threshold)
	currentTime := time.Now()
	for i := 0; i < 10; i++ {
		currentTime = schedule.Next(currentTime)
		blog.Infof("  [%d] %s", i, currentTime.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func (o *ImageProxyOption) checkExternalConfig() error {
	if op.ExternalConfigPath == "" {
		return errors.Errorf("config 'externalConfigPath' cannot be empty")
	}
	bs, err := os.ReadFile(op.ExternalConfigPath)
	if err != nil {
		return errors.Wrapf(err, "external config '%s' read failed", op.ExternalConfigPath)
	}
	op.ExternalConfig = ExternalConfig{}
	if err = json.Unmarshal(bs, &op.ExternalConfig); err != nil {
		return errors.Wrapf(err, "external config '%s' unmarshal failed", op.ExternalConfigPath)
	}

	op.ExternalConfig.HTTPProxy = strings.TrimSpace(op.ExternalConfig.HTTPProxy)
	if op.ExternalConfig.HTTPProxy != "" {
		httpProxyUrl, err = url.Parse(op.ExternalConfig.HTTPProxy)
		if err != nil {
			return errors.Wrapf(err, "set http_proxy '%s' failed", op.ExternalConfig.HTTPProxy)
		}
		if err = checkNetConnectivity(op.ExternalConfig.HTTPProxy); err != nil {
			return errors.Wrapf(err, "check http_proxy connectivity failed")
		}
		blog.Infof("set http_proxy '%s' success", op.ExternalConfig.HTTPProxy)
	} else {
		blog.Infof("not use http_proxy")
	}
	return nil
}

// checkNetConnectivity check whether the target can connect
func checkNetConnectivity(target string) error {
	afterTrim := strings.TrimPrefix(strings.TrimPrefix(target, "http://"), "https://")
	conn, err := net.DialTimeout("tcp", afterTrim, 5*time.Second)
	if err != nil {
		return errors.Wrapf(err, "dial target '%s' failed", target)
	}
	defer conn.Close()
	return nil
}

func (o *ImageProxyOption) checkExternalConfigBuiltInCerts() error {
	if op.ExternalConfig.BuiltInCerts == nil {
		return nil
	}
	for _, v := range op.ExternalConfig.BuiltInCerts {
		keyBase64, err := base64.StdEncoding.DecodeString(v.Key)
		if err != nil {
			return errors.Wrapf(err, "base64 decode built-in key '%s' failed", v.Key)
		}

		var certBase64 []byte
		certBase64, err = base64.StdEncoding.DecodeString(v.Cert)
		if err != nil {
			return errors.Wrapf(err, "base64 decode built-in cert '%s' failed", v.Cert)
		}
		v.Key = string(keyBase64)
		v.Cert = string(certBase64)
	}

	for _, mp := range op.ExternalConfig.RegistryMappings {
		v, ok := op.ExternalConfig.BuiltInCerts[mp.ProxyHost]
		if ok {
			mp.ProxyCert = v.Cert
			mp.ProxyKey = v.Key
			continue
		}
		if mp.ProxyCert != "" && mp.ProxyKey != "" {
			afterBase64, err := base64.StdEncoding.DecodeString(mp.ProxyCert)
			if err != nil {
				return errors.Wrapf(err, "base64 decode cert '%s' failed", mp.ProxyCert)
			}
			mp.ProxyCert = string(afterBase64)
			afterBase64, err = base64.StdEncoding.DecodeString(mp.ProxyKey)
			if err != nil {
				return errors.Wrapf(err, "base64 decode key '%s' failed", mp.ProxyKey)
			}
			mp.ProxyKey = string(afterBase64)
		}
	}
	return nil
}

func (o *ImageProxyOption) checkExternalConfigRegistries() error {
	for _, mp := range o.ExternalConfig.RegistryMappings {
		switch {
		case mp.Username != "" && mp.Password != "":
			blog.Infof("registry '%s' set user '%s' and password '%s'", mp.OriginalHost,
				mp.Username, mp.Password)
			continue
		case mp.Username != "" && mp.Password == "":
			return errors.Errorf("registry '%s' set user '%s' but not set password", mp.OriginalHost, mp.Username)
		case mp.Username == "" && mp.Password != "":
			return errors.Errorf("registry '%s' set password but not set user", mp.OriginalHost)
		case mp.Username == "" && mp.Password == "" && len(mp.Users) == 0:
			blog.Warnf("registry '%s' not set any user/passwords")
			continue
		}
		for i, auth := range mp.Users {
			if auth.Username == "" || auth.Password == "" {
				return errors.Errorf("registry '%s' users[%d] not set user or password", mp.OriginalHost, i)
			}
			blog.Infof("registry '%s' set user '%s' and password '%s'", mp.OriginalHost,
				auth.Username, auth.Password)
		}
	}
	return nil
}

func (o *ImageProxyOption) checkExternalConfigLeaderConfig() error {
	// should create kubernetes client if it deployed on kubernetes
	if op.ExternalConfig.LeaderConfig.ServiceNamespace != "" && op.ExternalConfig.LeaderConfig.ServiceName != "" {
		blog.Infof("server deployed on kubernetes env")
		config, err := rest.InClusterConfig()
		if err != nil {
			return errors.Wrapf(err, "get kubernetes in-cluster config failed")
		}
		op.k8sClient, err = kubernetes.NewForConfig(config)
		if err != nil {
			return errors.Wrapf(err, "create kubernetes client failed")
		}
		if err = op.WatchK8sService(); err != nil {
			return errors.Wrapf(err, "watch kubernetes service failed")
		}
	} else {
		blog.Infof("server deployed on non-kubernetes env")
		if len(op.ExternalConfig.LeaderConfig.Endpoints) == 0 {
			return errors.Errorf("server deployed on non-kubernetes env, but no endpoints is set")
		}
	}
	return nil
}

// Parse the config options
func Parse() *ImageProxyOption {
	conf.Parse(op)
	blog.InitLogs(op.LogConfig)

	if err := op.checkFilePath(); err != nil {
		blog.Fatalf("check filepath failed: %s", err.Error())
	}
	if err := op.parseConExpression(); err != nil {
		blog.Fatalf("parse cron failed: %s", err.Error())
	}
	if err := op.checkExternalConfig(); err != nil {
		blog.Fatalf("check external config failed: %s", err.Error())
	}
	if err := op.checkExternalConfigBuiltInCerts(); err != nil {
		blog.Fatalf("check external config built-in certs failed: %s", err.Error())
	}
	if err := op.checkExternalConfigLeaderConfig(); err != nil {
		blog.Fatalf("check external config leader config failed: %s", err.Error())
	}
	if err := op.checkExternalConfigRegistries(); err != nil {
		blog.Fatalf("check external config registries failed: %s", err.Error())
	}
	if op.ExternalConfig.DockerHubRegistry.OriginalHost == "" {
		op.ExternalConfig.DockerHubRegistry.OriginalHost = defaultDockerHub
	}
	return op
}

// GlobalOptions returns the global option
func GlobalOptions() *ImageProxyOption {
	return op
}
