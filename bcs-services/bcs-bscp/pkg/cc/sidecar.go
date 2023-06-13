/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package cc

import (
	"errors"
	"fmt"
	"net"

	"bscp.io/pkg/criteria/validator"
)

// SidecarSetting defines sidecar's related options.
type SidecarSetting struct {
	// Version is the sidecar's configuration file's version.
	Version   string           `yaml:"version"`
	Network   SidecarNetwork   `yaml:"network"`
	Upstream  SidecarUpstream  `yaml:"upstream"`
	Workspace SidecarWorkspace `yaml:"workspace"`
	AppSpec   SidecarAppSpec   `yaml:"appSpec"`
	Log       LogOption        `yaml:"log"`
}

func (s *SidecarSetting) trySetFlagBindIP(ip net.IP) error {
	return s.Network.trySetFlagBindIP(ip)
}

// trySetFlagPort set http and grpc port
func (s *SidecarSetting) trySetFlagPort(port, grpcPort int) error {
	return nil
}

func (s *SidecarSetting) trySetDefault() {
	s.Network.trySetDefault()
	s.Log.trySetDefault()

	if len(s.Log.LogDir) == 0 {
		// set the sidecar's log directory under its root directory as default.
		s.Log.LogDir = s.Workspace.RootDirectory + "logs/"
	}

}

// Validate the sidecar's settings is valid or not.
func (s *SidecarSetting) Validate() error {
	if err := s.Network.validate(); err != nil {
		return fmt.Errorf("invalid network, err: %v", err)
	}

	if err := s.Upstream.validate(); err != nil {
		return fmt.Errorf("invalid upstream, err: %v", err)
	}

	if err := s.Workspace.Validate(); err != nil {
		return fmt.Errorf("invalid workspace, err: %v", err)
	}

	if err := s.AppSpec.Validate(); err != nil {
		return fmt.Errorf("invalid appSpec, err: %v", err)
	}

	return nil
}

// SidecarNetwork define sidecar network
type SidecarNetwork struct {
	// BindIP is ip where server working on
	BindIP string `yaml:"bindIP"`
	// HttpPort is port where server listen to http port.
	HttpPort uint `yaml:"httpPort"`
	// ShutdownTimeoutSec is the max time in second for the sidecar to shutdown gracefully when it receive a
	// shutdown signal. if the shutdown timeout time reaches after the shutdown process starts, sidecar will
	// be forced to exit no matter the shutdown jobs has been finished or not. its min value is 5s and the
	// default value is 20s.
	ShutdownTimeoutSec int       `yaml:"shutdownTimeoutSec"`
	TLS                TLSConfig `yaml:"tls"`
}

func (sn *SidecarNetwork) trySetFlagBindIP(ip net.IP) error {
	if len(ip) != 0 {
		if len(sn.BindIP) != 0 {
			return errors.New("bind ip only can set by one of the flags or configuration file")
		}

		sn.BindIP = ip.String()
		return nil
	}

	return nil
}

// trySetDefault set the network's default value if user not configured.
func (sn *SidecarNetwork) trySetDefault() {
	if len(sn.BindIP) == 0 {
		sn.BindIP = "127.0.0.1"
	}

	if sn.ShutdownTimeoutSec == 0 {
		sn.ShutdownTimeoutSec = 20
	}
}

// validate network options
func (sn SidecarNetwork) validate() error {

	if len(sn.BindIP) == 0 {
		return errors.New("network bindIP is not set")
	}

	if ip := net.ParseIP(sn.BindIP); ip == nil {
		return errors.New("invalid network bindIP")
	}

	if sn.ShutdownTimeoutSec < 5 && sn.ShutdownTimeoutSec != 0 {
		return errors.New("invalid network shutdownTimeoutSec, should >= 5")
	}

	if err := sn.TLS.validate(); err != nil {
		return fmt.Errorf("network tls, %v", err)
	}

	return nil
}

// SidecarUpstream defines the sidecar's upstream connection related options.
type SidecarUpstream struct {
	// Endpoints are a list of addresses of the feed server's endpoints with ip:port.
	// Usually, it's recommended to use DNS.
	Endpoints      []string              `yaml:"endpoints"`
	Authentication SidecarAuthentication `yaml:"authentication"`
	// DialTimeoutMS is the timeout milliseconds for failing to establish the grpc connection.
	// if = 0, it means dial with no timeout, if > 0, then it should range between [50,15000]
	DialTimeoutMS uint      `yaml:"dialTimeoutMS"`
	TLS           TLSConfig `yaml:"tls"`
}

