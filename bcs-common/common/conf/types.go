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

package conf

import (
	goflag "flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/bitly/go-simplejson"
	"github.com/spf13/pflag"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
)

// FileConfig Config file, if set it will cover all the flag value it contains
type FileConfig struct {
	ConfigFile string `json:"file" short:"f" value:"" usage:"json file with configuration"`
}

// LogConfig Log configuration
// nolint
type LogConfig struct {
	LogDir     string `json:"log_dir" value:"./logs" usage:"If non-empty, write log files in this directory" mapstructure:"log_dir" yaml:"log_dir"`
	LogMaxSize uint64 `json:"log_max_size" value:"500" usage:"Max size (MB) per log file." mapstructure:"log_max_size" yaml:"log_max_size"`
	LogMaxNum  int    `json:"log_max_num" value:"10" usage:"Max num of log file. The oldest will be removed if there is a extra file created." mapstructure:"log_max_num" yaml:"log_max_num"`

	ToStdErr        bool   `json:"logtostderr" value:"false" usage:"log to standard error instead of files" mapstructure:"logtostderr" yaml:"logtostderr"`
	AlsoToStdErr    bool   `json:"alsologtostderr" value:"false" usage:"log to standard error as well as files" mapstructure:"alsologtostderr" yaml:"alsologtostderr"`
	Verbosity       int32  `json:"v" value:"0" usage:"log level for V logs" mapstructure:"v" yaml:"v"`
	StdErrThreshold string `json:"stderrthreshold" value:"2" usage:"logs at or above this threshold go to stderr" mapstructure:"stderrthreshold" yaml:"stderrthreshold"`
	VModule         string `json:"vmodule" value:"" usage:"comma-separated list of pattern=N settings for file-filtered logging" mapstructure:"vmodule" yaml:"vmodule"`
	TraceLocation   string `json:"log_backtrace_at" value:"" usage:"when logging hits line file:N, emit a stack trace" mapstructure:"log_backtrace_at" yaml:"log_backtrace_at"`
}

// ProcessConfig Process configuration
type ProcessConfig struct {
	PidDir string `json:"pid_dir" value:"./pid" usage:"The dir for pid file" mapstructure:"pid_dir"`
}

// ServiceConfig Service bind
type ServiceConfig struct {
	Address         string `json:"address" short:"a" value:"127.0.0.1" usage:"IP address to listen on for this service" mapstructure:"address"`         // nolint
	IPv6Address     string `json:"ipv6_address" value:"" usage:"IPv6 address to listen on for this service" mapstructure:"ipv6_address"`                // nolint
	Port            uint   `json:"port" short:"p" value:"8080" usage:"Port to listen on for this service" mapstructure:"port"`                          // nolint
	InsecureAddress string `json:"insecure_address" value:"" usage:"insecure IP address to listen on for this service" mapstructure:"insecure_address"` // nolint
	InsecurePort    uint   `json:"insecure_port" value:"" usage:"insecure port to listen on for this service" mapstructure:"insecure_port"`             // nolint
	ExternalIp      string `json:"external_ip" value:"" usage:"external IP address to listen on for this service" mapstructure:"external_ip"`           // nolint
	ExternalIPv6    string `json:"external_ipv6" value:"" usage:"external IPv6 address to listen on for this service" mapstructure:"external_ipv6"`     // nolint
	ExternalPort    uint   `json:"external_port" value:"" usage:"external port to listen on for this service" mapstructure:"external_port"`             // nolint
}

// LocalConfig Local info
type LocalConfig struct {
	LocalIP   string `json:"local_ip" value:"" usage:"IP address of this host" mapstructure:"local_ip"`
	LocalIPv6 string `json:"local_ipv6" value:"" usage:"IPv6 address of this host" mapstructure:"local_ipv6"`
}

// MetricConfig Metric info
type MetricConfig struct {
	MetricPort uint `json:"metric_port" value:"8081" usage:"Port to listen on for metric" mapstructure:"metric_port" `
}

