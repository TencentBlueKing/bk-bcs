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

package option

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"unicode"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

const (
	// ConfigSavePath is config file save path
	ConfigSavePath = "./.bscp"
	// ConfigInfoSaveName is save default param (business app operator token)
	ConfigInfoSaveName = "desc"
	// ScanAreaSaveName save add to the scan area
	ScanAreaSaveName = "record"

	// LocalDirInitCommand use to init local repo
	LocalDirInitCommand = "init"

	// ParamEnvValPrefix
	ParamEnvValPrefix = "BSCP_"
)

var (
	//GlobalOptions setting global options for all sub command
	GlobalOptions *Global
)

func init() {
	GlobalOptions = &Global{
		ConfigFile: "/etc/bscp/client.yaml",
		Business:   "",
		Operator:   "",
		Token:      "",
		Index:      0,
		Limit:      100,
	}
}

//Global all options shared in all sub commands
type Global struct {
	//ConfigFile for client connect to platform
	ConfigFile string
	//Business name for operation
	Business string
	//Operator user name for operation
	Operator string
	//Operator user token for operation
	Token string
	//Index for list
	Index int32
	//Limit for list
	Limit int32
}

type CurrentDirConf struct {
	App      string `yaml:"app"`
	Business string `yaml:"business"`
	Operator string `yaml:"operator"`
	Token    string `yaml:"token"`
}

// getConf get ConfigSavePath/ConfigInfoSaveName file content to CurrentDirConf struct
func (c *CurrentDirConf) getConf() (*CurrentDirConf, error) {
	yamlFile, err := ioutil.ReadFile(path.Clean(ConfigSavePath + "/" + ConfigInfoSaveName))
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, err
	}

	return c, err
}

// GetInitConfInfo get ConfigSavePath/ConfigInfoSaveName info
func GetInitConfInfo() (*CurrentDirConf, error) {
	var c CurrentDirConf
	conf, err := c.getConf()
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// ParseGlobalOption parse global option for all commands
func ParseGlobalOption(cmd *cobra.Command, args []string) error {
	// TODO() check input command flag format (to solve cobra package flags: -name == -n ame)

	// judge subcommand is admin command,if yes, only need operator and token
	if len(os.Args) >= 3 {
		if isAdminCMD(os.Args[2]) {
			err := SetGlobalAdminVarByName(cmd, "operator")
			if err != nil {
				return err
			}

			err = SetGlobalAdminVarByName(cmd, "token")
			if err != nil {
				return err
			}
			return nil
		}
	}

	// reinit judege
	if os.Args[1] == LocalDirInitCommand && isWorkSpaceInited() {
		basePath, _ := os.Getwd()
		return fmt.Errorf("reinitialized existing bscp operation director in %s/.bscp/", basePath)
	}
	// judge in operating the catalog correctly
	if os.Args[1] != LocalDirInitCommand && !isWorkSpaceInited() {
		return fmt.Errorf("not the operation directory where the bscp client executes commands, please switch to the correct directory, or initialize the directory")
	}

	// judge subcommand is local command, if yes, do not get globalVar
	if issubCommandLocalCMD(os.Args[1]) {
		return nil
	}

	err := SetGlobalVarByName(cmd, "operator")
	if err != nil {
		return err
	}

	err = SetGlobalVarByName(cmd, "token")
	if err != nil {
		return err
	}

	err = SetGlobalVarByName(cmd, "business")
	if err != nil {
		return err
	}
	return nil
}

// SetGlobalVarByName get global var from command -> env -> ConfigSavePath/ConfigInfoSaveName
func SetGlobalVarByName(cmd *cobra.Command, name string) error {
	valueFlag, _ := cmd.Flags().GetString(name)
	if len(valueFlag) == 0 {
		//business enviroment
		envName := ParamEnvValPrefix + strings.ToUpper(name)
		valueEnv := os.Getenv(envName)
		if len(valueEnv) != 0 {
			cmd.Flags().Set(name, valueEnv)
		} else {
			var c CurrentDirConf
			conf, err := c.getConf()
			if err != nil {
				return err
			}
			confName := string(unicode.ToUpper(rune(name[0]))) + name[1:]
			rValue := reflect.ValueOf(*conf)
			valueConf := rValue.FieldByName(confName).String()
			if len(valueConf) != 0 {
				cmd.Flags().Set(name, valueConf)
			} else {
				return fmt.Errorf("command line parameters, environment variables, %s file no %s found\n", path.Clean(ConfigSavePath+"/"+ConfigInfoSaveName), name)
			}
		}
	}
	return nil
}

// SetGlobalADMINVarByName get global admin var from command -> env -> ConfigSavePath/ConfigInfoSaveName
func SetGlobalAdminVarByName(cmd *cobra.Command, name string) error {
	valueFlag, _ := cmd.Flags().GetString(name)
	if len(valueFlag) == 0 {
		//business enviroment
		envName := ParamEnvValPrefix + strings.ToUpper(name)
		valueEnv := os.Getenv(envName)
		if len(valueEnv) != 0 {
			cmd.Flags().Set(name, valueEnv)
		} else {
			var c CurrentDirConf
			conf, err := c.getConf()
			if err != nil {
				return fmt.Errorf("admin command need you enter %s, not available from any channels now\n", name)
			}
			confName := string(unicode.ToUpper(rune(name[0]))) + name[1:]
			rValue := reflect.ValueOf(*conf)
			valueConf := rValue.FieldByName(confName).String()
			if len(valueConf) != 0 {
				cmd.Flags().Set(name, valueConf)
			} else {
				return fmt.Errorf("admin command need you enter %s, not available from any channels now\n", name)
			}
		}
	}
	return nil
}

// issubCommandLocalCMD judge command is local subcommand
func issubCommandLocalCMD(cmd string) bool {
	// local subcommand
	localCommands := []string{"init", "checkout", "co", "status", "st", "config", "info", "add"}
	for _, command := range localCommands {
		if cmd == command {
			return true
		}
	}
	return false
}

// isAdminCMD judge command is admin cmd , if yes only need token and operator
func isAdminCMD(cmd string) bool {
	// admin subcommand
	adminCommands := []string{"business", "bus", "shardingdb", "db", "sharding", "sd"}
	for _, command := range adminCommands {
		if cmd == command {
			return true
		}
	}
	return false
}

// isWorkSpaceInited judge workspace is inited
func isWorkSpaceInited() bool {
	_, err := os.Stat(ConfigSavePath)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
