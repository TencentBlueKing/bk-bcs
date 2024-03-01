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

package bcsegress

import (
	"crypto/md5" // NOCC:gas/crypto(设计如此)
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"sync"
	"text/template"
	"time"

	"k8s.io/klog"
)

const (
	defaultTemplateFile      = "./template/nginx-template.conf"
	defaultGenerateDirectory = "./generate/"
)

// Proxy interface for http/tcp network flow controlle
type Proxy interface {
	// GetHTTPRule ...
	// http server operation part
	GetHTTPRule(key string) (*HTTPConfig, error)
	ListHTTPRules() ([]*HTTPConfig, error)
	ListHTTPRulesByLabel(labels map[string]string) ([]*HTTPConfig, error)
	DeleteHTTPRule(key string) error
	UpdateHTTPRule(cfg *HTTPConfig) error

	// GetTCPRule ...
	// upstream tcp operation part
	GetTCPRule(key string) (*TCPConfig, error)
	GetTCPRuleByPort(port uint) (*TCPConfig, error)
	ListTCPRules() ([]*TCPConfig, error)
	ListTCPRulesByLabel(labels map[string]string) ([]*TCPConfig, error)
	DeleteTCPRule(key string) error
	UpdateTCPRule(cfg *TCPConfig) error

	// Reload proxy for new configuration. egress is the rule reference why proxy
	// need to reload, proxy stores error information relative to this egress rule
	// it's convenience for user to check egress last error for decision of reloading again
	Reload(egress string) error
	// LastError get last reload error information according to egress rule
	LastError(egress string) error
}

// generator data for template creation
type generator struct {
	TCPServer TCPList
}

// NewNginx create nginx instance as proxy implementation
func NewNginx(option *EgressOption) (Proxy, error) {
	if len(option.ProxyExecutable) == 0 {
		option.ProxyExecutable = nginxExecutable
	}
	if len(option.ProxyConfig) == 0 {
		option.ProxyConfig = nginxConfig
	}
	if len(option.GenerateDir) == 0 {
		option.GenerateDir = defaultGenerateDirectory
	}
	if len(option.TemplateFile) == 0 {
		option.TemplateFile = defaultTemplateFile
	}
	ngx := &Nginx{
		option:         option,
		tcpKeyConfigs:  make(map[string]*TCPConfig),
		tcpPortConfigs: make(map[uint]*TCPConfig),
		httpConfigs:    make(map[string]*HTTPConfig),
		lastError:      make(map[string]error),
	}
	// ensure workspace directory existence
	exist, err := fileExists(option.ProxyExecutable)
	if err != nil || !exist {
		klog.Errorf("Nginx proxy Executable %s is Lost", option.ProxyExecutable)
		return nil, fmt.Errorf("proxy %s lost", option.ProxyExecutable)
	}
	exist, err = fileExists(option.ProxyConfig)
	if err != nil || !exist {
		klog.Errorf("Nginx proxy init config %s is Lost", option.ProxyExecutable)
		return nil, fmt.Errorf("proxy config %s lost", option.ProxyConfig)
	}
	exist, err = fileExists(option.TemplateFile)
	if err != nil || !exist {
		klog.Errorf("Nginx proxy config template file %s is Lost", option.TemplateFile)
		return nil, fmt.Errorf("proxy config template file %s lost", option.TemplateFile)
	}
	err = os.MkdirAll(option.GenerateDir, os.ModePerm)
	if err != nil {
		klog.Warningf("mkdir %s failed, err %s", option.GenerateDir, err.Error())
		return nil, err
	}
	return ngx, nil
}

const (
	nginxExecutable = "/usr/local/nginx/sbin/nginx"
	nginxConfig     = "/usr/local/nginx/conf/nginx.conf"
)

// Nginx implementations for proxy interface
type Nginx struct {
	option *EgressOption
	// tcpLock for follow cachedata
	tcpLock sync.RWMutex
	// Key is indexer
	tcpKeyConfigs map[string]*TCPConfig
	// Port is indexer
	tcpPortConfigs map[uint]*TCPConfig
	// http data support
	httpLock sync.RWMutex
	// domain & port is indexer
	httpConfigs map[string]*HTTPConfig
	// egress rule error for last update
	errorLock sync.RWMutex
	lastError map[string]error
}

// GetHTTPRule get specified http rule implementation
func (ngx *Nginx) GetHTTPRule(key string) (*HTTPConfig, error) {
	ngx.httpLock.RLock()
	defer ngx.httpLock.Unlock()
	config, ok := ngx.httpConfigs[key]
	if ok {
		return config, nil
	}
	return nil, nil
}

// ListHTTPRules list all http rules implementation
func (ngx *Nginx) ListHTTPRules() ([]*HTTPConfig, error) {
	if len(ngx.httpConfigs) == 0 {
		return nil, nil
	}
	var l []*HTTPConfig
	for _, config := range ngx.httpConfigs {
		l = append(l, config)
	}
	return l, nil
}

