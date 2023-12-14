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

package app

import (
	"errors"
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
)

// Setting repo mock related setting.
type Setting struct {
	Network   Network      `yaml:"network"`
	Workspace Workspace    `yaml:"workspace"`
	Log       cc.LogOption `yaml:"log"`
}

// Workspace repo workspace related setting.
type Workspace struct {
	// RootDirectory absolute workspace root directory. it stores the repoMock's runtime related
	//  logs, files and metadata information.
	RootDirectory string `yaml:"rootDirectory"`
}

// Network defines all the network related options
type Network struct {
	// BindIP is ip where server working on
	BindIP string `yaml:"bindIP"`
	// Port is port where server listen to http port.
	Port uint      `yaml:"port"`
	TLS  TLSConfig `yaml:"tls"`
}

func (n Network) validate() error {

	if len(n.BindIP) == 0 {
		return errors.New("network bindIP is not set")
	}

	if ip := net.ParseIP(n.BindIP); ip == nil {
		return errors.New("invalid network bindIP")
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

// validate workspace setting.
func (w Workspace) validate() error {
	if len(w.RootDirectory) == 0 {
		return errors.New("root directory is not set")
	}

	return nil
}

// trySetFlagBindIP try set flag bind ip, bindIP only can set by one of the flag or configuration file.
func (s *Setting) trySetFlagBindIP(ip net.IP) error {
	if len(ip) != 0 {
		if len(s.Network.BindIP) != 0 {
			return errors.New("bind ip only can set by one of the flags or configuration file")
		}

		s.Network.BindIP = ip.String()
		return nil
	}

	return nil
}

// Validate repo mock setting.
func (s *Setting) Validate() error {
	if err := s.Network.validate(); err != nil {
		return err
	}

	if err := s.Workspace.validate(); err != nil {
		return err
	}

	return nil
}

func (s *Setting) trySetDefault() {
	if len(s.Network.BindIP) == 0 {
		s.Network.BindIP = "127.0.0.1"
	}

	if len(s.Log.LogDir) == 0 {
		s.Log.LogDir = "./log"
	}

	if s.Log.MaxPerFileSizeMB == 0 {
		s.Log.MaxPerFileSizeMB = 1024
	}

	if s.Log.MaxPerLineSizeKB == 0 {
		s.Log.MaxPerLineSizeKB = 10
	}

	if s.Log.MaxFileNum == 0 {
		s.Log.MaxFileNum = 5
	}
}

// LoadSettings load service's configuration
func LoadSettings(sys *cc.SysOption) (*Setting, error) {
	if len(sys.ConfigFiles) == 0 {
		return nil, errors.New("service's configuration file path is not configured")
	}

	// configure file is configured, then load configuration from file.
	file, err := os.ReadFile(sys.ConfigFiles[0])
	if err != nil {
		return nil, fmt.Errorf("load setting from file: %s failed, err: %v", sys.ConfigFiles[0], err)
	}

	s := new(Setting)
	if err := yaml.Unmarshal(file, s); err != nil {
		return nil, fmt.Errorf("unmarshal setting yaml from file: %s failed, err: %v", sys.ConfigFiles[0], err)
	}

	if err = s.trySetFlagBindIP(sys.BindIP); err != nil {
		return nil, err
	}

	// set the default value if user not configured.
	s.trySetDefault()

	if err := s.Validate(); err != nil {
		return nil, err
	}

	return s, nil
}