// ZkConfig bcs zookeeper for service discovery
type ZkConfig struct {
	BCSZk     string `json:"bcs_zookeeper" value:"127.0.0.1:2181" usage:"Zookeeper server for registering and discovering" mapstructure:"bcs_zookeeper" `          // nolint
	BCSZKIPv6 string `json:"bcs_zookeeper_ipv6" value:"" usage:"Zookeeper server ipv6 address for registering and discovering" mapstructure:"bcs_zookeeper_ipv6" ` // nolint
}

// CertConfig Server and client TLS config, can not be import with ClientCertOnlyConfig or ServerCertOnlyConfig
type CertConfig struct {
	CAFile         string `json:"ca_file" value:"" usage:"CA file. If server_cert_file/server_key_file/ca_file are all set, it will set up an HTTPS server required and verified client cert" mapstructure:"ca_file"`    // nolint
	ServerCertFile string `json:"server_cert_file" value:"" usage:"Server public key file(*.crt). If both server_cert_file and server_key_file are set, it will set up an HTTPS server" mapstructure:"server_cert_file"` // nolint
	ServerKeyFile  string `json:"server_key_file" value:"" usage:"Server private key file(*.key). If both server_cert_file and server_key_file are set, it will set up an HTTPS server" mapstructure:"server_key_file"`  // nolint
	ClientCertFile string `json:"client_cert_file" value:"" usage:"Client public key file(*.crt)" mapstructure:"client_cert_file"`                                                                                       // nolint
	ClientKeyFile  string `json:"client_key_file" value:"" usage:"Client private key file(*.key)" mapstructure:"client_key_file"`                                                                                        // nolint
}

// ClientOnlyCertConfig Client TLS config, can not be import with CertConfig or ServerCertOnlyConfig
type ClientOnlyCertConfig struct {
	CAFile         string `json:"ca_file" value:"" usage:"CA file. If server_cert_file/server_key_file/ca_file are all set, it will set up an HTTPS server required and verified client cert"` // nolint
	ClientCertFile string `json:"client_cert_file" value:"" usage:"Client public key file(*.crt)"`
	ClientKeyFile  string `json:"client_key_file" value:"" usage:"Client private key file(*.key)"`
}

// ServerOnlyCertConfig Server TLS config, can not be import with ClientCertOnlyConfig or CertConfig
type ServerOnlyCertConfig struct {
	CAFile         string `json:"ca_file" value:"" usage:"CA file. If server_cert_file/server_key_file/ca_file are all set, it will set up an HTTPS server required and verified client cert"` // nolint
	ServerCertFile string `json:"server_cert_file" value:"" usage:"Server public key file(*.crt). If both server_cert_file and server_key_file are set, it will set up an HTTPS server"`       // nolint
	ServerKeyFile  string `json:"server_key_file" value:"" usage:"Server private key file(*.key). If both server_cert_file and server_key_file are set, it will set up an HTTPS server"`       // nolint
}

// LicenseServerConfig License server config
type LicenseServerConfig struct {
	LSAddress        string `json:"ls_address" value:"" usage:"The license server address" mapstructure:"ls_address"`
	LSCAFile         string `json:"ls_ca_file" value:"" usage:"CA file for connecting to license server" mapstructure:"ls_ca_file"`                                         // nolint
	LSClientCertFile string `json:"ls_client_cert_file" value:"" usage:"Client public key file(*.crt) for connecting to license server" mapstructure:"ls_client_cert_file"` // nolint
	LSClientKeyFile  string `json:"ls_client_key_file" value:"" usage:"Client private key file(*.key) for connecting to license server" mapstructure:"ls_client_key_file"`  // nolint
}

