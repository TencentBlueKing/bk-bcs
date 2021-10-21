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

package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/types"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/urfave/cli"
)

// BcsEnv stores the env variables
type BcsEnv struct {
	ClusterID string `json:"clusterid,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

var env BcsEnv

var envPath = "/var/bcs/bcsenv.conf"

//InitEnv reading pre-Store ENV in client specified ENV file
func InitEnv() error {
	file, err := ioutil.ReadFile(envPath)

	if err != nil {
		return err
	}

	if errMarsh := codec.DecJson(file, &env); errMarsh != nil {
		return fmt.Errorf("failed to parse %s. decode error: %v", string(file), err)
	}

	return nil
}

//ShowEnv for client release version
func ShowEnv() {
	fmt.Printf("CLUSTERID=%s\n", env.ClusterID)
	fmt.Printf("NAMESPACE=%s\n", env.Namespace)
}

//SetEnv store clusterID & namespace in client cache file
func SetEnv(clusterID, namespace string) error {
	env.ClusterID = clusterID
	env.Namespace = namespace

	var file *os.File
	var err error
	file, err = os.OpenFile(envPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	defer func() {
		_ = file.Close()
	}()

	if err != nil {
		return fmt.Errorf("set env error, open file error: %v", err)
	}

	var bEnv []byte
	_ = codec.EncJson(env, &bEnv)

	_, err = io.WriteString(file, string(bEnv))
	if err != nil {
		return fmt.Errorf("set env error, write file error: %v", err)
	}

	return nil
}

// BcsCfg stores the config settings of client
type BcsCfg struct {
	ApiHost     string `json:"apiserver,omitempty"`
	EnableDebug bool   `json:"debug,omitempty"`
	BcsToken    string `json:"bcs_token,omitempty"`
	CAFile      string `json:"ca_file,omitempty"`
	CertFile    string `json:"cert_file,omitempty"`
	KeyFile     string `json:"key_file,omitempty"`

	CustomCAFile   string `json:"custom_ca_file,omitempty"`
	CustomCertFile string `json:"custom_cert_file,omitempty"`
	CustomKeyFile  string `json:"custom_key_file,omitempty"`
	CustomKeyPwd   string `json:"custom_key_password,omitempty"`

	clientSSL *tls.Config
}

var cfg BcsCfg

//InitCfg init configuration before clent run
func InitCfg() error {
	filePath := "/var/bcs/bcs.conf"
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err = codec.DecJson(file, &cfg); err != nil {
		return fmt.Errorf("failed to parse %s. decode error: %v", string(file), err)
	}

	keyPwd := static.ClientCertPwd
	if cfg.CustomCertFile != "" && cfg.CustomKeyFile != "" && cfg.CustomCAFile != "" {
		cfg.CAFile = cfg.CustomCAFile
		cfg.CertFile = cfg.CustomCertFile
		cfg.KeyFile = cfg.CustomKeyFile
		keyPwd = cfg.CustomKeyPwd
	}

	if cfg.CertFile != "" && cfg.KeyFile != "" && cfg.CAFile != "" {
		if cfg.clientSSL, err = ssl.ClientTslConfVerity(cfg.CAFile, cfg.CertFile, cfg.KeyFile, keyPwd); err != nil {
			return fmt.Errorf("failed to set client tls: %v", err)
		}
	}

	if !strings.Contains(cfg.ApiHost, "http") {
		if cfg.clientSSL != nil {
			cfg.ApiHost = fmt.Sprintf("https://%s", cfg.ApiHost)
		} else {
			cfg.ApiHost = fmt.Sprintf("http://%s", cfg.ApiHost)
		}
	}

	DebugPrintf("api address: %s\n", cfg.ApiHost)

	return nil
}

//GetClientOption init option for scheduler
func GetClientOption() types.ClientOptions {
	return types.ClientOptions{
		BcsApiAddress: cfg.ApiHost,
		BcsToken:      cfg.BcsToken,
		ClientSSL:     cfg.clientSSL,
	}
}

//DebugPrintln print only when debug mode
func DebugPrintln(a ...interface{}) {
	if cfg.EnableDebug {
		fmt.Println(a...)
	}
}

//DebugPrintf print only when debug mode
func DebugPrintf(format string, a ...interface{}) {
	if cfg.EnableDebug {
		fmt.Printf(format, a...)
	}
}

// ClientContext provides some methods when handling command.
type ClientContext struct {
	*cli.Context

	clusterID    string
	namespace    string
	allNamespace bool
}

//NewClientContext client context wrapper for bcs-client
func NewClientContext(c *cli.Context) *ClientContext {
	cc := &ClientContext{Context: c}
	cc.initEnv()
	return cc
}

//MustSpecified validate required command line option
func (cc *ClientContext) MustSpecified(key ...string) error {
	for _, k := range key {
		if k == OptionClusterID {
			if cc.clusterID == "" {
				return fmt.Errorf("cluster ID must be specified, options or env")
			}
			continue
		}
		if k == OptionNamespace {
			if cc.namespace == "" && !cc.allNamespace {
				return fmt.Errorf("namespace must be specified, options or env")
			}
			continue
		}

		if !cc.IsSet(k) {
			return fmt.Errorf("%s must be specified", k)
		}
	}

	return nil
}

func (cc *ClientContext) initEnv() {
	cc.clusterID = env.ClusterID
	if cc.IsSet(OptionClusterID) && cc.String(OptionClusterID) != "" {
		cc.clusterID = cc.String(OptionClusterID)
	}

	cc.namespace = env.Namespace
	if cc.IsSet(OptionNamespace) && cc.String(OptionNamespace) != "" {
		cc.namespace = cc.String(OptionNamespace)
	}

	if cc.IsSet(OptionAllNamespace) {
		cc.allNamespace = cc.Bool(OptionAllNamespace)
	}
}

//ClusterID get command line clusterid
func (cc *ClientContext) ClusterID() string {
	return cc.clusterID
}

//Namespace get command line namespace
func (cc *ClientContext) Namespace() string {
	return cc.namespace
}

//IsAllNamespace check if client get all namespaces data
func (cc *ClientContext) IsAllNamespace() bool {
	return cc.allNamespace
}

//FileData get --from-file details
func (cc *ClientContext) FileData() ([]byte, error) {
	if err := cc.MustSpecified(OptionFile); err != nil {
		return nil, err
	}

	return ioutil.ReadFile(cc.String(OptionFile))
}

//TryIndent try to indent json
func TryIndent(data interface{}) []byte {
	var bytesData []byte
	if err := codec.EncJson(data, &bytesData); err != nil {
		return []byte("data encode error")
	}
	return TryBytesIndent(bytesData)
}

//TryBytesIndent pretty print
func TryBytesIndent(data []byte) []byte {
	indented := &bytes.Buffer{}
	if err := json.Indent(indented, data, "", "  "); err == nil {
		return indented.Bytes()
	}
	return data
}

//ParseNamespaceFromJSON reading namespace from specified json file
func ParseNamespaceFromJSON(ctx []byte) (string, error) {
	js, err := simplejson.NewJson(ctx)
	if err != nil {
		return "", fmt.Errorf("decode json in file failed, err: %v", err)
	}

	jsMetaData := js.Get("metadata")
	namespace, _ := jsMetaData.Get("namespace").String()
	if namespace == "" {
		return "", fmt.Errorf("parse namespace failed or json structure error")
	}
	return namespace, nil
}

//ParseNameFromJSON reading name from specified json file
func ParseNameFromJSON(ctx []byte) (string, error) {
	js, err := simplejson.NewJson(ctx)
	if err != nil {
		return "", fmt.Errorf("decode json in file failed, err: %v", err)
	}

	jsMetaData := js.Get("metadata")
	name, _ := jsMetaData.Get("name").String()
	if name == "" {
		return "", fmt.Errorf("parse name failed or json structure error")
	}
	return name, nil
}

//ParseNamespaceNameFromJSON reading namespace & name from specified json file
func ParseNamespaceNameFromJSON(ctx []byte) (string, string, error) {
	js, err := simplejson.NewJson(ctx)
	if err != nil {
		return "", "", fmt.Errorf("decode json in file failed, err: %v", err)
	}

	jsMetaData := js.Get("metadata")
	namespace, _ := jsMetaData.Get("namespace").String()
	name, _ := jsMetaData.Get("name").String()
	if namespace == "" {
		return "", "", fmt.Errorf("parse namespace failed or json structure error")
	}
	if name == "" {
		return "", "", fmt.Errorf("parse name failed or json structure error")
	}
	return namespace, name, nil
}

//ParseAPIVersionAndKindFromJSON pase apiVersion & kind for validation
func ParseAPIVersionAndKindFromJSON(ctx []byte) (string, string, error) {
	js, err := simplejson.NewJson(ctx)
	if err != nil {
		return "", "", fmt.Errorf("decode json in file failed, err: %v", err)
	}

	version, _ := js.Get("apiVersion").String()
	kind, _ := js.Get("kind").String()
	if version == "" {
		return "", "", fmt.Errorf("parse apiVersion failed or json structure error")
	}
	if kind == "" {
		return "", "", fmt.Errorf("parse kind failed or json structure error")
	}
	return version, kind, nil
}

//GetIPList split ip string by comma
func GetIPList(ip string) []string {
	if len(ip) == 0 {
		return make([]string, 0)
	}

	return strings.Split(ip, ",")
}