// trySetDefault set the network's default value if user not configured.
func (sn *SidecarUpstream) trySetDefault() {
	if len(sn.Endpoints) == 0 {
		sn.Endpoints = []string{"127.0.0.1:9510"}
	}

	if sn.DialTimeoutMS == 0 {
		sn.DialTimeoutMS = 2000
	}

}

// validate network options
func (sn SidecarUpstream) validate() error {
	if len(sn.Endpoints) == 0 {
		return errors.New("upstream endpoints is not set")
	}

	for _, one := range sn.Endpoints {
		if len(one) == 0 {
			return errors.New("invalid endpoints which is empty")
		}
	}

	if sn.DialTimeoutMS > 0 {
		if sn.DialTimeoutMS < 0 || sn.DialTimeoutMS > 15000 {
			return errors.New("upstream.dialTimeoutMS should range between [50, 15000]")
		}
	}

	return nil
}

// SidecarAuthentication defines sidecar's authentication information.
type SidecarAuthentication struct {
	User  string `yaml:"user"`
	Token string `yaml:"token"`
}

// IsEnabled check if the sidecar's authentication is enabled.
func (sa SidecarAuthentication) IsEnabled() bool {
	if len(sa.User) == 0 && len(sa.Token) == 0 {
		return false
	}

	return true
}

// SidecarAppSpec defines the sidecar managed apps specifics.
type SidecarAppSpec struct {
	BizID        uint32        `yaml:"bizID"`
	Applications []AppMetadata `yaml:"applications"`
}

// Validate the sidecar's app spec is valid or not
func (as SidecarAppSpec) Validate() error {
	if as.BizID <= 0 {
		return errors.New("invalid biz id")
	}

	if len(as.Applications) == 0 {
		return errors.New("invalid appSpec.application, at least one application is needed")
	}

	if len(as.Applications) > validator.MaxAppMetas {
		return fmt.Errorf("at most %d applications is allowed in the appSpec", validator.MaxAppMetas)
	}

	for _, one := range as.Applications {
		if err := one.Validate(); err != nil {
			return fmt.Errorf("invalidate application information for app: %s, err: %v", one.App, err)
		}
	}

	return nil
}

// AppMetadata defines the app's metadata managed by sidecar.
type AppMetadata struct {
	App       string            `json:"app" yaml:"app"`
	Namespace string            `json:"namespace" yaml:"namespace"`
	Uid       string            `json:"uid" yaml:"uid"`
	Labels    map[string]string `json:"labels" yaml:"labels"`
}

// Validate the app metadata is valid or not
func (am AppMetadata) Validate() error {
	if am.App == "" {
		return errors.New("invalid app")
	}

	if err := validator.ValidateUid(am.Uid); err != nil {
		return fmt.Errorf("invalid app: %s uid: %s, err: %v", am.App, am.Uid, err)
	}

	return nil
}

// SidecarWorkspace defines sidecar's workspace options.
type SidecarWorkspace struct {
	// sidecar's absolute workspace root directory. this directory can only
	// be used by sidecar, business user must not use this directory.
	RootDirectory string              `yaml:"rootDirectory"`
	PurgePolicy   *SidecarPurgePolicy `json:"purgePolicy"`
}

// Validate the sidecar workspace is valid or not
func (sw SidecarWorkspace) Validate() error {
	if len(sw.RootDirectory) == 0 {
		return errors.New("sidecar's workspace root directory is empty, not allowed")
	}

	if sw.PurgePolicy != nil {
		if err := sw.PurgePolicy.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// SidecarPurgePolicy defines the sidecar's workspace purge policy
type SidecarPurgePolicy struct {
	// sidecar will do auto clean user's temporary config files if enabled
	// the auto clean policy if it is possible. Only unused config item
	// files will be removed when the workspace size is over than maxSizeMB.
	EnableAutoClean bool `yaml:"enableAutoClean"`

	// the max size of sidecar's workspace size. used when EnableAutoClean
	// is true, and it's value should also be set with a reasonable value.
	MaxSizeMB uint `yaml:"maxSizeMB"`

	// the minute interval to auto clean unused config files when EnableAutoClean
	// is true. it should be larger than 60.
	AutoCleanIntervalMin uint `yaml:"autoCleanIntervalMin"`
}

// Validate the validation of SidecarPurgePolicy
func (sp SidecarPurgePolicy) Validate() error {
	if sp.EnableAutoClean {
		if sp.MaxSizeMB == 0 {
			return errors.New("workspace.sidecarPurgePolicy is true, but workspace.maxSizeMB not set")
		}

		if sp.AutoCleanIntervalMin > 0 && sp.AutoCleanIntervalMin < 60 {
			return errors.New("workspace.sidecarPurgePolicy should >= 60")
		}
	}

	return nil
}