// ListHTTPRulesByLabel http operation implementation
func (ngx *Nginx) ListHTTPRulesByLabel(labels map[string]string) ([]*HTTPConfig, error) {
	ngx.httpLock.RLock()
	defer ngx.httpLock.Unlock()
	if len(ngx.httpConfigs) == 0 {
		return nil, nil
	}
	var l []*HTTPConfig
	for _, config := range ngx.httpConfigs {
		if config.LabelFilter(labels) {
			l = append(l, config)
		}
	}
	return l, nil
}

// DeleteHTTPRule delete specified http rule implementation
func (ngx *Nginx) DeleteHTTPRule(key string) error {
	ngx.httpLock.Lock()
	defer ngx.httpLock.Unlock()
	delete(ngx.httpConfigs, key)
	return nil
}

// UpdateHTTPRule update specified http rule implementation
func (ngx *Nginx) UpdateHTTPRule(cfg *HTTPConfig) error {
	ngx.httpLock.Lock()
	defer ngx.httpLock.Unlock()
	ngx.httpConfigs[cfg.Key()] = cfg
	return nil
}

// GetTCPRule tcp operation implementation
func (ngx *Nginx) GetTCPRule(key string) (*TCPConfig, error) {
	ngx.tcpLock.RLock()
	defer ngx.tcpLock.Unlock()
	config, ok := ngx.tcpKeyConfigs[key]
	if ok {
		return config, nil
	}
	return nil, nil
}

// GetTCPRuleByPort tcp operation implementation
func (ngx *Nginx) GetTCPRuleByPort(port uint) (*TCPConfig, error) {
	ngx.tcpLock.RLock()
	defer ngx.tcpLock.Unlock()
	config, ok := ngx.tcpPortConfigs[port]
	if ok {
		return config, nil
	}
	return nil, nil
}

// ListTCPRules tcp operation implementation
func (ngx *Nginx) ListTCPRules() ([]*TCPConfig, error) {
	ngx.tcpLock.RLock()
	defer ngx.tcpLock.Unlock()
	if len(ngx.tcpPortConfigs) != len(ngx.tcpKeyConfigs) {
		return nil, fmt.Errorf("nginx proxy tcp configuration is inconsistent")
	}
	var l []*TCPConfig
	for _, config := range ngx.tcpPortConfigs {
		l = append(l, config)
	}
	return l, nil
}

// ListTCPRulesByLabel tcp operation implementation
func (ngx *Nginx) ListTCPRulesByLabel(labels map[string]string) ([]*TCPConfig, error) {
	ngx.tcpLock.RLock()
	defer ngx.tcpLock.Unlock()
	if len(ngx.tcpPortConfigs) != len(ngx.tcpKeyConfigs) {
		return nil, fmt.Errorf("nginx proxy tcp configuration is inconsistent")
	}
	var l []*TCPConfig
	for _, config := range ngx.tcpPortConfigs {
		if config.LabelFilter(labels) {
			l = append(l, config)
		}
	}
	return l, nil
}

// DeleteTCPRule tcp operation implementation
func (ngx *Nginx) DeleteTCPRule(key string) error {
	ngx.tcpLock.Lock()
	defer ngx.tcpLock.Unlock()
	config, ok := ngx.tcpKeyConfigs[key]
	if !ok {
		return nil
	}
	delete(ngx.tcpKeyConfigs, key)
	_, pok := ngx.tcpPortConfigs[config.ProxyPort]
	if !pok {
		klog.Warningf("nginx proxy tcp port [%d] data is inconsistent with key data %s", config.ProxyPort, key)
		return nil
	}
	delete(ngx.tcpPortConfigs, config.ProxyPort)
	return nil
}

// UpdateTCPRule tcp operation implementation
func (ngx *Nginx) UpdateTCPRule(cfg *TCPConfig) error {
	ngx.tcpLock.Lock()
	defer ngx.tcpLock.Unlock()
	// upate port reference
	ngx.tcpKeyConfigs[cfg.Key()] = cfg
	ngx.tcpPortConfigs[cfg.ProxyPort] = cfg
	return nil
}

// Reload reload proxy for new configuration
func (ngx *Nginx) Reload(egress string) error {
	ngx.tcpLock.Lock()
	defer ngx.tcpLock.Unlock()
	ngx.httpLock.RLock()
	defer ngx.httpLock.Unlock()
	ngx.errorLock.Lock()
	defer ngx.errorLock.Unlock()
	// ready to generate nginx configuration from template
	output, err := ngx.configGeneration()
	if err != nil {
		klog.Errorf("proxy nginx config for egress [%s] generated %s failed, %s", egress, output, err.Error())
		ngx.lastError[egress] = err
		return err
	}
	// configuration validation
	if err = ngx.configValidation(output); err != nil {
		klog.Errorf("proxy nginx check egress %s new configuration %s failed, %s", egress, output, err.Error())
		ngx.lastError[egress] = err
		return err
	}
	changed, err := ngx.isConfigChanged(output)
	if err != nil {
		klog.Errorf("proxy nginx check all generation configuration %s for egress %s failed, %s", output, egress, err.Error())
		ngx.lastError[egress] = err
		return err
	}
	if !changed {
		klog.Warningf("proxy nginx new configuration %s for %s nothing changed with original one, skip reloading", output,
			egress)
		delete(ngx.lastError, egress)
		return nil
	}
	if err := ngx.reloadNginx(output); err != nil {
		// reload failed, recording last error message for decision of reloading again
		klog.Errorf("proxy nginx reload configuration %s for egress %s failed, %s", output, egress, err.Error())
		ngx.lastError[egress] = err
		return err
	}
	// reload successfully, clean relative last error
	ngx.lastError = make(map[string]error)
	return nil
}

