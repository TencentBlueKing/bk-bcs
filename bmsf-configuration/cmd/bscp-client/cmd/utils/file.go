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
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
)

// IsExists checks target dir/file exist or not.
func IsExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// WriteRCFileToScanArea to save recordConfigFiles struct to local (Data conversion protocol： []byte -> json -> base64)
func WriteRCFileToScanArea(recordConfigFiles map[string]service.ConfigFile) error {
	recordConfigJson, err := json.Marshal(recordConfigFiles)
	if err != nil {
		return err
	}
	recordConfigBase64 := base64.StdEncoding.EncodeToString(recordConfigJson)
	err = ioutil.WriteFile(path.Clean(option.ConfigSavePath+"/"+option.ScanAreaSaveName), []byte(recordConfigBase64), 0644)
	if err != nil {
		return err
	}
	return nil
}

// ReadRCFileFromScanArea to read to the scan area
func ReadRCFileFromScanArea() (map[string]service.ConfigFile, error) {
	// read content from record file
	recordConfigBase64, err := ioutil.ReadFile(path.Clean(option.ConfigSavePath + "/" + option.ScanAreaSaveName))
	if err != nil {
		return nil, err
	}

	// decoding and unmarshal
	var recordConfigFiles map[string]service.ConfigFile
	recordConfigBytes, err := base64.StdEncoding.DecodeString(string(recordConfigBase64))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(recordConfigBytes, &recordConfigFiles)
	if err != nil {
		return nil, err
	}

	return recordConfigFiles, nil
}

// DeleteEdFileFromRecord to deelete deleted file from record
func DeleteEdFileFromRecord(recordConfigFiles map[string]service.ConfigFile) map[string]service.ConfigFile {
	for cfgset, _ := range recordConfigFiles {
		_, err := os.Stat(cfgset)
		if err != nil {
			delete(recordConfigFiles, cfgset)
		}
	}
	WriteRCFileToScanArea(recordConfigFiles)
	return recordConfigFiles
}

// GetCurrentDirAllFiles to get current dir all config file
func GetCurrentDirAllFiles(paths string) []string {
	configFiles := make([]string, 0)
	filepath.Walk(paths,
		func(bscpAddFile string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if path.Dir(bscpAddFile) == ".bscp" || info.IsDir() {
				return nil
			}
			configFiles = append(configFiles, bscpAddFile)
			return nil
		})
	return configFiles
}

// get string md5
func StringMd5(str string) string {
	bytes := []byte(str)
	md5Byte := md5.Sum(bytes)
	md5str := fmt.Sprintf("%x", md5Byte)
	return md5str
}

// IsDir to determine whether the given path is a folder
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}