// CustomCertConfig xxx
type CustomCertConfig struct {
	CAFile         string `json:"custom_ca_file" value:"" usage:"Custom CA file. If server_cert_file/server_key_file/ca_file are all set, it will set up an HTTPS server required and verified client cert" mapstructure:"custom_ca_file"`    // nolint
	ServerCertFile string `json:"custom_server_cert_file" value:"" usage:"Custom Server public key file(*.crt). If both server_cert_file and server_key_file are set, it will set up an HTTPS server" mapstructure:"custom_server_cert_file"` // nolint
	ServerKeyFile  string `json:"custom_server_key_file" value:"" usage:"Custom Server private key file(*.key). If both server_cert_file and server_key_file are set, it will set up an HTTPS server" mapstructure:"custom_server_key_file"`  // nolint
	ServerKeyPwd   string `json:"custom_server_key_password" value:"" usage:"specific the custom server tls key file pwd" mapstructure:"custom_server_key_password"`                                                                          // nolint
	ClientCertFile string `json:"custom_client_cert_file" value:"" usage:"Custom Client public key file(*.crt)" mapstructure:"client_cert_file" mapstructure:"custom_client_cert_file"`                                                       // nolint
	ClientKeyFile  string `json:"custom_client_key_file" value:"" usage:"Custom Client private key file(*.key)" mapstructure:"client_key_file" mapstructure:"custom_client_key_file"`                                                         // nolint
	ClientKeyPwd   string `json:"custom_client_key_password" value:"" usage:"specific the custom client tls key file pwd" mapstructure:"custom_client_key_password"`                                                                          // nolint
}