func (ngx *Nginx) configGeneration() (string, error) {
	t, err := template.ParseFiles(ngx.option.TemplateFile)
	if err != nil {
		klog.Errorf("proxy nginx read configuration template %s failed, %s", ngx.option.TemplateFile, err.Error())
		return "", err
	}
	// create output file
	stamp := strconv.Itoa(int(time.Now().Unix()))
	filename := "nginx." + stamp + ".conf"
	output := path.Join(ngx.option.GenerateDir, filename)
	writer, err := os.Create(output)
	if err != nil {
		klog.Errorf("proxy nginx create generation file %s failed, %s", output, err.Error())
		return output, err
	}
	defer writer.Close()
	gen := ngx.dataGeneration()
	if err := t.Execute(writer, gen); err != nil {
		klog.Errorf("proxy nginx generate detail configuration %s failed, %s", output, err.Error())
		return output, err
	}
	return output, nil
}

func (ngx *Nginx) dataGeneration() *generator {
	gen := &generator{}
	for _, tcp := range ngx.tcpKeyConfigs {
		gen.TCPServer = append(gen.TCPServer, tcp)
	}
	sort.Sort(gen.TCPServer)
	return gen
}

func (ngx *Nginx) configValidation(filename string) error {
	command := fmt.Sprintf("%s -t -c %s", ngx.option.ProxyExecutable, filename)
	// NOCC:gas/subprocess(设计如此)
	cmd := exec.Command("/bin/sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("proxy nginx check config %s failed, err: %s. command output: %s", filename, err.Error(), string(output))
		return err
	}
	klog.V(3).Infof("proxy nginx validate config %s successfully, output: %s", filename, string(output))
	return nil
}

func (ngx *Nginx) isConfigChanged(filename string) (bool, error) {
	// new file
	newmd5, err := md5sum(filename)
	if err != nil {
		return false, err
	}
	oldmd5, err := md5sum(ngx.option.ProxyConfig)
	if err != nil {
		return false, err
	}
	if newmd5 == oldmd5 {
		return false, nil
	}
	return true, nil
}

func md5sum(filename string) (string, error) {
	config, err := os.Open(filename)
	if err != nil {
		klog.Errorf("Open file %s failed: %s", filename, err.Error())
		return "", fmt.Errorf("Open file %s failed: %s", filename, err.Error())
	}
	defer config.Close()
	// NOCC:gas/crypto(设计如此)
	// nolint
	md5Block := md5.New()
	_, err = io.Copy(md5Block, config)
	if err != nil {
		klog.Errorf("do io.Copy failed when calculate file %s md5, err %s", filename, err.Error())
		return "", fmt.Errorf("md5 %s failed %s", filename, err.Error())
	}
	md5Str := string(md5Block.Sum([]byte("")))
	return md5Str, nil
}

func (ngx *Nginx) reloadNginx(config string) error {
	// open all configuration files for replace
	src, sErr := os.Open(config)
	if sErr != nil {
		klog.Errorf("Read new config file [%s] failed: %s", config, sErr.Error())
		return sErr
	}
	defer src.Close()
	dst, dErr := os.OpenFile(ngx.option.ProxyConfig, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if dErr != nil {
		klog.Errorf("Read old config file %s failed: %s", ngx.option.ProxyConfig, dErr.Error())
		return dErr
	}
	defer dst.Close()
	// mv file
	_, err := io.Copy(dst, src)
	if err != nil {
		klog.Errorf("Copy new configurtion %s failed: %s", config, err.Error())
		return err
	}
	klog.V(3).Infof("Replace config file %s success", config)
	// ready to reload
	command := fmt.Sprintf("%s -s reload", ngx.option.ProxyExecutable)
	// NOCC:gas/subprocess(设计如此)
	cmd := exec.Command("/bin/sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("proxy nginx reload %s failed, %s. detail %s", config, err.Error(), string(output))
		return err
	}
	klog.V(3).Infof("proxy reload %s successfully, detail: %s", config, string(output))
	return nil
}

// LastError get last reload error information according to egress rule
func (ngx *Nginx) LastError(egress string) error {
	ngx.errorLock.RLock()
	defer ngx.errorLock.Unlock()
	err, ok := ngx.lastError[egress]
	if ok {
		return err
	}
	return nil
}

// fileExists check file exists
func fileExists(filename string) (bool, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