// Parse parse all config item
func Parse(config interface{}) {
	// load config to flag
	loadRawConfig(pflag.CommandLine, config)

	// parse flags
	util.InitFlags()
	if err := goflag.CommandLine.Parse([]string{}); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// parse config file if exists
	if err := decJSON(pflag.CommandLine, config); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func decJSON(fs *pflag.FlagSet, config interface{}) error {
	f := reflect.ValueOf(config).Elem().FieldByName("ConfigFile").String()
	if f == "" {
		return nil
	}
	raw, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	jsn, err := simplejson.NewJson(raw)
	if err != nil {
		return err
	}

	removeLowPriorityKey(fs, jsn, reflect.TypeOf(config).Elem())

	safeRaw, err := jsn.MarshalJSON()
	if err != nil {
		return err
	}

	return codec.DecJson(safeRaw, config)
}

func removeLowPriorityKey(fs *pflag.FlagSet, jsn *simplejson.Json, flagConfigType reflect.Type) {
	flagNames := make([]string, 0)
	n := flagConfigType.NumField()
	for i := 0; i < n; i++ {
		field := flagConfigType.Field(i)
		jsonName := field.Tag.Get("json")

		if jsonName == "" {
			switch field.Type.Kind() {
			case reflect.Struct:
				removeLowPriorityKey(fs, jsn, field.Type)
			case reflect.Ptr:
				removeLowPriorityKey(fs, jsn, field.Type.Elem())
			}
			continue
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			removeLowPriorityKey(fs, jsn.Get(jsonName), field.Type)
			continue
		case reflect.Ptr:
			removeLowPriorityKey(fs, jsn.Get(jsonName), field.Type.Elem())
		default:
			flagNames = append(flagNames, jsonName)
		}
	}

	for _, fn := range flagNames {
		if fs.Changed(fn) {
			jsn.Del(fn)
		}
	}
}

// loadRawConfig xxx
// Make field to flag by adding "json" "value" "usage"
func loadRawConfig(fs *pflag.FlagSet, config interface{}) {
	wrap2flag(fs, reflect.TypeOf(config).Elem(), reflect.ValueOf(config).Elem())
}

func loadConfig(fs *pflag.FlagSet, configType reflect.Type, configValue reflect.Value) {
	wrap2flag(fs, configType, configValue)
}

func wrap2flag(fs *pflag.FlagSet, configType reflect.Type, configValue reflect.Value) {
	n := configType.NumField()

	for i := 0; i < n; i++ {
		field := configType.Field(i)
		fieldV := configValue.Field(i)
		if !fieldV.IsValid() || !fieldV.CanSet() {
			continue
		}

		flagName, nameOk := field.Tag.Lookup("json")
		if !nameOk && !field.Anonymous {
			continue
		}

		_, valueOk := field.Tag.Lookup("value")
		flagUsage, usageOk := field.Tag.Lookup("usage")

		switch field.Type.Kind() {
		case reflect.Struct:
			loadConfig(fs, field.Type, fieldV)
			continue
		case reflect.Ptr:
			loadConfig(fs, field.Type.Elem(), fieldV.Elem())
			continue
		}

		// flag must have "json, value, usage" flags
		// json and flag can not be empty
		if !nameOk || !valueOk || !usageOk || flagName == "" || flagUsage == "" {
			continue
		}

		wrapFieldFlag(fs, field, fieldV)
	}
}

func wrapFieldFlag(fs *pflag.FlagSet, field reflect.StructField, fieldV reflect.Value) {
	flagName := field.Tag.Get("json")
	flagValue := field.Tag.Get("value")
	flagUsage := field.Tag.Get("usage")
	flagShortHand := field.Tag.Get("short")

	unsafePtr := unsafe.Pointer(fieldV.UnsafeAddr()) // nolint
	switch field.Type.Kind() {
	case reflect.String:
		fs.StringVarP((*string)(unsafePtr), flagName, flagShortHand, flagValue, flagUsage)
	case reflect.Bool:
		v := flagValue == "true"
		fs.BoolVarP((*bool)(unsafePtr), flagName, flagShortHand, v, flagUsage)
	case reflect.Uint:
		v, _ := strconv.ParseUint(flagValue, 10, 0)
		fs.UintVarP((*uint)(unsafePtr), flagName, flagShortHand, uint(v), flagUsage)
	case reflect.Uint32:
		v, _ := strconv.ParseUint(flagValue, 10, 32)
		fs.Uint32VarP((*uint32)(unsafePtr), flagName, flagShortHand, uint32(v), flagUsage)
	case reflect.Uint64:
		v, _ := strconv.ParseUint(flagValue, 10, 64)
		fs.Uint64VarP((*uint64)(unsafePtr), flagName, flagShortHand, v, flagUsage)
	case reflect.Int:
		v, _ := strconv.ParseInt(flagValue, 10, 0)
		fs.IntVarP((*int)(unsafePtr), flagName, flagShortHand, int(v), flagUsage)
	case reflect.Int32:
		v, _ := strconv.ParseInt(flagValue, 10, 32)
		fs.Int32VarP((*int32)(unsafePtr), flagName, flagShortHand, int32(v), flagUsage)
	case reflect.Int64:
		v, _ := strconv.ParseInt(flagValue, 10, 64)
		fs.Int64VarP((*int64)(unsafePtr), flagName, flagShortHand, v, flagUsage)
	case reflect.Float32:
		v, _ := strconv.ParseFloat(flagValue, 32)
		fs.Float32VarP((*float32)(unsafePtr), flagName, flagShortHand, float32(v), flagUsage)
	case reflect.Float64:
		v, _ := strconv.ParseFloat(flagValue, 64)
		fs.Float64VarP((*float64)(unsafePtr), flagName, flagShortHand, v, flagUsage)
	case reflect.Slice:
		arr := make([]string, 0)
		if flagValue != "" {
			arr = strings.Split(flagValue, ",")
		}
		switch field.Type.Elem().Kind() {
		case reflect.String:
			fs.StringSliceVarP((*[]string)(unsafePtr), flagName, flagShortHand, arr, flagUsage)
		case reflect.Int:
			intArr := make([]int, 0, len(arr))
			for _, si := range arr {
				ii, _ := strconv.ParseInt(si, 10, 0)
				intArr = append(intArr, int(ii))
			}
			fs.IntSliceVarP((*[]int)(unsafePtr), flagName, flagShortHand, intArr, flagUsage)
		}
	}
}

// InitIPv6AddressFiled 初始化 ServiceConfig 的 IPv6Address 字段
// 1.检查当前字段 IPv6Address 是否为合法IPv6，若是合法IPv6，则结束执行；否则，执行下一步。
// 2.依次遍历当前字段 IPv6Address、“localIpv6”环境变量，检查是否存在"IPv4,IPv6"地址表示法，
// 并检查IPv6地址合法性，若，存在并合法，则把新的IPv6地址，赋值给 IPv6Address字段，并结束执行 ；否则，执行下一步。
// 3.设置 IPv6Address 字段为默认值 "::1"
func (sc *ServiceConfig) InitIPv6AddressFiled() {
	sc.IPv6Address = util.InitIPv6Address(sc.IPv6Address)
}
